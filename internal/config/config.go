package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                   int
	Host                   string
	DatabasePath           string
	OpenClawGatewayURL     string
	OpenClawGatewayToken   string
	OpenClawConfigPath     string
	SyncInterval           time.Duration
	SyncOnStartup          bool
	Env                    string
	WatchdogInterval       time.Duration // How often the stuck-task watchdog runs (default 5m)
	WatchdogStaleThreshold time.Duration // Time without update before a task is considered stuck (default 30m)
	WatchdogMaxRetries     int           // Max re-notify attempts before resetting task (default 3)
}

func Load() *Config {
	port, _ := strconv.Atoi(getEnv("PORT", "8080"))
	
	// Parse sync interval (default: 5 minutes)
	syncIntervalStr := getEnv("SYNC_INTERVAL", "5m")
	syncInterval, err := time.ParseDuration(syncIntervalStr)
	if err != nil {
		syncInterval = 5 * time.Minute
	}
	
	// Parse sync on startup (default: true)
	syncOnStartup := getEnv("SYNC_ON_STARTUP", "true") == "true"

	// Watchdog: interval (default 5m), stale threshold (default 30m), max retries (default 3)
	watchdogIntervalStr := getEnv("WATCHDOG_INTERVAL", "5m")
	watchdogInterval, err := time.ParseDuration(watchdogIntervalStr)
	if err != nil {
		watchdogInterval = 5 * time.Minute
	}
	watchdogStaleStr := getEnv("WATCHDOG_STALE_THRESHOLD", "30m")
	watchdogStale, errStale := time.ParseDuration(watchdogStaleStr)
	if errStale != nil {
		watchdogStale = 30 * time.Minute
	}
	watchdogMaxRetries, _ := strconv.Atoi(getEnv("WATCHDOG_MAX_RETRIES", "3"))
	if watchdogMaxRetries <= 0 {
		watchdogMaxRetries = 3
	}

	return &Config{
		Port:                   port,
		Host:                   getEnv("HOST", "0.0.0.0"),
		DatabasePath:           getEnv("DATABASE_PATH", "./data/mission-control.db"),
		OpenClawGatewayURL:     getEnv("OPENCLAW_GATEWAY_URL", "ws://127.0.0.1:18789"),
		OpenClawGatewayToken:   getEnv("OPENCLAW_GATEWAY_TOKEN", ""),
		OpenClawConfigPath:     getEnv("OPENCLAW_CONFIG_PATH", ""), // Empty = use default ~/.openclaw/openclaw.json
		SyncInterval:           syncInterval,
		SyncOnStartup:          syncOnStartup,
		Env:                    getEnv("ENV", "development"),
		WatchdogInterval:       watchdogInterval,
		WatchdogStaleThreshold: watchdogStale,
		WatchdogMaxRetries:     watchdogMaxRetries,
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
