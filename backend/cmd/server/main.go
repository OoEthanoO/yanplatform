// Server entry point for the YanPlatform API.
package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"yanplatform/backend/internal/api"
	"yanplatform/backend/internal/config"
	"yanplatform/backend/internal/pipeline"
	"yanplatform/backend/internal/risk"
	"yanplatform/backend/internal/store"
)

func main() {
	log.Println("╔══════════════════════════════════════════════════════╗")
	log.Println("║  YanPlatform — Gallium/Germanium                     ║")
	log.Println("║  MVP v0.1.0                                         ║")
	log.Println("╚══════════════════════════════════════════════════════╝")

	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: Could not load config file: %v (using defaults)", err)
		cfg = config.DefaultConfig()
	}

	// Initialize store
	dataStore := store.New()

	// Load supplier seed data
	seedPath := filepath.Join("data", "suppliers_seed.json")
	if err := dataStore.LoadSupplierSeed(seedPath); err != nil {
		log.Printf("Warning: Could not load supplier seed data: %v", err)
	} else {
		log.Println("[Store] Loaded supplier seed data")
	}

	// Seed initial data (chokepoints, risk scores, sample events)
	dataStore.SeedInitialData()
	log.Println("[Store] Seeded initial chokepoints, risk scores, and events")

	// Initialize risk engine
	riskEngine := risk.NewEngine(dataStore, &cfg.Risk)

	// Recalculate risk scores with seed data
	riskEngine.RecalculateAll()
	log.Println("[Risk Engine] Initial risk scores computed")

	// Initialize pipeline clients
	nimClient := pipeline.NewNIMClient(&cfg.NIM)
	gdeltPipeline := pipeline.NewGDELTPipeline(dataStore, nimClient, &cfg.BigQuery)
	comtradePipeline := pipeline.NewComtradePipeline(dataStore, &cfg.Comtrade)

	// Start pipeline scheduler
	scheduler := pipeline.NewScheduler(gdeltPipeline, comtradePipeline, &cfg.Pipeline)
	scheduler.Start()
	log.Println("[Scheduler] Pipeline scheduler started")

	// Initialize and start API server
	server := api.NewServer(dataStore, riskEngine, &cfg.Server)

	addr := ":" + cfg.Server.Port
	log.Printf("[Server] Starting on %s", addr)
	log.Printf("[Server] API endpoints:")
	log.Printf("  GET /api/health              — Health check")
	log.Printf("  GET /api/risk/overview        — Risk dashboard overview")
	log.Printf("  GET /api/risk/chokepoints     — Chokepoint map data")
	log.Printf("  GET /api/risk/trends          — Risk score trends")
	log.Printf("  GET /api/reroute/simulate     — Shadow reroute simulation")
	log.Printf("  GET /api/events/recent        — Recent geopolitical events")
	log.Printf("  GET /api/trade/flows           — Trade flow data")
	log.Printf("  GET /api/suppliers             — Supplier directory")

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("\n[Server] Shutting down...")
		scheduler.Stop()
		os.Exit(0)
	}()

	if err := http.ListenAndServe(addr, server.Handler()); err != nil {
		log.Fatalf("[Server] Fatal: %v", err)
	}
}
