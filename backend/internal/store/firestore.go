package store

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
	"yanplatform/backend/internal/models"
)

// FirestoreStore provides a Firestore-backed implementation of Store.
type FirestoreStore struct {
	client *firestore.Client
}

// NewFirestoreStore creates a new Firestore store.
func NewFirestoreStore(projectID string) (*FirestoreStore, error) {
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create firestore client: %w", err)
	}

	return &FirestoreStore{client: client}, nil
}

// Close closes the Firestore client.
func (s *FirestoreStore) Close() error {
	return s.client.Close()
}

// LoadSupplierSeed is only for MemoryStore; Firestore persists data, so we can mock or seed manually.
func (s *FirestoreStore) LoadSupplierSeed(path string) error {
	return nil // No-op for Firestore
}

// --- Suppliers ---

// GetSuppliers returns all suppliers, optionally filtered by resource.
func (s *FirestoreStore) GetSuppliers(resource string) ([]models.Supplier, error) {
	ctx := context.Background()
	var suppliers []models.Supplier

	q := s.client.Collection("suppliers").Query
	if resource != "" {
		q = q.Where("resource", "==", resource)
	}

	iter := q.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("querying suppliers: %w", err)
		}

		var sup models.Supplier
		if err := doc.DataTo(&sup); err != nil {
			return nil, fmt.Errorf("parsing supplier %s: %w", doc.Ref.ID, err)
		}
		suppliers = append(suppliers, sup)
	}

	return suppliers, nil
}

// GetAlternativeSuppliers returns suppliers that are alternative reroute candidates.
func (s *FirestoreStore) GetAlternativeSuppliers(resource string) ([]models.Supplier, error) {
	ctx := context.Background()
	var alternatives []models.Supplier

	q := s.client.Collection("suppliers").Where("is_alternative", "==", true)
	if resource != "" {
		q = q.Where("resource", "==", resource)
	}

	iter := q.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("querying alternative suppliers: %w", err)
		}

		var sup models.Supplier
		if err := doc.DataTo(&sup); err != nil {
			return nil, fmt.Errorf("parsing supplier %s: %w", doc.Ref.ID, err)
		}
		alternatives = append(alternatives, sup)
	}

	return alternatives, nil
}

// --- Risk Scores ---

// SaveRiskScore upserts a risk score.
func (s *FirestoreStore) SaveRiskScore(score models.RiskScore) error {
	ctx := context.Background()
	_, err := s.client.Collection("riskScores").Doc(score.ID).Set(ctx, score)
	if err != nil {
		return fmt.Errorf("saving risk score: %w", err)
	}
	return nil
}

// GetRiskScores returns all risk scores, optionally filtered by resource.
func (s *FirestoreStore) GetRiskScores(resource string) ([]models.RiskScore, error) {
	ctx := context.Background()
	var riskScores []models.RiskScore

	q := s.client.Collection("riskScores").Query
	if resource != "" {
		q = q.Where("resource", "==", resource)
	}

	iter := q.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("querying risk scores: %w", err)
		}

		var rs models.RiskScore
		if err := doc.DataTo(&rs); err != nil {
			return nil, fmt.Errorf("parsing risk score: %w", err)
		}
		riskScores = append(riskScores, rs)
	}

	return riskScores, nil
}

// GetHighRiskZones returns risk scores above the threshold.
func (s *FirestoreStore) GetHighRiskZones(threshold float64) ([]models.RiskScore, error) {
	ctx := context.Background()
	var zones []models.RiskScore

	q := s.client.Collection("riskScores").Where("overall_score", ">=", threshold)
	iter := q.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("querying high risk zones: %w", err)
		}

		var rs models.RiskScore
		if err := doc.DataTo(&rs); err != nil {
			return nil, fmt.Errorf("parsing high risk zone: %w", err)
		}
		zones = append(zones, rs)
	}

	return zones, nil
}

// --- Events ---

// SaveEvent adds a GDELT event.
func (s *FirestoreStore) SaveEvent(event models.GDELTEvent) error {
	ctx := context.Background()
	_, err := s.client.Collection("events").Doc(event.ID).Set(ctx, event)
	if err != nil {
		return fmt.Errorf("saving event: %w", err)
	}
	return nil
}

// GetRecentEvents returns the most recent N events.
func (s *FirestoreStore) GetRecentEvents(limit int) ([]models.GDELTEvent, error) {
	ctx := context.Background()
	var events []models.GDELTEvent

	q := s.client.Collection("events").OrderBy("event_date", firestore.Desc)
	if limit > 0 {
		q = q.Limit(limit)
	}

	iter := q.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("querying recent events: %w", err)
		}

		var evt models.GDELTEvent
		if err := doc.DataTo(&evt); err != nil {
			return nil, fmt.Errorf("parsing event: %w", err)
		}

		// Adjust unmarshaled times (Firestore struct mapping might lose local tz accuracy but it's UTC anyway in backend)
		events = append(events, evt)
	}

	return events, nil
}

