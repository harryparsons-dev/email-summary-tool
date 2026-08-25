package main

import (
	"context"
	"database/sql"
	"email-summary-tool/config"
	"email-summary-tool/database/migrations"
	"email-summary-tool/routes"
	"email-summary-tool/server"
	"log"
	"net/url"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func main() {
	jwtConfig, err := config.NewJWTConfig(os.Getenv("SECRET_KEY"))
	if err != nil {
		log.Fatal(err)
	}

	postgresUser := os.Getenv("POSTGRES_USER")
	if postgresUser == "" {
		log.Fatal("POSTGRES_USER is not set")
	}

	postgresPassword := os.Getenv("POSTGRES_PASSWORD")
	if postgresPassword == "" {
		log.Fatal("POSTGRES_PASSWORD is not set")
	}

	postgresDB := os.Getenv("POSTGRES_DB")
	if postgresDB == "" {
		log.Fatal("POSTGRES_DB is not set")
	}

	databaseURL := (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(postgresUser, postgresPassword),
		Host:     "postgres:5432",
		Path:     postgresDB,
		RawQuery: "sslmode=disable",
	}).String()

	sqlDB := sql.OpenDB(pgdriver.NewConnector(pgdriver.WithDSN(databaseURL)))
	db := bun.NewDB(sqlDB, pgdialect.New())

	if err := db.PingContext(context.Background()); err != nil {
		_ = db.Close()
		log.Fatalf("Unable to connect to database: %v", err)
	}

	if err := migrations.Up(db); err != nil {
		_ = db.Close()
		log.Fatalf("Unable to apply database migrations: %v", err)
	}

	s := server.NewServer(jwtConfig, db)
	defer s.Close()

	routes.InitializeRoutes(s)

	if err := s.Start(":8080"); err != nil {
		log.Fatal(err)
	}
}
