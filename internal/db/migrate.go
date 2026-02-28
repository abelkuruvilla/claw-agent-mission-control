package db

import (
	"database/sql"
	"embed"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// Migrate runs all pending migrations
func Migrate(db *sql.DB) error {
	// Create migrations tracking table if not exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Get list of migration files
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	// Sort migration files
	var files []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Apply each migration
	for _, file := range files {
		version := strings.TrimSuffix(file, ".up.sql")

		// Check if already applied
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE version = ?", version).Scan(&count)
		if err != nil {
			return err
		}
		if count > 0 {
			continue // Already applied
		}

		// Read and execute migration
		content, err := migrationsFS.ReadFile("migrations/" + file)
		if err != nil {
			return err
		}

		_, err = db.Exec(string(content))
		if err != nil {
			// Ignore "already exists" errors for idempotency
			if !strings.Contains(err.Error(), "already exists") {
				log.Printf("Migration %s failed: %v", file, err)
				return err
			}
		}

		// Record migration
		_, err = db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version)
		if err != nil {
			return err
		}

		log.Printf("Applied migration: %s", file)
	}

	return nil
}

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	if dir == "" || dir == "." {
		dir = "./data"
	}
	return os.MkdirAll(dir, 0755)
}
