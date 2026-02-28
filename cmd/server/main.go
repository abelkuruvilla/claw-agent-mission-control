package main

import (
	"context"
	"database/sql"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/abelkuruvilla/claw-agent-mission-control/internal/api"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/config"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/db"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/openclaw"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/queue"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/store"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/sync"
	"github.com/abelkuruvilla/claw-agent-mission-control/internal/ui"
)

func main() {
	// Load config
	cfg := config.Load()

	// Ensure data directory exists
	if err := db.EnsureDataDir(cfg.DatabasePath); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Initialize database
	sqlDB, err := sql.Open("sqlite3", cfg.DatabasePath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}
	defer sqlDB.Close()

	// Verify connection
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	log.Println("Running database migrations...")
	if err := db.Migrate(sqlDB); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}
	log.Println("Database ready")

	// Create store
	st := store.New(sqlDB)

	// Create OpenClaw config reader
	configReader := openclaw.NewConfigReader(cfg.OpenClawConfigPath)
	log.Printf("Using OpenClaw config: %s", configReader.GetConfigPath())

	// Create sync service
	syncService := sync.NewSyncService(st, configReader)

	// Sync on startup if enabled
	if cfg.SyncOnStartup {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		
		if err := syncService.SyncOnce(ctx); err != nil {
			log.Printf("Warning: Initial sync failed: %v", err)
		}
	}

	// Start periodic sync
	ctx := context.Background()
	syncService.StartPeriodicSync(ctx, cfg.SyncInterval)

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// Create and start server
	server := api.NewServer(cfg, st)

	// Serve embedded UI
	assets, err := ui.Assets()
	if err != nil {
		log.Println("Warning: Could not load UI assets:", err)
	} else {
		server.ServeUI(assets)
	}

	// Start task queue processor (checks every 10 minutes for queued tasks)
	queueProcessor := queue.NewProcessor(st, server.AgentSender(), server.Hub(), server.TaskHandler())
	queueProcessor.Start(ctx, 10*time.Minute)

	// Start stuck-task watchdog (re-notifies or resets tasks stuck in active states)
	watchdog := queue.NewWatchdog(st, server.Hub(), server.TaskHandler(), cfg.WatchdogStaleThreshold, cfg.WatchdogMaxRetries)
	watchdog.Start(ctx, cfg.WatchdogInterval)

	// Start server in goroutine
	go func() {
		log.Printf("Starting Claw Agent Mission Control on %s:%d", cfg.Host, cfg.Port)
		if err := server.Start(); err != nil {
			log.Fatal("Server error:", err)
		}
	}()

	// Wait for shutdown signal
	<-sigChan
	log.Println("Shutting down gracefully...")
	
	// Stop background services
	watchdog.Stop()
	queueProcessor.Stop()
	syncService.StopPeriodicSync()
	
	log.Println("Shutdown complete")
}
