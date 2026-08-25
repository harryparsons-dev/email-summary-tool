package server

import (
	"email-summary-tool/config"
	"log"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"
)

type Server struct {
	E   *echo.Echo
	Db  *bun.DB
	JWT *config.JWTConfig
}

func NewServer(jwtConfig *config.JWTConfig, db *bun.DB) *Server {
	return &Server{
		E:   echo.New(),
		Db:  db,
		JWT: jwtConfig,
	}
}

func (s *Server) Start(address string) error {
	return s.E.Start(address)
}

func (s *Server) Close() {
	if err := s.Db.Close(); err != nil {
		log.Printf("Unable to close Bun database: %v", err)
	}
}
