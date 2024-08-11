package main

import (
	"context"
	"database/sql"
	"feedscollector/internal"
	"feedscollector/internal/gatherer"
	"feedscollector/internal/infrastructure/config"
	"feedscollector/internal/infrastructure/persistence"
	"feedscollector/internal/server"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Application failed: %v", err)
	}
}

func run() error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if err := setupLogging(cfg); err != nil {
		return err
	}

	db, err := sql.Open("sqlite3", cfg.Database.Path)
	if err != nil {
		internal.ErrorLogger.Fatalf("Error opening database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			internal.ErrorLogger.Fatalf("Error closing database: %v", err)
		}
	}(db)

	// Run migrations
	if err := persistence.Migrate(db); err != nil {
		return err
	}

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	go server.RunAPIServer(ctxWithCancel, db, cfg)
	go gatherer.RunGathererLoop(ctxWithCancel, db, cfg)

	// Handle graceful shutdown on Ctrl+C
	handleShutdown(cancel)

	return nil
}

func loadConfig() (*config.Config, error) {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		return nil, err
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func setupLogging(cfg *config.Config) error {
	closeInfoLoggingFunc, err := internal.InitLogging(cfg.Logging.InfoLog, internal.InfoLogLevel)
	if err != nil {
		return err
	}
	defer closeInfoLoggingFunc()

	closeErrorLoggingFunc, err := internal.InitLogging(cfg.Logging.ErrorLog, internal.ErrorLogLevel)
	if err != nil {
		return err
	}
	defer closeErrorLoggingFunc()

	return nil
}

func handleShutdown(cancel context.CancelFunc) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	cancel()
}
