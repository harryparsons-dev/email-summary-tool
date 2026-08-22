package server

import (
	"context"
	"database/sql"
	"email-summary-tool/config"
	"email-summary-tool/database/migrations"
	appvalidator "email-summary-tool/validator"
	"log"
	"net/url"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

type Server struct {
	E   *echo.Echo
	DB  *bun.DB
	JWT *config.JWTConfig
}

func NewServer(jwtConfig *config.JWTConfig) *Server {
	e := echo.New()
	e.Validator = appvalidator.New()

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

	return &Server{
		E:   e,
		DB:  db,
		JWT: jwtConfig,
	}
}

func (s *Server) Start(address string) error {
	return s.E.Start(address)
}

func (s *Server) Close() {
	if err := s.DB.Close(); err != nil {
		log.Printf("Unable to close Bun database: %v", err)
	}
}