// --- Trade Flows ---

// SaveTradeFlow adds a trade flow record.
func (s *FirestoreStore) SaveTradeFlow(flow models.TradeFlow) error {
	ctx := context.Background()
	_, err := s.client.Collection("tradeFlows").Doc(flow.ID).Set(ctx, flow)
	if err != nil {
		return fmt.Errorf("saving trade flow: %w", err)
	}
	return nil
}

// GetTradeFlows returns trade flows filtered by resource.
func (s *FirestoreStore) GetTradeFlows(resource string) ([]models.TradeFlow, error) {
	ctx := context.Background()
	var flows []models.TradeFlow

	q := s.client.Collection("tradeFlows").Query
	if resource != "" {
		q = q.Where("resource", "==", resource)
	}

	iter := q.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("querying trade flows: %w", err)
		}

		var tf models.TradeFlow
		if err := doc.DataTo(&tf); err != nil {
			return nil, fmt.Errorf("parsing trade flow: %w", err)
		}
		flows = append(flows, tf)
	}

	return flows, nil
}

// --- Reroute Results ---

// SaveRerouteResult stores a reroute simulation result.
func (s *FirestoreStore) SaveRerouteResult(result models.RerouteResult) error {
	ctx := context.Background()
	_, err := s.client.Collection("rerouteResults").Doc(result.ID).Set(ctx, result)
	if err != nil {
		return fmt.Errorf("saving reroute result: %w", err)
	}
	return nil
}

// GetLatestRerouteResult returns the most recent reroute for a resource.
func (s *FirestoreStore) GetLatestRerouteResult(resource string) (*models.RerouteResult, error) {
	ctx := context.Background()
	
	q := s.client.Collection("rerouteResults").OrderBy("simulated_at", firestore.Desc)
	if resource != "" {
		q = s.client.Collection("rerouteResults").Where("resource", "==", resource).OrderBy("simulated_at", firestore.Desc)
	}
	q = q.Limit(1)

	iter := q.Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, fmt.Errorf("querying latest reroute result: %w", err)
	}

	var result models.RerouteResult
	if err := doc.DataTo(&result); err != nil {
		return nil, fmt.Errorf("parsing reroute result: %w", err)
	}

	return &result, nil
}

// --- Chokepoints ---

// SaveChokepoint upserts a chokepoint.
func (s *FirestoreStore) SaveChokepoint(cp models.Chokepoint) error {
	ctx := context.Background()
	_, err := s.client.Collection("chokepoints").Doc(cp.ID).Set(ctx, cp)
	if err != nil {
		return fmt.Errorf("saving chokepoint: %w", err)
	}
	return nil
}

// GetChokepoints returns all chokepoints, optionally filtered by resource.
func (s *FirestoreStore) GetChokepoints(resource string) ([]models.Chokepoint, error) {
	ctx := context.Background()
	var chokepoints []models.Chokepoint

	q := s.client.Collection("chokepoints").Query
	if resource != "" {
		q = q.Where("resource", "==", resource)
	}

	iter := q.Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("querying chokepoints: %w", err)
		}

		var cp models.Chokepoint
		if err := doc.DataTo(&cp); err != nil {
			return nil, fmt.Errorf("parsing chokepoint: %w", err)
		}
		chokepoints = append(chokepoints, cp)
	}

	return chokepoints, nil
}

// --- Seed Helpers ---

// SeedInitialData relies on the MemoryStore seeding logic, but we could duplicate it here.
// For now, if we want to seed Firestore, we just invoke it once manually or use a migration script.
func (s *FirestoreStore) SeedInitialData() error {
	now := time.Now()

	// Seed some base entities to ensure it functions if database is empty...
	chokepoints := []models.Chokepoint{
		{
			ID: "cp-china-yunnan-ga", Name: "Yunnan Gallium Processing Hub",
			Type: "production", Country: "China", Region: "Yunnan Province",
			GlobalSharePct: 40.0, Resource: "gallium", RiskLevel: "critical",
			Latitude: 25.0389, Longitude: 102.7183,
		},
		{
			ID: "cp-malacca-strait", Name: "Strait of Malacca",
			Type: "shipping", Country: "International", Region: "Southeast Asia",
			GlobalSharePct: 60.0, Resource: "gallium", RiskLevel: "elevated",
			Latitude: 2.5, Longitude: 101.8,
		},
	}

	for _, cp := range chokepoints {
		_ = s.SaveChokepoint(cp)
	}
	
	riskScore := models.RiskScore{
		ID: "risk-china-ga", Region: "China", Country: "China", Resource: "gallium",
		OverallScore: 82.0, SupplyConcentration: 95.0, GeopoliticalTension: 75.0,
		TradePolicySignal: 80.0, LogisticsRisk: 50.0,
		ComputedAt: now, IsHighRisk: true,
	}
	_ = s.SaveRiskScore(riskScore)
	
	return nil
}
