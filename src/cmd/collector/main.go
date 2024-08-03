package main

import (
	"context"
	"database/sql"
	"feedscollector/internal"
	"feedscollector/internal/gatherer"
	"feedscollector/internal/infrastructure/config"
	"feedscollector/internal/server"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
)

func parseConfig() *config.Config {
	configPath := flag.String("config", "config.yaml", "path to cfg file")
	flag.Parse()

	log.Printf("Reading a cfg file: %s\n", *configPath)
	cfg, err := config.ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error reading cfg file: %v", err)
	}
	if err := config.ValidateConfig(cfg); err != nil {
		log.Fatalf("Invalid config: %v", err)
	}
	return cfg
}

func main() {
	cfg := parseConfig()

	closeInfoLogFile, err := internal.InitLogging(cfg.Logging.InfoLog, internal.InfoLogLevel)
	if err != nil {
		log.Fatal(err)
	}
	if closeInfoLogFile != nil {
		defer closeInfoLogFile()
	}

	closeErrorLogFiles, err := internal.InitLogging(cfg.Logging.ErrorLog, internal.ErrorLogLevel)
	if err != nil {
		log.Panic(err)
	}
	if closeErrorLogFiles != nil {
		defer closeErrorLogFiles()
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
