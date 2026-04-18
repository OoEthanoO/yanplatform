package store

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"sync"
	"time"

	"yanplatform/backend/internal/models"
)

// Store defines the data access interface for YanPlatform.
type Store interface {
	LoadSupplierSeed(path string) error
	GetSuppliers(resource string) ([]models.Supplier, error)
	GetAlternativeSuppliers(resource string) ([]models.Supplier, error)
	SaveRiskScore(score models.RiskScore) error
	GetRiskScores(resource string) ([]models.RiskScore, error)
	GetHighRiskZones(threshold float64) ([]models.RiskScore, error)
	SaveEvent(event models.GDELTEvent) error
	GetRecentEvents(limit int) ([]models.GDELTEvent, error)
	SaveTradeFlow(flow models.TradeFlow) error
	GetTradeFlows(resource string) ([]models.TradeFlow, error)
	SaveRerouteResult(result models.RerouteResult) error
	GetLatestRerouteResult(resource string) (*models.RerouteResult, error)
	GetRerouteResults(resource string, limit int) ([]models.RerouteResult, error)
	SaveChokepoint(cp models.Chokepoint) error
	GetChokepoints(resource string) ([]models.Chokepoint, error)
	SaveResource(res models.Resource) error
	GetResources() ([]models.Resource, error)
	GetResource(id string) (*models.Resource, error)
	SaveCluster(cluster models.ResourceCluster) error
	GetClusters() ([]models.ResourceCluster, error)
	GetCluster(id string) (*models.ResourceCluster, error)
	SaveRiskHistory(snapshot models.RiskScoreSnapshot) error
	GetRiskHistory(resource string, days int) ([]models.RiskScoreSnapshot, error)
	SaveAlert(alert models.AlertRecord) error
	GetRecentAlerts(limit int) ([]models.AlertRecord, error)
	AcknowledgeAlert(id string) error
	SeedInitialData() error
}

// MemoryStore provides an in-memory implementation of Store.
type MemoryStore struct {
	mu             sync.RWMutex
	suppliers      []models.Supplier
	riskScores     []models.RiskScore
	events         []models.GDELTEvent
	tradeFlows     []models.TradeFlow
	rerouteResults []models.RerouteResult
	chokepoints    []models.Chokepoint
	resources      []models.Resource
	clusters       []models.ResourceCluster
	riskHistory    []models.RiskScoreSnapshot
	alerts         []models.AlertRecord
}

// NewMemoryStore creates a new in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

// LoadSupplierSeed loads supplier data from a JSON seed file.
func (s *MemoryStore) LoadSupplierSeed(path string) error {
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
func (s *MemoryStore) GetSuppliers(resource string) ([]models.Supplier, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if resource == "" {
		result := make([]models.Supplier, len(s.suppliers))
		copy(result, s.suppliers)
		return result, nil
	}

	var filtered []models.Supplier
	for _, sup := range s.suppliers {
		if sup.Resource == resource {
			filtered = append(filtered, sup)
		}
	}
	return filtered, nil
}

// GetAlternativeSuppliers returns suppliers that are alternative reroute candidates.
func (s *MemoryStore) GetAlternativeSuppliers(resource string) ([]models.Supplier, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var alternatives []models.Supplier
	for _, sup := range s.suppliers {
		if sup.IsAlternative && (resource == "" || sup.Resource == resource) {
			alternatives = append(alternatives, sup)
		}
	}
	return alternatives, nil
}

// --- Risk Scores ---

// SaveRiskScore upserts a risk score.
func (s *MemoryStore) SaveRiskScore(score models.RiskScore) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.riskScores {
		if existing.ID == score.ID {
			s.riskScores[i] = score
			return nil
		}
	}
	s.riskScores = append(s.riskScores, score)
	return nil
}

// GetRiskScores returns all risk scores, optionally filtered by resource.
func (s *MemoryStore) GetRiskScores(resource string) ([]models.RiskScore, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if resource == "" {
		result := make([]models.RiskScore, len(s.riskScores))
		copy(result, s.riskScores)
		return result, nil
	}

	var filtered []models.RiskScore
	for _, rs := range s.riskScores {
		if rs.Resource == resource {
			filtered = append(filtered, rs)
		}
	}
	return filtered, nil
}

