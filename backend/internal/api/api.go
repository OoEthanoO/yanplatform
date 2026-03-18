// Package api provides the HTTP REST API that serves data to the Flutter frontend.
package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"yanplatform/backend/internal/config"
	"yanplatform/backend/internal/risk"
	"yanplatform/backend/internal/store"
	"yanplatform/backend/internal/models"
)

// Server is the HTTP API server.
type Server struct {
	store      *store.Store
	riskEngine *risk.Engine
	config     *config.ServerConfig
	mux        *http.ServeMux
}

// NewServer creates a new API server.
func NewServer(s *store.Store, engine *risk.Engine, cfg *config.ServerConfig) *Server {
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
	s.mux.HandleFunc("GET /api/reroute/simulate", s.handleRerouteSimulate)
	s.mux.HandleFunc("GET /api/events/recent", s.handleRecentEvents)
	s.mux.HandleFunc("GET /api/trade/flows", s.handleTradeFlows)
	s.mux.HandleFunc("GET /api/suppliers", s.handleSuppliers)
	s.mux.HandleFunc("GET /api/health", s.handleHealth)
}

// --- Handlers ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"service": "yanplatform-supply-chain-risk",
		"version": "0.1.0",
	})
}

func (s *Server) handleRiskOverview(w http.ResponseWriter, r *http.Request) {
	gaScores := s.store.GetRiskScores("gallium")
	geScores := s.store.GetRiskScores("germanium")
	events := s.store.GetRecentEvents(100)
	highRisk := s.store.GetHighRiskZones(70)

	var gaRisk, geRisk models.RiskScore
	for _, rs := range gaScores {
		if rs.Country == "China" {
			gaRisk = rs
			break
		}
	}
	for _, rs := range geScores {
		if rs.Country == "China" {
			geRisk = rs
			break
		}
	}

	overview := models.RiskOverview{
		GalliumRisk:   gaRisk,
		GermaniumRisk: geRisk,
		RecentEvents:  len(events),
		HighRiskZones: len(highRisk),
		LastUpdated:   time.Now(),
	}

	s.writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handleChokepoints(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	chokepoints := s.store.GetChokepoints(resource)
	s.writeJSON(w, http.StatusOK, chokepoints)
}

func (s *Server) handleRiskTrends(w http.ResponseWriter, r *http.Request) {
	// Return all risk scores for trend display
	allScores := s.store.GetRiskScores("")
	s.writeJSON(w, http.StatusOK, allScores)
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

func (s *Server) handleRecentEvents(w http.ResponseWriter, r *http.Request) {
	events := s.store.GetRecentEvents(20)
	s.writeJSON(w, http.StatusOK, events)
}

func (s *Server) handleTradeFlows(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	flows := s.store.GetTradeFlows(resource)
	s.writeJSON(w, http.StatusOK, flows)
}

func (s *Server) handleSuppliers(w http.ResponseWriter, r *http.Request) {
	resource := r.URL.Query().Get("resource")
	suppliers := s.store.GetSuppliers(resource)
	s.writeJSON(w, http.StatusOK, suppliers)
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
