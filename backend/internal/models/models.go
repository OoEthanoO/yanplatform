// Package models defines the core data structures for the YanPlatform system.
package models

import "time"

// Resource represents a tracked critical mineral/material.
type Resource struct {
	ID   string `json:"id" firestore:"id"`
	Name string `json:"name" firestore:"name"`
	// HS commodity codes for trade data lookup
	HSCodes []string `json:"hs_codes" firestore:"hs_codes"`
	// PrimaryRegion is the main region we monitor for risk (e.g. "China")
	PrimaryRegion string `json:"primary_region" firestore:"primary_region"`
}

// ResourceCluster groups resources together (e.g. "Semiconductors", "EV Battery Belt").
type ResourceCluster struct {
	ID          string   `json:"id" firestore:"id"`
	Name        string   `json:"name" firestore:"name"`
	Description string   `json:"description" firestore:"description"`
	ResourceIDs []string `json:"resource_ids" firestore:"resource_ids"`
}

// Supplier represents a known producer/refiner of a tracked resource.
type Supplier struct {
	ID               string  `json:"id" firestore:"id"`
	Name             string  `json:"name" firestore:"name"`
	Country          string  `json:"country" firestore:"country"`
	Region           string  `json:"region" firestore:"region"`
	Resource         string  `json:"resource" firestore:"resource"`         // "gallium" or "germanium"
	CapacityTonnesYr float64 `json:"capacity_tonnes_yr" firestore:"capacity_tonnes_yr"`
	NeutralityScore  float64 `json:"neutrality_score" firestore:"neutrality_score"` // 0.0 (hostile) to 1.0 (neutral/allied)
	Latitude         float64 `json:"latitude" firestore:"latitude"`
	Longitude        float64 `json:"longitude" firestore:"longitude"`
	IsAlternative    bool    `json:"is_alternative" firestore:"is_alternative"` // true if this is a reroute candidate
}

// RiskScore represents a computed risk assessment for a specific region/supplier.
type RiskScore struct {
	ID                   string    `json:"id" firestore:"id"`
	Region               string    `json:"region" firestore:"region"`
	Country              string    `json:"country" firestore:"country"`
	Resource             string    `json:"resource" firestore:"resource"`
	OverallScore         float64   `json:"overall_score" firestore:"overall_score"`                   // 0-100
	SupplyConcentration  float64   `json:"supply_concentration" firestore:"supply_concentration"`     // 0-100
	GeopoliticalTension  float64   `json:"geopolitical_tension" firestore:"geopolitical_tension"`     // 0-100
	TradePolicySignal    float64   `json:"trade_policy_signal" firestore:"trade_policy_signal"`       // 0-100
	LogisticsRisk        float64   `json:"logistics_risk" firestore:"logistics_risk"`                 // 0-100
	ComputedAt           time.Time `json:"computed_at" firestore:"computed_at"`
	IsHighRisk           bool      `json:"is_high_risk" firestore:"is_high_risk"`
}

// GDELTEvent represents a processed geopolitical event from GDELT.
type GDELTEvent struct {
	ID              string    `json:"id" firestore:"id"`
	EventDate       time.Time `json:"event_date" firestore:"event_date"`
	Actor1Name      string    `json:"actor1_name" firestore:"actor1_name"`
	Actor1Country   string    `json:"actor1_country" firestore:"actor1_country"`
	Actor2Name      string    `json:"actor2_name" firestore:"actor2_name"`
	Actor2Country   string    `json:"actor2_country" firestore:"actor2_country"`
	EventType       string    `json:"event_type" firestore:"event_type"`
	Description     string    `json:"description" firestore:"description"`
	AvgTone         float64   `json:"avg_tone" firestore:"avg_tone"`
	GoldsteinScale  float64   `json:"goldstein_scale" firestore:"goldstein_scale"` // -10 to +10
	SourceURL       string    `json:"source_url" firestore:"source_url"`
	Relevance       float64   `json:"relevance" firestore:"relevance"`       // 0-1 NIM-classified relevance
	SentimentLabel  string    `json:"sentiment_label" firestore:"sentiment_label"` // "escalation", "de-escalation", "neutral"
	IngestedAt      time.Time `json:"ingested_at" firestore:"ingested_at"`
}

// TradeFlow represents import/export data from UN Comtrade.
type TradeFlow struct {
	ID             string    `json:"id" firestore:"id"`
	Year           int       `json:"year" firestore:"year"`
	Month          int       `json:"month" firestore:"month"`
	ReporterCountry string  `json:"reporter_country" firestore:"reporter_country"`
	PartnerCountry  string  `json:"partner_country" firestore:"partner_country"`
	HSCode         string    `json:"hs_code" firestore:"hs_code"`
	Resource       string    `json:"resource" firestore:"resource"`
	FlowType       string    `json:"flow_type" firestore:"flow_type"` // "import" or "export"
	ValueUSD       float64   `json:"value_usd" firestore:"value_usd"`
	WeightKg       float64   `json:"weight_kg" firestore:"weight_kg"`
	IngestedAt     time.Time `json:"ingested_at" firestore:"ingested_at"`
}

