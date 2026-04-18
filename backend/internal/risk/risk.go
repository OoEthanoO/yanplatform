// Package risk implements the risk scoring and shadow reroute engines.
package risk

import (
	"fmt"
	"sort"
	"time"

	"yanplatform/backend/internal/config"
	"yanplatform/backend/internal/models"
	"yanplatform/backend/internal/store"
)

// Engine computes risk scores and runs reroute simulations.
type Engine struct {
	Store  store.Store
	Config *config.RiskConfig
}

// NewEngine creates a new risk scoring engine.
func NewEngine(s store.Store, cfg *config.RiskConfig) *Engine {
	return &Engine{Store: s, Config: cfg}
}

// ComputeRiskScore calculates the overall risk score for a resource in a specific region.
// Uses weighted factors: supply concentration (40%), geopolitical tension (30%),
// trade policy signals (20%), logistics risk (10%).
func (e *Engine) ComputeRiskScore(region, resource string) models.RiskScore {
	suppliers, _ := e.Store.GetSuppliers(resource)
	events, _ := e.Store.GetRecentEvents(50)

	// 1. Supply Concentration Score (0-100)
	concentrationScore := e.computeConcentration(suppliers, region)

	// 2. Geopolitical Tension Score (0-100)
	tensionScore := e.computeGeopoliticalTension(events, region)

	// 3. Trade Policy Signal Score (0-100)
	policyScore := e.computeTradePolicySignal(events, region)

	// 4. Logistics Risk Score (0-100)
	logisticsScore := e.computeLogisticsRisk(region)

	// Weighted overall score
	overall := (concentrationScore * e.Config.WeightSupplyConcentration) +
		(tensionScore * e.Config.WeightGeopoliticalTension) +
		(policyScore * e.Config.WeightTradePolicySignal) +
		(logisticsScore * e.Config.WeightLogisticsRisk)

	return models.RiskScore{
		ID:                  fmt.Sprintf("risk-%s-%s", region, resource),
		Region:              region,
		Country:             region,
		Resource:            resource,
		OverallScore:        clamp(overall, 0, 100),
		SupplyConcentration: concentrationScore,
		GeopoliticalTension: tensionScore,
		TradePolicySignal:   policyScore,
		LogisticsRisk:       logisticsScore,
		ComputedAt:          time.Now(),
		IsHighRisk:          overall >= e.Config.HighRiskThreshold,
	}
}

// computeConcentration calculates how concentrated supply is in a given region.
func (e *Engine) computeConcentration(suppliers []models.Supplier, region string) float64 {
	var totalCapacity, regionCapacity float64

	for _, s := range suppliers {
		totalCapacity += s.CapacityTonnesYr
		if s.Country == region {
			regionCapacity += s.CapacityTonnesYr
		}
	}

	if totalCapacity == 0 {
		return 0
	}

	// Score is the percentage of global supply in this region
	return (regionCapacity / totalCapacity) * 100
}

// computeGeopoliticalTension derives a tension score from recent GDELT events.
func (e *Engine) computeGeopoliticalTension(events []models.GDELTEvent, region string) float64 {
	var totalTone float64
	var count int

	for _, evt := range events {
		if evt.Actor1Country == region || evt.Actor2Country == region {
			totalTone += evt.GoldsteinScale
			count++
		}
	}

	if count == 0 {
		return 50 // Default moderate tension when no data
	}

	avgGoldstein := totalTone / float64(count)
	score := ((avgGoldstein * -1) + 10) * 5
	return clamp(score, 0, 100)
}

// computeTradePolicySignal scores trade policy activity (export controls, tariffs, quotas).
func (e *Engine) computeTradePolicySignal(events []models.GDELTEvent, region string) float64 {
	var escalationCount, deescalationCount int

	for _, evt := range events {
		if evt.Actor1Country == region || evt.Actor2Country == region {
			switch evt.SentimentLabel {
			case "escalation":
				escalationCount++
			case "de-escalation":
				deescalationCount++
			}
		}
	}

	total := escalationCount + deescalationCount
	if total == 0 {
		return 30 // Default low-moderate when no data
	}

	ratio := float64(escalationCount) / float64(total)
	return clamp(ratio*100, 0, 100)
}

