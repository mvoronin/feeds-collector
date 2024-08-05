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
	flag.Parse()

	configPath := flag.String("config", "config.yaml", "path to cfg file")
	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatalf("Error reading cfg file: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}

	closeInfoLoggingFunc, err := internal.InitLogging(cfg.Logging.InfoLog, internal.InfoLogLevel)
	if err != nil {
		log.Fatalf("Error initializing logging: %v", err)
	}
	defer closeInfoLoggingFunc()
	closeErrorLoggingFunc, err := internal.InitLogging(cfg.Logging.ErrorLog, internal.ErrorLogLevel)
	if err != nil {
		log.Fatalf("Error initializing logging: %v", err)
	}
	defer closeErrorLoggingFunc()

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
		log.Fatalf("Error running migrations: %v", err)
	}

	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)
	go server.RunAPIServer(ctxWithCancel, db, cfg)
	go gatherer.RunGathererLoop(ctxWithCancel, db, cfg)

	// Handle graceful shutdown on Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	cancel()
}