// GetHighRiskZones returns risk scores above the threshold.
func (s *MemoryStore) GetHighRiskZones(threshold float64) ([]models.RiskScore, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var zones []models.RiskScore
	for _, rs := range s.riskScores {
		if rs.OverallScore >= threshold {
			zones = append(zones, rs)
		}
	}
	return zones, nil
}

// --- Events ---

// SaveEvent adds a GDELT event.
func (s *MemoryStore) SaveEvent(event models.GDELTEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Avoid duplicates
	for _, existing := range s.events {
		if existing.ID == event.ID {
			return nil
		}
	}
	s.events = append(s.events, event)
	return nil
}

// GetRecentEvents returns the most recent N events.
func (s *MemoryStore) GetRecentEvents(limit int) ([]models.GDELTEvent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sorted := make([]models.GDELTEvent, len(s.events))
	copy(sorted, s.events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].EventDate.After(sorted[j].EventDate)
	})

	if limit > 0 && limit < len(sorted) {
		return sorted[:limit], nil
	}
	return sorted, nil
}

// --- Trade Flows ---

// SaveTradeFlow adds a trade flow record.
func (s *MemoryStore) SaveTradeFlow(flow models.TradeFlow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.tradeFlows {
		if existing.ID == flow.ID {
			return nil
		}
	}
	s.tradeFlows = append(s.tradeFlows, flow)
	return nil
}

// GetTradeFlows returns trade flows filtered by resource.
func (s *MemoryStore) GetTradeFlows(resource string) ([]models.TradeFlow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if resource == "" {
		result := make([]models.TradeFlow, len(s.tradeFlows))
		copy(result, s.tradeFlows)
		return result, nil
	}

	var filtered []models.TradeFlow
	for _, tf := range s.tradeFlows {
		if tf.Resource == resource {
			filtered = append(filtered, tf)
		}
	}
	return filtered, nil
}

// --- Reroute Results ---

// SaveRerouteResult stores a reroute simulation result.
func (s *MemoryStore) SaveRerouteResult(result models.RerouteResult) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rerouteResults = append(s.rerouteResults, result)
	return nil
}

// GetLatestRerouteResult returns the most recent reroute for a resource.
func (s *MemoryStore) GetLatestRerouteResult(resource string) (*models.RerouteResult, error) {
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
	return latest, nil
}

// GetRerouteResults returns recent reroute results, optionally filtered by resource.
func (s *MemoryStore) GetRerouteResults(resource string, limit int) ([]models.RerouteResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []models.RerouteResult
	for _, r := range s.rerouteResults {
		if resource == "" || r.Resource == resource {
			filtered = append(filtered, r)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].SimulatedAt.After(filtered[j].SimulatedAt)
	})

	if limit > 0 && limit < len(filtered) {
		filtered = filtered[:limit]
	}
	return filtered, nil
}

// --- Chokepoints ---

// SaveChokepoint upserts a chokepoint.
func (s *MemoryStore) SaveChokepoint(cp models.Chokepoint) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.chokepoints {
		if existing.ID == cp.ID {
			s.chokepoints[i] = cp
			return nil
		}
	}
	s.chokepoints = append(s.chokepoints, cp)
	return nil
}

// GetChokepoints returns all chokepoints, optionally filtered by resource.
func (s *MemoryStore) GetChokepoints(resource string) ([]models.Chokepoint, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if resource == "" {
		result := make([]models.Chokepoint, len(s.chokepoints))
		copy(result, s.chokepoints)
		return result, nil
	}

	var filtered []models.Chokepoint
	for _, cp := range s.chokepoints {
		if cp.Resource == resource {
			filtered = append(filtered, cp)
		}
	}
	return filtered, nil
}

// --- Resources ---

