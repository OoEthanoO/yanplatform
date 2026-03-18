package risk

import (
	"testing"

	"yanplatform/backend/internal/config"
	"yanplatform/backend/internal/store"
)

func setupTestEngine() (*Engine, *store.Store) {
	s := store.New()

	seedPath := "../../data/suppliers_seed.json"
	if err := s.LoadSupplierSeed(seedPath); err != nil {
		// Tests can still run without seed file
	}
	s.SeedInitialData()

	cfg := &config.RiskConfig{
		HighRiskThreshold:         70.0,
		RerouteTriggerThreshold:   70.0,
		WeightSupplyConcentration: 0.40,
		WeightGeopoliticalTension: 0.30,
		WeightTradePolicySignal:   0.20,
		WeightLogisticsRisk:       0.10,
	}

	return NewEngine(s, cfg), s
}

func TestComputeRiskScore_ChinaGallium(t *testing.T) {
	engine, _ := setupTestEngine()

	score := engine.ComputeRiskScore("China", "gallium")

	if score.OverallScore < 50 {
		t.Errorf("China gallium risk should be high, got %.1f", score.OverallScore)
	}
	if !score.IsHighRisk {
		t.Error("China gallium should be flagged as high risk")
	}
	if score.SupplyConcentration < 70 {
		t.Errorf("China supply concentration for gallium should be very high, got %.1f", score.SupplyConcentration)
	}
}

func TestComputeRiskScore_CanadaGermanium(t *testing.T) {
	engine, _ := setupTestEngine()

	score := engine.ComputeRiskScore("Canada", "germanium")

	if score.OverallScore > 50 {
		t.Errorf("Canada germanium risk should be moderate/low, got %.1f", score.OverallScore)
	}
	if score.IsHighRisk {
		t.Error("Canada should not be flagged as high risk")
	}
}

func TestSimulateReroute_Gallium(t *testing.T) {
	engine, _ := setupTestEngine()

	// Recalculate to ensure China is above threshold
	engine.RecalculateAll()

	result := engine.SimulateReroute("gallium")

	if result == nil {
		t.Fatal("Expected reroute result for gallium, got nil")
	}
	if result.Resource != "gallium" {
		t.Errorf("Expected resource 'gallium', got '%s'", result.Resource)
	}
	if len(result.Alternatives) == 0 {
		t.Error("Expected at least one alternative supplier")
	}
	if len(result.Alternatives) > 3 {
		t.Errorf("Expected at most 3 alternatives, got %d", len(result.Alternatives))
	}

	// Verify alternatives are not from China
	for _, alt := range result.Alternatives {
		if alt.Country == "China" {
			t.Errorf("Alternative supplier should not be from China: %s", alt.SupplierName)
		}
	}
}

func TestSimulateReroute_Germanium(t *testing.T) {
	engine, _ := setupTestEngine()
	engine.RecalculateAll()

	result := engine.SimulateReroute("germanium")

	if result == nil {
		t.Fatal("Expected reroute result for germanium, got nil")
	}
	if len(result.Alternatives) == 0 {
		t.Error("Expected at least one alternative supplier for germanium")
	}

	// Verify alternatives have positive feasibility
	for _, alt := range result.Alternatives {
		if alt.FeasibilityScore <= 0 {
			t.Errorf("Alternative %s has non-positive feasibility: %.1f", alt.SupplierName, alt.FeasibilityScore)
		}
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		value, min, max, expected float64
	}{
		{50, 0, 100, 50},
		{-5, 0, 100, 0},
		{150, 0, 100, 100},
		{0, 0, 100, 0},
		{100, 0, 100, 100},
	}

	for _, tt := range tests {
		result := clamp(tt.value, tt.min, tt.max)
		if result != tt.expected {
			t.Errorf("clamp(%v, %v, %v) = %v, want %v", tt.value, tt.min, tt.max, result, tt.expected)
		}
	}
}

func TestRecalculateAll(t *testing.T) {
	engine, s := setupTestEngine()

	engine.RecalculateAll()

	// Should have risk scores for multiple regions
	allScores := s.GetRiskScores("")
	if len(allScores) == 0 {
		t.Error("Expected risk scores after recalculation")
	}

	// China should be highest risk
	var chinaGa, canadaGe float64
	for _, rs := range allScores {
		if rs.Country == "China" && rs.Resource == "gallium" {
			chinaGa = rs.OverallScore
		}
		if rs.Country == "Canada" && rs.Resource == "germanium" {
			canadaGe = rs.OverallScore
		}
	}

	if chinaGa <= canadaGe {
		t.Errorf("China gallium risk (%.1f) should be higher than Canada germanium (%.1f)", chinaGa, canadaGe)
	}
}
