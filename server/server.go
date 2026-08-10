package server

import (
	"context"
	"email-summary-tool/routes"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

type Server struct {
	e  *echo.Echo
	Db *pgxpool.Pool
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

	routes.InitializeRoutes(e)

	var greeting string
	err = dbpool.QueryRow(context.Background(), "select 'Hello, world!'").Scan(&greeting)
	if err != nil {
		log.Fatalf("QueryRow failed: %v", err)
	}

	log.Printf("Database connection successful: %s", greeting)

	return &Server{e: e, Db: dbpool}
}

func (s *Server) Start(address string) error {
	return s.e.Start(address)
}

func (s *Server) Close() {
	s.Db.Close()
}