// computeLogisticsRisk scores shipping/transport vulnerability.
func (e *Engine) computeLogisticsRisk(region string) float64 {
	risks := map[string]float64{
		"China":         55.0,
		"Japan":         45.0,
		"South Korea":   40.0,
		"Germany":       30.0,
		"Belgium":       25.0,
		"Canada":        20.0,
		"United Kingdom": 30.0,
	}

	if score, ok := risks[region]; ok {
		return score
	}
	return 35 // Default moderate risk
}

// SimulateReroute runs a shadow reroute simulation for a given resource.
func (e *Engine) SimulateReroute(resource string) *models.RerouteResult {
	riskScores, _ := e.Store.GetRiskScores(resource)
	var triggerScore *models.RiskScore
	for i, rs := range riskScores {
		if rs.OverallScore >= e.Config.RerouteTriggerThreshold {
			triggerScore = &riskScores[i]
			break
		}
	}

	if triggerScore == nil {
		return nil // No disruption scenario triggered
	}

	suppliers, _ := e.Store.GetSuppliers(resource)
	var disruptedCapacity float64
	for _, s := range suppliers {
		if s.Country == triggerScore.Country {
			disruptedCapacity += s.CapacityTonnesYr
		}
	}

	alternatives, _ := e.Store.GetSuppliers(resource)

	var ranked []models.RerouteAlternative
	for _, alt := range alternatives {
		absorptionPct := 0.0
		if disruptedCapacity > 0 {
			absorptionPct = (alt.CapacityTonnesYr / disruptedCapacity) * 100
		}

		feasibility := (alt.NeutralityScore * 40) +
			(clamp(absorptionPct, 0, 100) * 0.3) +
			((100 - e.computeLogisticsRisk(alt.Country)) * 0.3)

		ranked = append(ranked, models.RerouteAlternative{
			SupplierID:       alt.ID,
			SupplierName:     alt.Name,
			Country:          alt.Country,
			CapacityTonnes:   alt.CapacityTonnesYr,
			AbsorptionPct:    clamp(absorptionPct, 0, 100),
			FeasibilityScore: clamp(feasibility, 0, 100),
			LeadTimeDays:     estimateLeadTime(alt.Country),
			Latitude:         alt.Latitude,
			Longitude:        alt.Longitude,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].FeasibilityScore > ranked[j].FeasibilityScore
	})

	if len(ranked) > 3 {
		ranked = ranked[:3]
	}

	result := &models.RerouteResult{
		ID:               fmt.Sprintf("reroute-%s-%d", resource, time.Now().Unix()),
		TriggerRegion:    triggerScore.Region,
		TriggerRiskScore: triggerScore.OverallScore,
		Resource:         resource,
		Alternatives:     ranked,
		SimulatedAt:      time.Now(),
	}

	_ = e.Store.SaveRerouteResult(*result)
	return result
}

// RecalculateAll recalculates risk scores for all monitored resources.
func (e *Engine) RecalculateAll() {
	resources, err := e.Store.GetResources()
	if err != nil {
		fmt.Printf("[Risk Engine] Error fetching resources: %v\n", err)
		return
	}

	today := time.Now().Format("2006-01-02")

	for _, r := range resources {
		if r.PrimaryRegion != "" {
			score := e.ComputeRiskScore(r.PrimaryRegion, r.ID)
			_ = e.Store.SaveRiskScore(score)

			// Save daily snapshot for time-series history
			snapshot := models.RiskScoreSnapshot{
				ID:                  fmt.Sprintf("hist-%s-%s-%s", r.ID, r.PrimaryRegion, today),
				Date:                today,
				Region:              score.Region,
				Country:             score.Country,
				Resource:            score.Resource,
				OverallScore:        score.OverallScore,
				SupplyConcentration: score.SupplyConcentration,
				GeopoliticalTension: score.GeopoliticalTension,
				TradePolicySignal:   score.TradePolicySignal,
				LogisticsRisk:       score.LogisticsRisk,
				RecordedAt:          time.Now(),
			}
			_ = e.Store.SaveRiskHistory(snapshot)
		}
	}
}

// estimateLeadTime returns an estimated lead time in days for shipping from a country.
func estimateLeadTime(country string) int {
	times := map[string]int{
		"Canada":         14,
		"Japan":          21,
		"South Korea":    25,
		"Germany":        18,
		"Belgium":        16,
		"United Kingdom": 15,
	}
	if t, ok := times[country]; ok {
		return t
	}
	return 30
}

// clamp restricts a value to a range.
func clamp(value, min, max float64) float64 {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
