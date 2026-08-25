// Package migrations applies the application's database migrations.
package migrations

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	postgresmigrate "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/uptrace/bun"
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
	if version, name, ok := successfulUpMigration(format, args...); ok {
		log.Printf("database migration %s ran successfully (version=%s)", name, version)
		return
	}

	log.Printf("database migration: "+format, args...)
}

func (migrationLogger) Verbose() bool {
	return false
}

// successfulUpMigration recognizes the completion message emitted by
// golang-migrate after it has run an up migration and marked its version clean.
func successfulUpMigration(format string, args ...any) (version, name string, ok bool) {
	if len(args) == 0 {
		return "", "", false
	}

	description, isString := args[0].(string)
	if !isString {
		return "", "", false
	}

	fields := strings.Fields(description)
	if len(fields) != 2 {
		return "", "", false
	}

	version, direction, found := strings.Cut(fields[0], "/")
	if !found || version == "" || direction != "u" || fields[1] == "" {
		return "", "", false
	}

	// In non-verbose mode the successful completion format is "%v (%v)".
	// Checking it prevents unrelated library messages from being reported as a
	// successful migration if they happen to begin with a migration identifier.
	if strings.TrimSpace(format) != "%v (%v)" {
		return "", "", false
	}

	return version, fields[1], true
}

// New creates a migrator backed by the application's Bun database.
// The caller owns the returned migrator and must close it when finished.
func New(db *bun.DB) (*migrate.Migrate, error) {
	sourceDriver, err := iofs.New(migrationFiles, migrationsPath)
	if err != nil {
		return nil, fmt.Errorf("create migration source: %w", err)
	}

	ctx := context.Background()
	conn, err := db.DB.Conn(ctx)
	if err != nil {
		_ = sourceDriver.Close()
		return nil, fmt.Errorf("get migration connection: %w", err)
	}

	driver, err := postgresmigrate.WithConnection(ctx, conn, &postgresmigrate.Config{
		MigrationsTable: migrationsTable,
	})
	if err != nil {
		_ = sourceDriver.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("create PostgreSQL migration driver: %w", err)
	}

	migrator, err := migrate.NewWithInstance(sourceName, sourceDriver, "postgres", driver)
	if err != nil {
		_ = sourceDriver.Close()
		_ = driver.Close()
		return nil, fmt.Errorf("create migrator: %w", err)
	}

	return migrator, nil
}

// Up compares the embedded migration versions with the version recorded in the
// migrations table and applies every pending migration in order.
func Up(db *bun.DB) error {
	migrator, err := New(db)
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
