// Package migrations configures golang-migrate for the application's database.
// Creating a Migrator does not apply any up or down migrations.
package migrations

import (
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	pgxmigrate "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

const SourceURL = "file://database/migrations/sql"

// New creates a migrator backed by the application's existing connection pool.
// The caller owns the returned migrator and must close it when finished.
func New(pool *pgxpool.Pool) (*migrate.Migrate, error) {
	db := stdlib.OpenDBFromPool(pool)
	driver, err := pgxmigrate.WithInstance(db, &pgxmigrate.Config{})
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("create pgx migration driver: %w", err)
	}

	migrator, err := migrate.NewWithDatabaseInstance(SourceURL, "pgx5", driver)
	if err != nil {
		_ = driver.Close()
		return nil, fmt.Errorf("create migrator: %w", err)
	}

	return migrator, nil
}
