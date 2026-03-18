// Package store provides the Firestore data access layer.
// For MVP development without Firestore credentials, it uses an in-memory store
// seeded from JSON files.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
	"time"

	"yanplatform/backend/internal/models"
)

// Store provides data access for all collections.
type Store struct {
	mu             sync.RWMutex
	suppliers      []models.Supplier
	riskScores     []models.RiskScore
	events         []models.GDELTEvent
	tradeFlows     []models.TradeFlow
	rerouteResults []models.RerouteResult
	chokepoints    []models.Chokepoint
}

// New creates a new in-memory store.
func New() *Store {
	return &Store{}
}

// LoadSupplierSeed loads supplier data from a JSON seed file.
func (s *Store) LoadSupplierSeed(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading supplier seed: %w", err)
	}

	var suppliers []models.Supplier
	if err := json.Unmarshal(data, &suppliers); err != nil {
		return fmt.Errorf("parsing supplier seed: %w", err)
	}

	s.mu.Lock()
	s.suppliers = suppliers
	s.mu.Unlock()

	return nil
}

// --- Suppliers ---

// GetSuppliers returns all suppliers, optionally filtered by resource.
func (s *Store) GetSuppliers(resource string) []models.Supplier {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if resource == "" {
		result := make([]models.Supplier, len(s.suppliers))
		copy(result, s.suppliers)
		return result
	}

	var filtered []models.Supplier
	for _, sup := range s.suppliers {
		if sup.Resource == resource {
			filtered = append(filtered, sup)
		}
	}
	return filtered
}

// GetAlternativeSuppliers returns suppliers that are alternative reroute candidates.
func (s *Store) GetAlternativeSuppliers(resource string) []models.Supplier {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var alternatives []models.Supplier
	for _, sup := range s.suppliers {
		if sup.IsAlternative && (resource == "" || sup.Resource == resource) {
			alternatives = append(alternatives, sup)
		}
	}
	return alternatives
}

// --- Risk Scores ---

// SaveRiskScore upserts a risk score.
func (s *Store) SaveRiskScore(score models.RiskScore) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.riskScores {
		if existing.ID == score.ID {
			s.riskScores[i] = score
			return
		}
	}
	s.riskScores = append(s.riskScores, score)
}

// GetRiskScores returns all risk scores, optionally filtered by resource.
func (s *Store) GetRiskScores(resource string) []models.RiskScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if resource == "" {
		result := make([]models.RiskScore, len(s.riskScores))
		copy(result, s.riskScores)
		return result
	}

	var filtered []models.RiskScore
	for _, rs := range s.riskScores {
		if rs.Resource == resource {
			filtered = append(filtered, rs)
		}
	}
	return filtered
}

// GetHighRiskZones returns risk scores above the threshold.
func (s *Store) GetHighRiskZones(threshold float64) []models.RiskScore {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var zones []models.RiskScore
	for _, rs := range s.riskScores {
		if rs.OverallScore >= threshold {
			zones = append(zones, rs)
		}
	}
	return zones
}

// --- Events ---

// SaveEvent adds a GDELT event.
func (s *Store) SaveEvent(event models.GDELTEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Avoid duplicates
	for _, existing := range s.events {
		if existing.ID == event.ID {
			return
		}
	}
	s.events = append(s.events, event)
}

// GetRecentEvents returns the most recent N events.
func (s *Store) GetRecentEvents(limit int) []models.GDELTEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sorted := make([]models.GDELTEvent, len(s.events))
	copy(sorted, s.events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].EventDate.After(sorted[j].EventDate)
	})

	if limit > 0 && limit < len(sorted) {
		return sorted[:limit]
	}
	return sorted
}

// --- Trade Flows ---

// SaveTradeFlow adds a trade flow record.
func (s *Store) SaveTradeFlow(flow models.TradeFlow) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.tradeFlows {
		if existing.ID == flow.ID {
			return
		}
	}
	s.tradeFlows = append(s.tradeFlows, flow)
}

// GetTradeFlows returns trade flows filtered by resource.
func (s *Store) GetTradeFlows(resource string) []models.TradeFlow {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if resource == "" {
		result := make([]models.TradeFlow, len(s.tradeFlows))
		copy(result, s.tradeFlows)
		return result
	}

	var filtered []models.TradeFlow
	for _, tf := range s.tradeFlows {
		if tf.Resource == resource {
			filtered = append(filtered, tf)
		}
	}
	return filtered
}

// --- Reroute Results ---

// SaveRerouteResult stores a reroute simulation result.
func (s *Store) SaveRerouteResult(result models.RerouteResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rerouteResults = append(s.rerouteResults, result)
}

// GetLatestRerouteResult returns the most recent reroute for a resource.
func (s *Store) GetLatestRerouteResult(resource string) *models.RerouteResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var latest *models.RerouteResult
	for i := range s.rerouteResults {
		r := &s.rerouteResults[i]
		if r.Resource == resource {
			if latest == nil || r.SimulatedAt.After(latest.SimulatedAt) {
				latest = r
			}
		}
	}
	return latest
}

