// Package migrations applies the application's database migrations.
package migrations

import (
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const (
	// SourceURL is retained for callers that use the on-disk migration path.
	// The API runner uses the embedded copy so production builds are self-contained.
	SourceURL       = "file://database/migrations/sql"
	sourceName      = "embedded migrations"
	migrationsPath  = "sql"
	migrationsTable = "migrations"
)

// migrationFiles is embedded so migrations are available in the production
// image, which only contains the compiled API binary.
//
//go:embed sql/*.sql
var migrationFiles embed.FS

type migrationLogger struct{}

func (migrationLogger) Printf(format string, args ...any) {
	log.Printf("database migration: "+format, args...)
}

func (migrationLogger) Verbose() bool {
	return false
}

// New creates a migrator backed by the application's existing connection pool.
// The caller owns the returned migrator and must close it when finished.
func New(pool *pgxpool.Pool) (*migrate.Migrate, error) {
	sourceDriver, err := iofs.New(migrationFiles, migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("create migration source: %w", err)
	}

	db := stdlib.OpenDBFromPool(pool)
	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		_ = sourceDriver.Close()
		_ = db.Close()
		return nil, fmt.Errorf("create pgx migration driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance(sourceName, sourceDriver, "pgx5", driver)
	if err != nil {
		_ = sourceDriver.Close()
		_ = driver.Close()
		return nil, fmt.Errorf("create migrator: %w", err)
	}

	return migrator, nil
}

// Up compares the embedded migration versions with the version recorded in the
// migrations table and applies every pending migration in order.
func Up(pool *pgxpool.Pool) error {
	migrator, err := New(pool)
	if err != nil {
		return err
	}
	migrator.Log = migrationLogger{}

	migrationErr := migrator.Up()
	if errors.Is(migrationErr, migrate.ErrNoChange) {
		migrationErr = nil
		logMigrationVersion(migrator, "already up to date")
	} else if migrationErr != nil {
		migrationErr = fmt.Errorf("apply migrations: %w", migrationErr)
	} else {
		logMigrationVersion(migrator, "complete")
	}

	sourceErr, databaseErr := migrator.Close()
	if sourceErr != nil {
		sourceErr = fmt.Errorf("close migration source: %w", sourceErr)
	}
	if databaseErr != nil {
		databaseErr = fmt.Errorf("close migration database connection: %w", databaseErr)
	}

	return errors.Join(migrationErr, sourceErr, databaseErr)
}

func logMigrationVersion(migrator *migrate.Migrate, status string) {
	version, dirty, err := migrator.Version()
	if err != nil {
		log.Printf("database migrations %s", status)
		return
	}

	log.Printf("database migrations %s (version=%d, dirty=%t)", status, version, dirty)
}
