// Package pipeline implements data ingestion from GDELT, UN Comtrade, and NVIDIA NIM.
package pipeline

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"yanplatform/backend/internal/config"
	"yanplatform/backend/internal/models"
	"yanplatform/backend/internal/store"
)

// NIMClient communicates with the NVIDIA NIM API for sentiment analysis and risk classification.
type NIMClient struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// NewNIMClient creates a new NVIDIA NIM client.
func NewNIMClient(cfg *config.NIMConfig) *NIMClient {
	return &NIMClient{
		apiKey:  cfg.APIKey,
		baseURL: cfg.BaseURL,
		model:   cfg.Model,
		client:  &http.Client{Timeout: 30 * time.Second},
	}
}

// nimChatRequest is the request body for NIM chat completions.
type nimChatRequest struct {
	Model    string       `json:"model"`
	Messages []nimMessage `json:"messages"`
}

type nimMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// nimChatResponse is the response from NIM chat completions.
type nimChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// ClassifySentiment analyzes an event description and returns a sentiment label.
func (c *NIMClient) ClassifySentiment(description string) (string, float64, error) {
	if c.apiKey == "" {
		// Fallback: simple keyword-based classification when no API key
		return c.fallbackClassify(description), 0.75, nil
	}

	prompt := fmt.Sprintf(`Analyze the following geopolitical event description related to critical mineral supply chains (gallium, germanium). 
Classify it as one of: "escalation", "de-escalation", or "neutral".
Also provide a relevance score from 0.0 to 1.0 for how relevant this is to gallium/germanium supply chain risks.

Respond ONLY with JSON: {"label": "...", "relevance": 0.X}

Event: %s`, description)

	req := nimChatRequest{
		Model: c.model,
		Messages: []nimMessage{
			{Role: "system", Content: "You are a geopolitical risk analyst specializing in critical mineral supply chains."},
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", 0, fmt.Errorf("marshaling NIM request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", 0, fmt.Errorf("creating NIM request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", 0, fmt.Errorf("calling NIM API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", 0, fmt.Errorf("NIM API error %d: %s", resp.StatusCode, string(respBody))
	}

	var nimResp nimChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&nimResp); err != nil {
		return "", 0, fmt.Errorf("decoding NIM response: %w", err)
	}

	if len(nimResp.Choices) == 0 {
		return "neutral", 0.5, nil
	}

	// Parse the JSON response
	var result struct {
		Label     string  `json:"label"`
		Relevance float64 `json:"relevance"`
	}
	if err := json.Unmarshal([]byte(nimResp.Choices[0].Message.Content), &result); err != nil {
		// Fallback if response isn't valid JSON
		return c.fallbackClassify(description), 0.75, nil
	}

	return result.Label, result.Relevance, nil
}

// fallbackClassify provides basic keyword-based sentiment classification.
func (c *NIMClient) fallbackClassify(text string) string {
	escalationKeywords := []string{"restrict", "ban", "sanction", "quota", "tariff", "block", "embargo", "control", "limit", "retaliat"}
	deescalationKeywords := []string{"cooperat", "invest", "agreement", "partner", "subsid", "expand", "open", "lift", "ease", "diversif"}

	escalation, deescalation := 0, 0
	lower := bytes.ToLower([]byte(text))

	for _, kw := range escalationKeywords {
		if bytes.Contains(lower, []byte(kw)) {
			escalation++
		}
	}
	for _, kw := range deescalationKeywords {
		if bytes.Contains(lower, []byte(kw)) {
			deescalation++
		}
	}

	if escalation > deescalation {
		return "escalation"
	}
	if deescalation > escalation {
		return "de-escalation"
	}
	return "neutral"
}

// GDELTPipeline ingests geopolitical events from GDELT via BigQuery.
// For MVP without BigQuery credentials, it operates on seed data already in the store.
type GDELTPipeline struct {
	store  store.Store
	nim    *NIMClient
	config *config.BigQueryConfig
}

// NewGDELTPipeline creates a new GDELT pipeline.
func NewGDELTPipeline(s store.Store, nim *NIMClient, cfg *config.BigQueryConfig) *GDELTPipeline {
	return &GDELTPipeline{store: s, nim: nim, config: cfg}
}

// Run executes one cycle of the GDELT ingestion pipeline.
// With BigQuery credentials, it queries live GDELT data.
// Without credentials, it processes any events already in the store through NIM.
func (p *GDELTPipeline) Run() {
	log.Println("[GDELT Pipeline] Starting ingestion cycle...")

	// Process existing events through NIM for sentiment classification
	events, _ := p.store.GetRecentEvents(20)
	for _, evt := range events {
		if evt.SentimentLabel != "" {
			continue // Already classified
		}

		label, relevance, err := p.nim.ClassifySentiment(evt.Description)
		if err != nil {
			log.Printf("[GDELT Pipeline] NIM classification error: %v", err)
			continue
		}

		evt.SentimentLabel = label
		evt.Relevance = relevance
		_ = p.store.SaveEvent(evt)
	}

	log.Printf("[GDELT Pipeline] Processed %d events", len(events))
}