// RerouteResult represents a shadow reroute simulation result.
type RerouteResult struct {
	ID                 string              `json:"id" firestore:"id"`
	TriggerRegion      string              `json:"trigger_region" firestore:"trigger_region"`
	TriggerRiskScore   float64             `json:"trigger_risk_score" firestore:"trigger_risk_score"`
	Resource           string              `json:"resource" firestore:"resource"`
	Alternatives       []RerouteAlternative `json:"alternatives" firestore:"alternatives"`
	SimulatedAt        time.Time           `json:"simulated_at" firestore:"simulated_at"`
}

// RerouteAlternative represents one alternative supplier in a reroute simulation.
type RerouteAlternative struct {
	SupplierID      string  `json:"supplier_id" firestore:"supplier_id"`
	SupplierName    string  `json:"supplier_name" firestore:"supplier_name"`
	Country         string  `json:"country" firestore:"country"`
	CapacityTonnes  float64 `json:"capacity_tonnes" firestore:"capacity_tonnes"`
	AbsorptionPct   float64 `json:"absorption_pct" firestore:"absorption_pct"` // % of disrupted supply this can absorb
	FeasibilityScore float64 `json:"feasibility_score" firestore:"feasibility_score"` // 0-100
	LeadTimeDays    int     `json:"lead_time_days" firestore:"lead_time_days"`
	Latitude        float64 `json:"latitude" firestore:"latitude"`
	Longitude       float64 `json:"longitude" firestore:"longitude"`
}

// RiskScoreSnapshot represents a daily snapshot of a risk score for historical tracking.
type RiskScoreSnapshot struct {
	ID                   string    `json:"id" firestore:"id"`
	Date                 string    `json:"date" firestore:"date"` // YYYY-MM-DD
	Region               string    `json:"region" firestore:"region"`
	Country              string    `json:"country" firestore:"country"`
	Resource             string    `json:"resource" firestore:"resource"`
	OverallScore         float64   `json:"overall_score" firestore:"overall_score"`
	SupplyConcentration  float64   `json:"supply_concentration" firestore:"supply_concentration"`
	GeopoliticalTension  float64   `json:"geopolitical_tension" firestore:"geopolitical_tension"`
	TradePolicySignal    float64   `json:"trade_policy_signal" firestore:"trade_policy_signal"`
	LogisticsRisk        float64   `json:"logistics_risk" firestore:"logistics_risk"`
	RecordedAt           time.Time `json:"recorded_at" firestore:"recorded_at"`
}

// AlertRecord represents a system-generated alert when risk thresholds are breached.
type AlertRecord struct {
	ID                string    `json:"id" firestore:"id"`
	Resource          string    `json:"resource" firestore:"resource"`
	Region            string    `json:"region" firestore:"region"`
	RiskScore         float64   `json:"risk_score" firestore:"risk_score"`
	Threshold         float64   `json:"threshold" firestore:"threshold"`
	AlternativesCount int       `json:"alternatives_count" firestore:"alternatives_count"`
	RerouteResultID   string    `json:"reroute_result_id" firestore:"reroute_result_id"`
	Message           string    `json:"message" firestore:"message"`
	Severity          string    `json:"severity" firestore:"severity"` // "critical", "warning", "info"
	CreatedAt         time.Time `json:"created_at" firestore:"created_at"`
	Acknowledged      bool      `json:"acknowledged" firestore:"acknowledged"`
}

// RiskOverview is the API response for the dashboard overview.
type RiskOverview struct {
	ResourceRisks map[string]RiskScore `json:"resource_risks"`
	RecentEvents  int                  `json:"recent_events"`
	HighRiskZones int                  `json:"high_risk_zones"`
	LastUpdated   time.Time            `json:"last_updated"`
}

// Chokepoint represents a geographic bottleneck in the supply chain.
type Chokepoint struct {
	ID                  string  `json:"id" firestore:"id"`
	Name                string  `json:"name" firestore:"name"`
	Type                string  `json:"type" firestore:"type"` // "production", "shipping", "processing"
	Country             string  `json:"country" firestore:"country"`
	Region              string  `json:"region" firestore:"region"`
	GlobalSharePct      float64 `json:"global_share_pct" firestore:"global_share_pct"`
	Resource            string  `json:"resource" firestore:"resource"`
	RiskLevel           string  `json:"risk_level" firestore:"risk_level"` // "critical", "elevated", "low"
	Latitude            float64 `json:"latitude" firestore:"latitude"`
	Longitude           float64 `json:"longitude" firestore:"longitude"`
}