// SaveResource upserts a resource.
func (s *MemoryStore) SaveResource(res models.Resource) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.resources {
		if existing.ID == res.ID {
			s.resources[i] = res
			return nil
		}
	}
	s.resources = append(s.resources, res)
	return nil
}

// GetResources returns all tracked resources.
func (s *MemoryStore) GetResources() ([]models.Resource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.Resource, len(s.resources))
	copy(result, s.resources)
	return result, nil
}

// GetResource returns a single resource by ID.
func (s *MemoryStore) GetResource(id string) (*models.Resource, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, res := range s.resources {
		if res.ID == id {
			return &res, nil
		}
	}
	return nil, nil
}

// --- Clusters ---

// SaveCluster upserts a resource cluster.
func (s *MemoryStore) SaveCluster(cluster models.ResourceCluster) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.clusters {
		if existing.ID == cluster.ID {
			s.clusters[i] = cluster
			return nil
		}
	}
	s.clusters = append(s.clusters, cluster)
	return nil
}

// GetClusters returns all resource clusters.
func (s *MemoryStore) GetClusters() ([]models.ResourceCluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]models.ResourceCluster, len(s.clusters))
	copy(result, s.clusters)
	return result, nil
}

// GetCluster returns a single cluster by ID.
func (s *MemoryStore) GetCluster(id string) (*models.ResourceCluster, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, c := range s.clusters {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}

// --- Risk History ---

// SaveRiskHistory stores a daily risk score snapshot.
func (s *MemoryStore) SaveRiskHistory(snapshot models.RiskScoreSnapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deduplicate by ID
	for i, existing := range s.riskHistory {
		if existing.ID == snapshot.ID {
			s.riskHistory[i] = snapshot
			return nil
		}
	}
	s.riskHistory = append(s.riskHistory, snapshot)
	return nil
}

// GetRiskHistory returns risk score snapshots for a resource over N days.
func (s *MemoryStore) GetRiskHistory(resource string, days int) ([]models.RiskScoreSnapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	cutoff := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	var filtered []models.RiskScoreSnapshot
	for _, snap := range s.riskHistory {
		if (resource == "" || snap.Resource == resource) && snap.Date >= cutoff {
			filtered = append(filtered, snap)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Date < filtered[j].Date
	})

	return filtered, nil
}

// --- Alerts ---

// SaveAlert stores an alert record.
func (s *MemoryStore) SaveAlert(alert models.AlertRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, existing := range s.alerts {
		if existing.ID == alert.ID {
			s.alerts[i] = alert
			return nil
		}
	}
	s.alerts = append(s.alerts, alert)
	return nil
}

// GetRecentAlerts returns the most recent alerts.
func (s *MemoryStore) GetRecentAlerts(limit int) ([]models.AlertRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	sorted := make([]models.AlertRecord, len(s.alerts))
	copy(sorted, s.alerts)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].CreatedAt.After(sorted[j].CreatedAt)
	})

	if limit > 0 && limit < len(sorted) {
		return sorted[:limit], nil
	}
	return sorted, nil
}

// AcknowledgeAlert marks an alert as acknowledged.
func (s *MemoryStore) AcknowledgeAlert(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, alert := range s.alerts {
		if alert.ID == id {
			s.alerts[i].Acknowledged = true
			return nil
		}
	}
	return fmt.Errorf("alert %s not found", id)
}

// --- Seed Helpers ---