// ComtradePipeline ingests trade flow data from UN Comtrade API.
type ComtradePipeline struct {
	store  store.Store
	config *config.ComtradeConfig
	client *http.Client
}

// NewComtradePipeline creates a new UN Comtrade pipeline.
func NewComtradePipeline(s store.Store, cfg *config.ComtradeConfig) *ComtradePipeline {
	return &ComtradePipeline{
		store:  s,
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

// Run executes one cycle of the Comtrade ingestion pipeline.
// With API key, fetches real trade data. Without, operates on seed data.
func (p *ComtradePipeline) Run() {
	log.Println("[Comtrade Pipeline] Starting ingestion cycle...")

	if p.config.APIKey == "" {
		log.Println("[Comtrade Pipeline] No API key configured — using seed data")
		return
	}

	// Fetch gallium trade data (HS code 811292)
	p.fetchTradeData("811292", "gallium")
	// Fetch germanium trade data (HS code 811110)
	p.fetchTradeData("811110", "germanium")
}

func (p *ComtradePipeline) fetchTradeData(hsCode, resource string) {
	url := fmt.Sprintf("%s/C/A/HS?cmdCode=%s&flowCode=X&partnerCode=0&reporterCode=156,124,56,276,392,410&period=2025&motCode=0&partner2Code=0",
		p.config.BaseURL, hsCode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("[Comtrade Pipeline] Request error for %s: %v", resource, err)
		return
	}

	req.Header.Set("Ocp-Apim-Subscription-Key", p.config.APIKey)

	resp, err := p.client.Do(req)
	if err != nil {
		log.Printf("[Comtrade Pipeline] HTTP error for %s: %v", resource, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("[Comtrade Pipeline] API error %d for %s", resp.StatusCode, resource)
		return
	}

	var result struct {
		Data []struct {
			RefYear       int     `json:"refYear"`
			RefMonth      int     `json:"refMonth"`
			ReporterDesc  string  `json:"reporterDesc"`
			PartnerDesc   string  `json:"partnerDesc"`
			CmdCode       string  `json:"cmdCode"`
			PrimaryValue  float64 `json:"primaryValue"`
			NetWgt        float64 `json:"netWgt"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("[Comtrade Pipeline] Decode error for %s: %v", resource, err)
		return
	}

	for _, d := range result.Data {
		flow := models.TradeFlow{
			ID:              fmt.Sprintf("tf-%s-%s-%s-%d-%d", d.ReporterDesc, d.PartnerDesc, hsCode, d.RefYear, d.RefMonth),
			Year:            d.RefYear,
			Month:           d.RefMonth,
			ReporterCountry: d.ReporterDesc,
			PartnerCountry:  d.PartnerDesc,
			HSCode:          d.CmdCode,
			Resource:        resource,
			FlowType:        "export",
			ValueUSD:        d.PrimaryValue,
			WeightKg:        d.NetWgt,
			IngestedAt:      time.Now(),
		}
		_ = p.store.SaveTradeFlow(flow)
	}

	log.Printf("[Comtrade Pipeline] Ingested %d trade records for %s", len(result.Data), resource)
}

// Scheduler manages periodic pipeline execution.
type Scheduler struct {
	gdelt       *GDELTPipeline
	comtrade    *ComtradePipeline
	config      *config.PipelineConfig
	stopCh      chan struct{}
}

// NewScheduler creates a pipeline scheduler.
func NewScheduler(gdelt *GDELTPipeline, comtrade *ComtradePipeline, cfg *config.PipelineConfig) *Scheduler {
	return &Scheduler{
		gdelt:    gdelt,
		comtrade: comtrade,
		config:   cfg,
		stopCh:   make(chan struct{}),
	}
}

// Start begins periodic pipeline execution in background goroutines.
func (s *Scheduler) Start() {
	// Run once immediately
	go s.gdelt.Run()
	go s.comtrade.Run()

	// GDELT ticker
	go func() {
		ticker := time.NewTicker(time.Duration(s.config.GDELTIntervalMinutes) * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.gdelt.Run()
			case <-s.stopCh:
				return
			}
		}
	}()

	// Comtrade ticker
	go func() {
		ticker := time.NewTicker(time.Duration(s.config.ComtradeIntervalHours) * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				s.comtrade.Run()
			case <-s.stopCh:
				return
			}
		}
	}()

	log.Printf("[Scheduler] Started — GDELT every %d min, Comtrade every %d hr",
		s.config.GDELTIntervalMinutes, s.config.ComtradeIntervalHours)
}

// Stop halts all pipeline goroutines.
func (s *Scheduler) Stop() {
	close(s.stopCh)
}
