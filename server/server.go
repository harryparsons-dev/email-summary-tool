package server

import (
	"context"
	"email-summary-tool/database/migrations"
	"email-summary-tool/routes"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

type Server struct {
	e  *echo.Echo
	DB *pgxpool.Pool
}

func NewServer() *Server {
	e := echo.New()

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	dbpool, err := pgxpool.New(context.Background(), databaseURL)
	if err != nil {
		log.Fatalf("Unable to create connection pool: %v", err)
	}

	if err := dbpool.Ping(context.Background()); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	if err := migrations.Up(dbpool); err != nil {
		dbpool.Close()
		log.Fatalf("Unable to apply database migrations: %v", err)
	}

	routes.InitializeRoutes(e)

	return &Server{e: e, DB: dbpool}
}

func (s *Server) Start(address string) error {
	return s.e.Start(address)
}

func (s *Server) Close() {
	s.DB.Close()
}
