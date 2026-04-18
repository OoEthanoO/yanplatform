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
	"yanplatform/backend/internal/webhook"
)

func main() {
	log.Println("╔══════════════════════════════════════════════════════╗")
	log.Println("║  YanPlatform — Supply Chain Intelligence             ║")
	log.Println("║  v0.2.0 — Phase 4: Active Alerting & Telemetry      ║")
	log.Println("╚══════════════════════════════════════════════════════╝")

	// Load configuration
	cfg, err := config.Load("config.json")
	if err != nil {
		log.Printf("Warning: Could not load config file: %v (using defaults)", err)
		cfg = config.DefaultConfig()
	}

	var dataStore store.Store
	if cfg.Firebase.UseFirestore {
		fsStore, err := store.NewFirestoreStore(cfg.Firebase.ProjectID)
		if err != nil {
			log.Fatalf("[Store] Failed to initialize Firestore: %v", err)
		}
		dataStore = fsStore
		log.Println("[Store] Initialized Firestore backend")
	} else {
		dataStore = store.NewMemoryStore()
		
		seedPath := filepath.Join("data", "suppliers_seed.json")
		if err := dataStore.LoadSupplierSeed(seedPath); err != nil {
			log.Printf("Warning: Could not load supplier seed data: %v", err)
		} else {
			log.Println("[Store] Loaded supplier seed data")
		}

		dataStore.SeedInitialData()
		log.Println("[Store] Seeded initial data (resources, chokepoints, risk scores, events, 30-day history, alerts)")
	}

	// Initialize risk engine
	riskEngine := risk.NewEngine(dataStore, &cfg.Risk)

	// Recalculate risk scores with seed data
	riskEngine.RecalculateAll()
	log.Println("[Risk Engine] Initial risk scores computed + daily snapshots saved")

	// Initialize pipeline clients
	nimClient := pipeline.NewNIMClient(&cfg.NIM)
	gdeltPipeline := pipeline.NewGDELTPipeline(dataStore, nimClient, &cfg.BigQuery)
	comtradePipeline := pipeline.NewComtradePipeline(dataStore, &cfg.Comtrade)

	// Initialize webhook client
	webhookClient := webhook.NewClient(&cfg.Webhook)
	if cfg.Webhook.Enabled {
		log.Printf("[Webhook] Enabled — platform: %s", cfg.Webhook.Platform)
	} else {
		log.Println("[Webhook] Disabled (set WEBHOOK_URL to enable)")
	}

	// Start pipeline scheduler
	scheduler := pipeline.NewScheduler(gdeltPipeline, comtradePipeline, riskEngine, webhookClient, &cfg.Pipeline)
	scheduler.Start()
	log.Println("[Scheduler] Pipeline scheduler started")

	// Initialize and start API server
	server := api.NewServer(dataStore, riskEngine, &cfg.Server)

	addr := ":" + cfg.Server.Port
	log.Printf("[Server] Starting on %s", addr)
	log.Printf("[Server] API endpoints:")
	log.Printf("  GET  /api/health                    — Health check")
	log.Printf("  GET  /api/risk/overview              — Risk dashboard overview")
	log.Printf("  GET  /api/risk/chokepoints            — Chokepoint map data")
	log.Printf("  GET  /api/risk/trends                — Risk score trends")
	log.Printf("  GET  /api/risk/history                — Time-series risk history")
	log.Printf("  GET  /api/reroute/simulate            — On-demand reroute simulation")
	log.Printf("  GET  /api/reroute/latest              — Latest autonomous reroute result")
	log.Printf("  GET  /api/reroute/history             — Reroute simulation history")
	log.Printf("  GET  /api/events/recent               — Recent geopolitical events")
	log.Printf("  GET  /api/trade/flows                 — Trade flow data")
	log.Printf("  GET  /api/suppliers                   — Supplier directory")
	log.Printf("  GET  /api/resources                   — Tracked resources")
	log.Printf("  GET  /api/alerts/recent               — Recent system alerts")
	log.Printf("  POST /api/alerts/{id}/acknowledge     — Acknowledge an alert")

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
