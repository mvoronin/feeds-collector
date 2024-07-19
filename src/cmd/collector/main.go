package main

import (
	"context"
	"database/sql"
	"errors"
	"feedscollector/internal"
	"feedscollector/internal/gatherer"
	"feedscollector/internal/server"
	"feedscollector/pkg/utils"
	"flag"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func runMigrations(db *sql.DB) error {
	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		return err
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"sqlite3", driver)
	if err != nil {
		return err
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}

	return nil
}

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	config, err := utils.ReadConfig(*configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	log.Printf("Reading a config file: %s\n", *configPath)

	closeInfoLogFile, err := internal.InitLogging(config.Logging.InfoLog, internal.InfoLogLevel)
	if err != nil {
		log.Fatal(err)
	}
	if closeInfoLogFile != nil {
		defer closeInfoLogFile()
	}

	closeErrorLogFiles, err := internal.InitLogging(config.Logging.ErrorLog, internal.ErrorLogLevel)
	if err != nil {
		log.Panic(err)
	}
	if closeErrorLogFiles != nil {
		defer closeErrorLogFiles()
	}

	db, err := sql.Open("sqlite3", config.Database.Path)
	if err != nil {
		internal.ErrorLogger.Fatalf("Error opening database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			internal.ErrorLogger.Fatalf("Error closing database: %v", err)
		}
	}(db)

	var wg sync.WaitGroup // нужна ли мне эта WaitGroup?
	ctx := context.Background()
	ctxWithCancel, cancel := context.WithCancel(ctx)

	wg.Add(1)
	go server.RunAPIServer(ctxWithCancel, db, config, &wg)

	wg.Add(1)
	go gatherer.RunGathererLoop(ctxWithCancel, db, config, &wg)

	// Handle graceful shutdown on Ctrl+C
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh

	cancel()
	wg.Wait()
}
