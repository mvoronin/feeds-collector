package persistence

import (
	"database/sql"
	"embed"
	"fmt"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

type SQLiteProvider struct {
	DB *sql.DB
}

func NewSQLiteProvider(db *sql.DB) *SQLiteProvider {
	return &SQLiteProvider{DB: db}
}

func Migrate(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("could not set dialect: %w", err)
	}
	pathMigrations := "migrations"
	err := goose.Up(db, pathMigrations)
	if err != nil {
		return fmt.Errorf("could not run migrations: %w", err)
	}
	return nil
}
