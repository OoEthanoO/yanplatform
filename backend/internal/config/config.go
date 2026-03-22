// Package config provides configuration management for the application.
package config

import (
	"encoding/json"
	"os"
)

// Config holds all application configuration.
type Config struct {
	Server    ServerConfig    `json:"server"`
	Firebase  FirebaseConfig  `json:"firebase"`
	BigQuery  BigQueryConfig  `json:"bigquery"`
	NIM       NIMConfig       `json:"nim"`
	Comtrade  ComtradeConfig  `json:"comtrade"`
	Pipeline  PipelineConfig  `json:"pipeline"`
	Risk      RiskConfig      `json:"risk"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port string `json:"port"`
	// AllowedOrigins for CORS (Flutter Web dev server)
	AllowedOrigins []string `json:"allowed_origins"`
}

// FirebaseConfig holds Firebase/Firestore settings.
type FirebaseConfig struct {
	ProjectID          string `json:"project_id"`
	CredentialsFile    string `json:"credentials_file"`
	UseFirestore       bool   `json:"use_firestore"`
}

// BigQueryConfig holds BigQuery settings for GDELT.
type BigQueryConfig struct {
	ProjectID string `json:"project_id"`
	// GDELT dataset is public: gdelt-bq.gdeltv2
	GDELTDataset string `json:"gdelt_dataset"`
}

// NIMConfig holds NVIDIA NIM API settings.
type NIMConfig struct {
	APIKey   string `json:"api_key"`
	BaseURL  string `json:"base_url"`
	Model    string `json:"model"`
}

// ComtradeConfig holds UN Comtrade API settings.
type ComtradeConfig struct {
	APIKey  string `json:"api_key"`
	BaseURL string `json:"base_url"`
}

// PipelineConfig holds data pipeline scheduling settings.
type PipelineConfig struct {
	GDELTIntervalMinutes    int `json:"gdelt_interval_minutes"`
	ComtradeIntervalHours   int `json:"comtrade_interval_hours"`
	RiskRecalcIntervalMins  int `json:"risk_recalc_interval_mins"`
}

// RiskConfig holds risk scoring thresholds.
type RiskConfig struct {
	HighRiskThreshold       float64 `json:"high_risk_threshold"`
	RerouteTriggerThreshold float64 `json:"reroute_trigger_threshold"`
	// Weights for risk score components (must sum to 1.0)
	WeightSupplyConcentration float64 `json:"weight_supply_concentration"`
	WeightGeopoliticalTension float64 `json:"weight_geopolitical_tension"`
	WeightTradePolicySignal   float64 `json:"weight_trade_policy_signal"`
	WeightLogisticsRisk       float64 `json:"weight_logistics_risk"`
}

// Load reads configuration from a JSON file, with environment variable overrides.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// Try loading from file
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := json.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	}

	// Environment variable overrides
	if v := os.Getenv("PORT"); v != "" {
		cfg.Server.Port = v
	}
	if v := os.Getenv("FIREBASE_PROJECT_ID"); v != "" {
		cfg.Firebase.ProjectID = v
	}
	if v := os.Getenv("USE_FIRESTORE"); v == "true" || v == "1" {
		cfg.Firebase.UseFirestore = true
	}
	if v := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); v != "" {
		cfg.Firebase.CredentialsFile = v
	}
	if v := os.Getenv("BIGQUERY_PROJECT_ID"); v != "" {
		cfg.BigQuery.ProjectID = v
	}
	if v := os.Getenv("NIM_API_KEY"); v != "" {
		cfg.NIM.APIKey = v
	}
	if v := os.Getenv("COMTRADE_API_KEY"); v != "" {
		cfg.Comtrade.APIKey = v
	}

	return cfg, nil
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:           "8080",
			AllowedOrigins: []string{"*", "http://localhost:3000", "http://localhost:8080", "http://localhost:5000", "http://localhost:9090"},
		},
		Firebase: FirebaseConfig{
			ProjectID: "yanplatform-dev",
		},
		BigQuery: BigQueryConfig{
			ProjectID:    "yanplatform-dev",
			GDELTDataset: "gdelt-bq.gdeltv2",
		},
		NIM: NIMConfig{
			BaseURL: "https://integrate.api.nvidia.com/v1",
			Model:   "meta/llama-3.1-8b-instruct",
		},
		Comtrade: ComtradeConfig{
			BaseURL: "https://comtradeapi.un.org/data/v1/get",
		},
		Pipeline: PipelineConfig{
			GDELTIntervalMinutes:   60,
			ComtradeIntervalHours:  24,
			RiskRecalcIntervalMins: 30,
		},
		Risk: RiskConfig{
			HighRiskThreshold:         70.0,
			RerouteTriggerThreshold:   70.0,
			WeightSupplyConcentration: 0.40,
			WeightGeopoliticalTension: 0.30,
			WeightTradePolicySignal:   0.20,
			WeightLogisticsRisk:       0.10,
		},
	}
}