// SeedInitialData populates the store with baseline resources, clusters, chokepoints, and risk scores.
func (s *MemoryStore) SeedInitialData() error {
	now := time.Now()

	// 1. Initial Resources
	resources := []models.Resource{
		{ID: "gallium", Name: "Gallium", PrimaryRegion: "China", HSCodes: []string{"811292"}},
		{ID: "germanium", Name: "Germanium", PrimaryRegion: "China", HSCodes: []string{"811110"}},
		{ID: "lithium", Name: "Lithium", PrimaryRegion: "Australia", HSCodes: []string{"283691"}},
		{ID: "cobalt", Name: "Cobalt", PrimaryRegion: "DR Congo", HSCodes: []string{"810520"}},
		{ID: "graphite", Name: "Graphite", PrimaryRegion: "China", HSCodes: []string{"250410"}},
	}
	for _, r := range resources {
		s.SaveResource(r)
	}

	// 2. Initial Clusters
	clusters := []models.ResourceCluster{
		{
			ID: "cluster-semiconductors", Name: "Semiconductor Criticals",
			Description: "Minerals essential for high-frequency chips and vacuum tubes",
			ResourceIDs: []string{"gallium", "germanium"},
		},
		{
			ID: "cluster-green-energy", Name: "Green Energy / EV Battery Belt",
			Description: "Minerals essential for electric vehicle batteries and energy storage",
			ResourceIDs: []string{"lithium", "cobalt", "graphite"},
		},
	}
	for _, c := range clusters {
		s.SaveCluster(c)
	}

	// 3. Known chokepoints
	chokepoints := []models.Chokepoint{
		// Gallium/Germanium
		{
			ID: "cp-china-yunnan-ga", Name: "Yunnan Gallium Processing Hub",
			Type: "production", Country: "China", Region: "Yunnan Province",
			GlobalSharePct: 40.0, Resource: "gallium", RiskLevel: "critical",
			Latitude: 25.0389, Longitude: 102.7183,
		},
		{
			ID: "cp-china-yunnan-ge", Name: "Yunnan Germanium Processing",
			Type: "production", Country: "China", Region: "Yunnan Province",
			GlobalSharePct: 45.0, Resource: "germanium", RiskLevel: "critical",
			Latitude: 25.0389, Longitude: 102.7183,
		},
		// Lithium
		{
			ID: "cp-australia-pilbara", Name: "Pilbara Lithium Mines",
			Type: "production", Country: "Australia", Region: "Western Australia",
			GlobalSharePct: 48.0, Resource: "lithium", RiskLevel: "low",
			Latitude: -21.6, Longitude: 119.1,
		},
		// Cobalt
		{
			ID: "cp-drc-kolwezi", Name: "Kolwezi Cobalt Mining Region",
			Type: "production", Country: "DR Congo", Region: "Lualaba Province",
			GlobalSharePct: 70.0, Resource: "cobalt", RiskLevel: "critical",
			Latitude: -10.7, Longitude: 25.5,
		},
		// Graphite
		{
			ID: "cp-china-heilongjiang-graphite", Name: "Heilongjiang Graphite Hub",
			Type: "production", Country: "China", Region: "Heilongjiang Province",
			GlobalSharePct: 65.0, Resource: "graphite", RiskLevel: "critical",
			Latitude: 47.0, Longitude: 129.0,
		},
	}

	for _, cp := range chokepoints {
		s.SaveChokepoint(cp)
	}

	// Initial risk scores
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
			ID: "risk-australia-li", Region: "Australia", Country: "Australia", Resource: "lithium",
			OverallScore: 25.0, SupplyConcentration: 50.0, GeopoliticalTension: 15.0,
			TradePolicySignal: 10.0, LogisticsRisk: 25.0,
			ComputedAt: now, IsHighRisk: false,
		},
		{
			ID: "risk-drc-co", Region: "DR Congo", Country: "DR Congo", Resource: "cobalt",
			OverallScore: 88.0, SupplyConcentration: 70.0, GeopoliticalTension: 90.0,
			TradePolicySignal: 40.0, LogisticsRisk: 85.0,
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

	// Generate 30 days of historical risk data with realistic drift
	baseScores := map[string]map[string]float64{
		"gallium": {
			"overall": 82.0, "concentration": 95.0,
			"tension": 75.0, "policy": 80.0, "logistics": 50.0,
		},
		"germanium": {
			"overall": 78.0, "concentration": 85.0,
			"tension": 75.0, "policy": 80.0, "logistics": 45.0,
		},
		"lithium": {
			"overall": 25.0, "concentration": 50.0,
			"tension": 15.0, "policy": 10.0, "logistics": 25.0,
		},
		"cobalt": {
			"overall": 88.0, "concentration": 70.0,
			"tension": 90.0, "policy": 40.0, "logistics": 85.0,
		},
		"graphite": {
			"overall": 72.0, "concentration": 65.0,
			"tension": 70.0, "policy": 75.0, "logistics": 55.0,
		},
	}

	regionMap := map[string]string{
		"gallium": "China", "germanium": "China",
		"lithium": "Australia", "cobalt": "DR Congo",
		"graphite": "China",
	}

	rng := rand.New(rand.NewSource(42)) // deterministic for consistent demos

	for resource, base := range baseScores {
		region := regionMap[resource]
		for day := 30; day >= 0; day-- {
			date := now.AddDate(0, 0, -day)
			dateStr := date.Format("2006-01-02")

			// Create a drift pattern: scores gradually increase over time with some noise
			drift := float64(30-day) * 0.3 // slight upward trend
			noise := (rng.Float64() - 0.5) * 8  // +/- 4 points of noise

			clampVal := func(v, mn, mx float64) float64 {
				return math.Max(mn, math.Min(mx, v))
			}

			snap := models.RiskScoreSnapshot{
				ID:                  fmt.Sprintf("hist-%s-%s-%s", resource, region, dateStr),
				Date:                dateStr,
				Region:              region,
				Country:             region,
				Resource:            resource,
				OverallScore:        clampVal(base["overall"]+drift+noise, 0, 100),
				SupplyConcentration: clampVal(base["concentration"]+(rng.Float64()-0.5)*4, 0, 100),
				GeopoliticalTension: clampVal(base["tension"]+drift*0.8+noise*0.5, 0, 100),
				TradePolicySignal:   clampVal(base["policy"]+(rng.Float64()-0.5)*6, 0, 100),
				LogisticsRisk:       clampVal(base["logistics"]+(rng.Float64()-0.5)*3, 0, 100),
				RecordedAt:          date,
			}
			s.SaveRiskHistory(snap)
		}
	}

	// Seed sample alerts (autonomous triggers from the past)
	sampleAlerts := []models.AlertRecord{
		{
			ID: "alert-001", Resource: "cobalt", Region: "DR Congo",
			RiskScore: 88.0, Threshold: 70.0, AlternativesCount: 3,
			RerouteResultID: "reroute-cobalt-sample",
			Message:  "CRITICAL: Cobalt risk in DR Congo has reached 88.0. Autonomous reroute simulation complete. 3 alternative suppliers identified.",
			Severity: "critical", CreatedAt: now.Add(-2 * time.Hour), Acknowledged: false,
		},
		{
			ID: "alert-002", Resource: "gallium", Region: "China",
			RiskScore: 82.0, Threshold: 70.0, AlternativesCount: 3,
			RerouteResultID: "reroute-gallium-sample",
			Message:  "CRITICAL: Gallium risk in China has reached 82.0. Autonomous reroute simulation complete. 3 alternative suppliers identified.",
			Severity: "critical", CreatedAt: now.Add(-6 * time.Hour), Acknowledged: false,
		},
		{
			ID: "alert-003", Resource: "germanium", Region: "China",
			RiskScore: 78.0, Threshold: 70.0, AlternativesCount: 3,
			RerouteResultID: "reroute-germanium-sample",
			Message:  "WARNING: Germanium risk in China has reached 78.0. Autonomous reroute simulation complete. 3 alternative suppliers identified.",
			Severity: "warning", CreatedAt: now.Add(-24 * time.Hour), Acknowledged: true,
		},
		{
			ID: "alert-004", Resource: "graphite", Region: "China",
			RiskScore: 72.0, Threshold: 70.0, AlternativesCount: 2,
			RerouteResultID: "reroute-graphite-sample",
			Message:  "WARNING: Graphite risk in China has reached 72.0. Autonomous reroute simulation complete. 2 alternative suppliers identified.",
			Severity: "warning", CreatedAt: now.Add(-48 * time.Hour), Acknowledged: true,
		},
	}
	for _, a := range sampleAlerts {
		s.SaveAlert(a)
	}

	return nil
}
