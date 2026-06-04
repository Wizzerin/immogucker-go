package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib" // Driver for database/sql
)

// InitDB establishes a connection to the database and applies migrations
func InitDB(connURL string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connURL)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Verify that the database is reachable
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database is unreachable: %w", err)
	}

	log.Println("Successfully connected to PostgreSQL")

	err = runMigrations(db)
	if err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	return db, nil
}

// runMigrations applies any pending database migrations
func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Build the absolute URL, converting Windows slashes (\ -> /) if necessary
	// Result format: file://E:/golang_tasks/immogucker-go/migrations
	migrationURL := fmt.Sprintf("file://%s/migrations", filepath.ToSlash(wd))
	log.Printf("Migrations path: %s", migrationURL)

	m, err := migrate.NewWithDatabaseInstance(
		"file://./migrations",
		"postgres", driver)
	if err != nil {
		return err
	}

	log.Println("Starting database migrations...")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return err
	}

	log.Println("Migrations applied successfully (or already up to date)")
	return nil
}
