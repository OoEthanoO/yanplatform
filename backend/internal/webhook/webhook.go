// Package webhook provides a lightweight webhook client for sending
// supply chain risk alerts to Slack, Discord, or Microsoft Teams.
package webhook

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"yanplatform/backend/internal/config"
)

// AlertPayload contains the data for a webhook notification.
type AlertPayload struct {
	Resource         string    `json:"resource"`
	Region           string    `json:"region"`
	RiskScore        float64   `json:"risk_score"`
	Threshold        float64   `json:"threshold"`
	AlternativeCount int       `json:"alternative_count"`
	RerouteResultID  string    `json:"reroute_result_id"`
	Timestamp        time.Time `json:"timestamp"`
}

// Client sends formatted webhook notifications.
type Client struct {
	url      string
	platform string
	enabled  bool
	client   *http.Client
}

// NewClient creates a new webhook client from config.
func NewClient(cfg *config.WebhookConfig) *Client {
	return &Client{
		url:      cfg.URL,
		platform: cfg.Platform,
		enabled:  cfg.Enabled,
		client:   &http.Client{Timeout: 10 * time.Second},
	}
}

// SendAlert sends a formatted alert to the configured webhook endpoint.
// Returns nil if webhooks are disabled or URL is empty.
func (c *Client) SendAlert(alert AlertPayload) error {
	if !c.enabled || c.url == "" {
		log.Printf("[Webhook] Skipping alert (disabled or no URL configured)")
		return nil
	}

	var payload []byte
	var err error

	switch c.platform {
	case "slack":
		payload, err = c.formatSlack(alert)
	case "discord":
		payload, err = c.formatDiscord(alert)
	case "teams":
		payload, err = c.formatTeams(alert)
	default:
		payload, err = c.formatDiscord(alert) // default to Discord
	}

	if err != nil {
		return fmt.Errorf("formatting webhook payload: %w", err)
	}

	req, err := http.NewRequest("POST", c.url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("creating webhook request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("sending webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	log.Printf("[Webhook] Alert sent successfully for %s in %s (score: %.1f)",
		alert.Resource, alert.Region, alert.RiskScore)
	return nil
}

// formatSlack builds a Slack Block Kit message.
func (c *Client) formatSlack(alert AlertPayload) ([]byte, error) {
	severity := "⚠️ WARNING"
	color := "#ff9800"
	if alert.RiskScore >= 80 {
		severity = "🚨 CRITICAL"
		color = "#f44336"
	}

	msg := map[string]any{
		"attachments": []map[string]any{
			{
				"color": color,
				"blocks": []map[string]any{
					{
						"type": "header",
						"text": map[string]string{
							"type": "plain_text",
							"text": fmt.Sprintf("%s SUPPLY CHAIN ALERT", severity),
						},
					},
					{
						"type": "section",
						"text": map[string]string{
							"type": "mrkdwn",
							"text": fmt.Sprintf(
								"*%s* risk in *%s* has reached *%.1f/100* (threshold: %.1f)\n\nAutonomous reroute simulation complete. *%d alternative supplier(s)* identified.",
								capitalize(alert.Resource), alert.Region,
								alert.RiskScore, alert.Threshold,
								alert.AlternativeCount,
							),
						},
					},
					{
						"type": "context",
						"elements": []map[string]string{
							{
								"type": "mrkdwn",
								"text": fmt.Sprintf("📊 YanPlatform · %s", alert.Timestamp.Format("2006-01-02 15:04:05 UTC")),
							},
						},
					},
				},
			},
		},
	}
	return json.Marshal(msg)
}

// formatDiscord builds a Discord embed message.
func (c *Client) formatDiscord(alert AlertPayload) ([]byte, error) {
	severity := "⚠️ WARNING"
	color := 0xFF9800 // orange
	if alert.RiskScore >= 80 {
		severity = "🚨 CRITICAL"
		color = 0xF44336 // red
	}

	msg := map[string]any{
		"content": fmt.Sprintf("%s **SUPPLY CHAIN ALERT**", severity),
		"embeds": []map[string]any{
			{
				"title":       fmt.Sprintf("%s Risk — %s", capitalize(alert.Resource), alert.Region),
				"description": fmt.Sprintf("Risk score **%.1f/100** exceeds threshold of %.1f.\n\nAutonomous reroute simulation complete. **%d alternative supplier(s)** identified.", alert.RiskScore, alert.Threshold, alert.AlternativeCount),
				"color":       color,
				"fields": []map[string]any{
					{"name": "Resource", "value": capitalize(alert.Resource), "inline": true},
					{"name": "Region", "value": alert.Region, "inline": true},
					{"name": "Risk Score", "value": fmt.Sprintf("%.1f / 100", alert.RiskScore), "inline": true},
					{"name": "Alternatives Found", "value": fmt.Sprintf("%d", alert.AlternativeCount), "inline": true},
				},
				"footer": map[string]string{
					"text": "YanPlatform — Autonomous Supply Chain Intelligence",
				},
				"timestamp": alert.Timestamp.Format(time.RFC3339),
			},
		},
	}
	return json.Marshal(msg)
}

// formatTeams builds a Microsoft Teams Adaptive Card / MessageCard.
func (c *Client) formatTeams(alert AlertPayload) ([]byte, error) {
	severity := "⚠️ WARNING"
	themeColor := "FF9800"
	if alert.RiskScore >= 80 {
		severity = "🚨 CRITICAL"
		themeColor = "F44336"
	}

	msg := map[string]any{
		"@type":      "MessageCard",
		"@context":   "http://schema.org/extensions",
		"themeColor": themeColor,
		"summary":    fmt.Sprintf("%s: %s risk in %s", severity, capitalize(alert.Resource), alert.Region),
		"sections": []map[string]any{
			{
				"activityTitle":    fmt.Sprintf("%s SUPPLY CHAIN ALERT", severity),
				"activitySubtitle": "YanPlatform — Autonomous Supply Chain Intelligence",
				"facts": []map[string]string{
					{"name": "Resource", "value": capitalize(alert.Resource)},
					{"name": "Region", "value": alert.Region},
					{"name": "Risk Score", "value": fmt.Sprintf("%.1f / 100", alert.RiskScore)},
					{"name": "Threshold", "value": fmt.Sprintf("%.1f", alert.Threshold)},
					{"name": "Alternatives Found", "value": fmt.Sprintf("%d", alert.AlternativeCount)},
				},
				"text": fmt.Sprintf("Autonomous reroute simulation complete. **%d alternative supplier(s)** identified.", alert.AlternativeCount),
			},
		},
	}
	return json.Marshal(msg)
}

// capitalize returns the input with the first letter uppercased.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}
