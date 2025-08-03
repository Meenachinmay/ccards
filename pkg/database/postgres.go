package database

import (
	"ccards/pkg/config"
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
)

var embedMigrations embed.FS

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(cfg config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db, nil
}

func RunMigrations(db *sql.DB, migrationsPath string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, migrationsPath); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return fmt.Errorf("failed to get migration version: %w", err)
	}

	log.Printf("Successfully migrated database to version: %d", version)
	return nil
}

func MigrateDown(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Down(db, "migrations"); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

func GetMigrationStatus(db *sql.DB) (int64, error) {
	return goose.GetDBVersion(db)
}