// --- Chokepoints ---

// SaveChokepoint upserts a chokepoint.
func (s *Store) SaveChokepoint(cp models.Chokepoint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.chokepoints {
		if existing.ID == cp.ID {
			s.chokepoints[i] = cp
			return
		}
	}
	s.chokepoints = append(s.chokepoints, cp)
}

// GetChokepoints returns all chokepoints, optionally filtered by resource.
func (s *Store) GetChokepoints(resource string) []models.Chokepoint {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if resource == "" {
		result := make([]models.Chokepoint, len(s.chokepoints))
		copy(result, s.chokepoints)
		return result
	}

	var filtered []models.Chokepoint
	for _, cp := range s.chokepoints {
		if cp.Resource == resource {
			filtered = append(filtered, cp)
		}
	}
	return filtered
}

// --- Seed Helpers ---

// SeedInitialData populates the store with baseline chokepoints and computed risk scores.
func (s *Store) SeedInitialData() {
	now := time.Now()

	// Known chokepoints for gallium and germanium
	chokepoints := []models.Chokepoint{
		{
			ID: "cp-china-yunnan-ga", Name: "Yunnan Gallium Processing Hub",
			Type: "production", Country: "China", Region: "Yunnan Province",
			GlobalSharePct: 40.0, Resource: "gallium", RiskLevel: "critical",
			Latitude: 25.0389, Longitude: 102.7183,
		},
		{
			ID: "cp-china-inner-mongolia-ga", Name: "Inner Mongolia Gallium Refineries",
			Type: "production", Country: "China", Region: "Inner Mongolia",
			GlobalSharePct: 35.0, Resource: "gallium", RiskLevel: "critical",
			Latitude: 40.8174, Longitude: 111.7656,
		},
		{
			ID: "cp-china-yunnan-ge", Name: "Yunnan Germanium Processing",
			Type: "production", Country: "China", Region: "Yunnan Province",
			GlobalSharePct: 45.0, Resource: "germanium", RiskLevel: "critical",
			Latitude: 25.0389, Longitude: 102.7183,
		},
		{
			ID: "cp-china-inner-mongolia-ge", Name: "Inner Mongolia Germanium Refinery",
			Type: "production", Country: "China", Region: "Inner Mongolia",
			GlobalSharePct: 25.0, Resource: "germanium", RiskLevel: "critical",
			Latitude: 40.8174, Longitude: 111.7656,
		},
		{
			ID: "cp-malacca-strait", Name: "Strait of Malacca",
			Type: "shipping", Country: "International", Region: "Southeast Asia",
			GlobalSharePct: 60.0, Resource: "gallium", RiskLevel: "elevated",
			Latitude: 2.5, Longitude: 101.8,
		},
		{
			ID: "cp-suez-canal", Name: "Suez Canal",
			Type: "shipping", Country: "Egypt", Region: "Suez",
			GlobalSharePct: 30.0, Resource: "germanium", RiskLevel: "elevated",
			Latitude: 30.4574, Longitude: 32.3499,
		},
	}

	for _, cp := range chokepoints {
		s.SaveChokepoint(cp)
	}

	// Initial risk scores based on known supply concentration
	riskScores := []models.RiskScore{
		{
			ID: "risk-china-ga", Region: "China", Country: "China", Resource: "gallium",
			OverallScore: 82.0, SupplyConcentration: 95.0, GeopoliticalTension: 75.0,
			TradePolicySignal: 80.0, LogisticsRisk: 50.0,
			ComputedAt: now, IsHighRisk: true,
		},
		{
			ID: "risk-china-ge", Region: "China", Country: "China", Resource: "germanium",
			OverallScore: 78.0, SupplyConcentration: 85.0, GeopoliticalTension: 75.0,
			TradePolicySignal: 80.0, LogisticsRisk: 45.0,
			ComputedAt: now, IsHighRisk: true,
		},
		{
			ID: "risk-canada-ge", Region: "Canada", Country: "Canada", Resource: "germanium",
			OverallScore: 18.0, SupplyConcentration: 15.0, GeopoliticalTension: 10.0,
			TradePolicySignal: 5.0, LogisticsRisk: 30.0,
			ComputedAt: now, IsHighRisk: false,
		},
		{
			ID: "risk-japan-ga", Region: "Japan", Country: "Japan", Resource: "gallium",
			OverallScore: 22.0, SupplyConcentration: 5.0, GeopoliticalTension: 15.0,
			TradePolicySignal: 10.0, LogisticsRisk: 40.0,
			ComputedAt: now, IsHighRisk: false,
		},
		{
			ID: "risk-germany-ga", Region: "Germany", Country: "Germany", Resource: "gallium",
			OverallScore: 15.0, SupplyConcentration: 8.0, GeopoliticalTension: 8.0,
			TradePolicySignal: 5.0, LogisticsRisk: 25.0,
			ComputedAt: now, IsHighRisk: false,
		},
	}

	for _, rs := range riskScores {
		s.SaveRiskScore(rs)
	}

	// Seed some example trade flows
	tradeFlows := []models.TradeFlow{
		{ID: "tf-cn-us-ga-2025", Year: 2025, Month: 6, ReporterCountry: "China", PartnerCountry: "United States", HSCode: "811292", Resource: "gallium", FlowType: "export", ValueUSD: 45000000, WeightKg: 120000, IngestedAt: now},
		{ID: "tf-cn-jp-ga-2025", Year: 2025, Month: 6, ReporterCountry: "China", PartnerCountry: "Japan", HSCode: "811292", Resource: "gallium", FlowType: "export", ValueUSD: 38000000, WeightKg: 95000, IngestedAt: now},
		{ID: "tf-cn-de-ge-2025", Year: 2025, Month: 6, ReporterCountry: "China", PartnerCountry: "Germany", HSCode: "811110", Resource: "germanium", FlowType: "export", ValueUSD: 52000000, WeightKg: 35000, IngestedAt: now},
		{ID: "tf-ca-us-ge-2025", Year: 2025, Month: 6, ReporterCountry: "Canada", PartnerCountry: "United States", HSCode: "811110", Resource: "germanium", FlowType: "export", ValueUSD: 12000000, WeightKg: 8000, IngestedAt: now},
		{ID: "tf-be-us-ge-2025", Year: 2025, Month: 6, ReporterCountry: "Belgium", PartnerCountry: "United States", HSCode: "811110", Resource: "germanium", FlowType: "export", ValueUSD: 15000000, WeightKg: 10000, IngestedAt: now},
	}

	for _, tf := range tradeFlows {
		s.SaveTradeFlow(tf)
	}

	// Seed example GDELT events
	events := []models.GDELTEvent{
		{
			ID: "evt-001", EventDate: now.Add(-24 * time.Hour),
			Actor1Name: "China Ministry of Commerce", Actor1Country: "China",
			Actor2Name: "Global Semiconductor Industry", Actor2Country: "",
			EventType: "RESTRICT", Description: "China announces tightened export controls on gallium and germanium products effective immediately",
			AvgTone: -3.5, GoldsteinScale: -5.0, SourceURL: "https://example.com/china-export-controls",
			Relevance: 0.95, SentimentLabel: "escalation", IngestedAt: now,
		},
		{
			ID: "evt-002", EventDate: now.Add(-48 * time.Hour),
			Actor1Name: "European Commission", Actor1Country: "EU",
			Actor2Name: "Critical Raw Materials Alliance", Actor2Country: "EU",
			EventType: "COOPERATE", Description: "EU announces €2 billion investment in domestic critical mineral processing capacity",
			AvgTone: 4.2, GoldsteinScale: 7.0, SourceURL: "https://example.com/eu-critical-minerals",
			Relevance: 0.82, SentimentLabel: "de-escalation", IngestedAt: now,
		},
		{
			ID: "evt-003", EventDate: now.Add(-72 * time.Hour),
			Actor1Name: "US Commerce Department", Actor1Country: "United States",
			Actor2Name: "China", Actor2Country: "China",
			EventType: "DEMAND", Description: "US urges allies to diversify gallium supply chains away from Chinese dependence",
			AvgTone: -1.8, GoldsteinScale: -3.0, SourceURL: "https://example.com/us-gallium-diversify",
			Relevance: 0.88, SentimentLabel: "escalation", IngestedAt: now,
		},
		{
			ID: "evt-004", EventDate: now.Add(-96 * time.Hour),
			Actor1Name: "Japan Ministry of Economy", Actor1Country: "Japan",
			Actor2Name: "Dowa Holdings", Actor2Country: "Japan",
			EventType: "COOPERATE", Description: "Japan subsidizes gallium recycling facility expansion to reduce import dependence",
			AvgTone: 3.8, GoldsteinScale: 6.0, SourceURL: "https://example.com/japan-gallium-recycling",
			Relevance: 0.75, SentimentLabel: "de-escalation", IngestedAt: now,
		},
		{
			ID: "evt-005", EventDate: now.Add(-12 * time.Hour),
			Actor1Name: "China State Council", Actor1Country: "China",
			Actor2Name: "Rare Earth Industry", Actor2Country: "China",
			EventType: "RESTRICT", Description: "China signals potential export quota reductions for strategic minerals in Q2 2026",
			AvgTone: -4.1, GoldsteinScale: -6.5, SourceURL: "https://example.com/china-quota-signal",
			Relevance: 0.92, SentimentLabel: "escalation", IngestedAt: now,
		},
	}

	for _, evt := range events {
		s.SaveEvent(evt)
	}
}
