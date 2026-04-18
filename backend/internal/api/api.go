// Package api provides the HTTP REST API that serves data to the Flutter frontend.
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"yanplatform/backend/internal/config"
	"yanplatform/backend/internal/models"
	"yanplatform/backend/internal/risk"
	"yanplatform/backend/internal/store"
)

// Server is the HTTP API server.
type Server struct {
	store      store.Store
	riskEngine *risk.Engine
	config     *config.ServerConfig
	mux        *http.ServeMux
}

// NewServer creates a new API server.
func NewServer(s store.Store, engine *risk.Engine, cfg *config.ServerConfig) *Server {
	srv := &Server{
		store:      s,
		riskEngine: engine,
		config:     cfg,
		mux:        http.NewServeMux(),
	}

	srv.registerRoutes()
	return srv
}

// Handler returns the HTTP handler with CORS middleware.
func (s *Server) Handler() http.Handler {
	return s.corsMiddleware(s.mux)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /api/risk/overview", s.handleRiskOverview)
	s.mux.HandleFunc("GET /api/risk/chokepoints", s.handleChokepoints)
	s.mux.HandleFunc("GET /api/risk/trends", s.handleRiskTrends)
	s.mux.HandleFunc("GET /api/risk/history", s.handleRiskHistory)
	s.mux.HandleFunc("GET /api/reroute/simulate", s.handleRerouteSimulate)
	s.mux.HandleFunc("GET /api/reroute/latest", s.handleRerouteLatest)
	s.mux.HandleFunc("GET /api/reroute/history", s.handleRerouteHistory)
	s.mux.HandleFunc("GET /api/events/recent", s.handleRecentEvents)
	s.mux.HandleFunc("GET /api/trade/flows", s.handleTradeFlows)
	s.mux.HandleFunc("GET /api/suppliers", s.handleSuppliers)
	s.mux.HandleFunc("GET /api/resources", s.handleResources)
	s.mux.HandleFunc("GET /api/alerts/recent", s.handleAlertsRecent)
	s.mux.HandleFunc("POST /api/alerts/{id}/acknowledge", s.handleAlertAcknowledge)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
}

// --- Handlers ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "yanplatform-supply-chain-risk",
		"version": "0.2.0",
	})
}

func (s *Server) handleRiskOverview(w http.ResponseWriter, r *http.Request) {
	resources, _ := s.store.GetResources()
	events, _ := s.store.GetRecentEvents(100)
	highRisk, _ := s.store.GetHighRiskZones(70)

	resourceRisks := make(map[string]models.RiskScore)

	for _, res := range resources {
		scores, _ := s.store.GetRiskScores(res.ID)
		for _, rs := range scores {
			if rs.Country == res.PrimaryRegion {
				resourceRisks[res.ID] = rs
				break
			}
		}
	}

	overview := models.RiskOverview{
		ResourceRisks: resourceRisks,
		RecentEvents:  len(events),
		HighRiskZones: len(highRisk),
		LastUpdated:   time.Now(),
	}

	s.writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handleChokepoints(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	chokepoints, _ := s.store.GetChokepoints(resource)
	s.writeJSON(w, http.StatusOK, chokepoints)
}

func (s *Server) handleRiskTrends(w http.ResponseWriter, r *http.Request) {
	// Return all risk scores for trend display
	allScores, _ := s.store.GetRiskScores("")
	s.writeJSON(w, http.StatusOK, allScores)
}

func (s *Server) handleRiskHistory(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	daysStr := r.URL.Query().Get("days")
	days := 30
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	history, err := s.store.GetRiskHistory(resource, days)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, history)
}

func (s *Server) handleRerouteSimulate(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	if resource == "" {
		resource = "gallium"
	}

	result := s.riskEngine.SimulateReroute(resource)
	if result == nil {
		s.writeJSON(w, http.StatusOK, map[string]string{
			"status":  "no_disruption",
			"message": "No regions currently exceed the reroute trigger threshold",
		})
		return
	}

	s.writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleRerouteLatest(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	if resource == "" {
		resource = "gallium"
	}

	result, err := s.store.GetLatestRerouteResult(resource)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if result == nil {
		s.writeJSON(w, http.StatusOK, map[string]string{
			"status":  "no_results",
			"message": "No autonomous reroute simulations found for this resource",
		})
		return
	}
	s.writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleRerouteHistory(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	results, err := s.store.GetRerouteResults(resource, limit)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, results)
}

func (s *Server) handleRecentEvents(w http.ResponseWriter, r *http.Request) {
	events, _ := s.store.GetRecentEvents(20)
	s.writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleTradeFlows(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	flows, _ := s.store.GetTradeFlows(resource)
	s.writeJSON(w, http.StatusOK, flows)
}

func (s *Server) handleSuppliers(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	suppliers, _ := s.store.GetSuppliers(resource)
	s.writeJSON(w, http.StatusOK, suppliers)
}

func (s *Server) handleResources(w http.ResponseWriter, r *http.Request) {
	resources, err := s.store.GetResources()
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, resources)
}

func (s *Server) handleAlertsRecent(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	alerts, err := s.store.GetRecentAlerts(limit)
	if err != nil {
		s.writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	s.writeJSON(w, http.StatusOK, alerts)
}

func (s *Server) handleAlertAcknowledge(w http.ResponseWriter, r *http.Request) {
	// Extract ID from path: /api/alerts/{id}/acknowledge
	path := r.URL.Path
	parts := strings.Split(path, "/")
	// Expected: ["", "api", "alerts", "{id}", "acknowledge"]
	if len(parts) < 5 {
		s.writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid path"})
		return
	}
	alertID := parts[3]

	if err := s.store.AcknowledgeAlert(alertID); err != nil {
		s.writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	s.writeJSON(w, http.StatusOK, map[string]string{"status": "acknowledged", "id": alertID})
}

// --- Middleware ---

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		allowed := false
		for _, o := range s.config.AllowedOrigins {
			if o == origin || o == "*" {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(s.config.AllowedOrigins) > 0 {
			w.Header().Set("Access-Control-Allow-Origin", s.config.AllowedOrigins[0])
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// --- Helpers ---

func (s *Server) writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("[API] JSON encode error: %v", err)
	}
}
