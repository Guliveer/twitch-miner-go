// db-seed loads YAML account configs and upserts them into the database.
// It is meant to be run once when migrating from file-based to DB mode.
//
// Usage:
//
//	go run ./cmd/db-seed [--config <dir>] [--dry-run]
//
// The tool reads DB_DSN from the environment (or .env file).
// Schema is created automatically on first connection.
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/Guliveer/twitch-miner-go/internal/config"
	"github.com/Guliveer/twitch-miner-go/internal/store"
)

func main() {
	configDir := flag.String("config", "configs", "path to the YAML config directory")
	dryRun := flag.Bool("dry-run", false, "print what would be inserted without writing to DB")
	flag.Parse()

	_ = godotenv.Load()

	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		fatalf("DB_DSN is not set")
	}

	configs, err := config.LoadAllAccountConfigs(*configDir)
	if err != nil {
		fatalf("loading configs from %s: %v", *configDir, err)
	}
	fmt.Printf("Loaded %d account(s) from %s\n", len(configs), *configDir)

	if *dryRun {
		for _, cfg := range configs {
			fmt.Printf("  [dry-run] would upsert: %s (enabled=%v)\n", cfg.Username, cfg.IsEnabled())
		}
		return
	}

	st, err := store.OpenPostgres(dsn)
	if err != nil {
		fatalf("connecting to database: %v", err)
	}
	defer st.Close()

	var seeded, failed int
	for _, cfg := range configs {
		if err := config.Validate(cfg); err != nil {
			fmt.Fprintf(os.Stderr, "  SKIP  %s: invalid config: %v\n", cfg.Username, err)
			failed++
			continue
		}
		blob, err := config.AccountConfigToJSON(cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR serialising %s: %v\n", cfg.Username, err)
			failed++
			continue
		}
		row := store.AccountRow{
			Username:   cfg.Username,
			ConfigJSON: blob,
			Enabled:    cfg.IsEnabled(),
			UpdatedAt:  time.Now(),
		}
		if err := st.UpsertAccount(row); err != nil {
			fmt.Fprintf(os.Stderr, "  ERROR upserting %s: %v\n", cfg.Username, err)
			failed++
			continue
		}
		fmt.Printf("  OK  %s\n", cfg.Username)
		seeded++
	}

	fmt.Printf("\nDone: %d seeded, %d failed.\n", seeded, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "error: "+format+"\n", args...)
	os.Exit(1)
}
