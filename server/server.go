package server

import (
	"email-summary-tool/config"
	appvalidator "email-summary-tool/validator"
	"log"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"
)

type Server struct {
	E   *echo.Echo
	Db  *bun.DB
	JWT *config.JWTConfig

	// InboundDomain is the domain used for generated project email addresses.
	InboundDomain string
}

func NewServer(jwtConfig *config.JWTConfig, db *bun.DB, inboundDomain string) *Server {
	e := echo.New()
	e.Validator = appvalidator.New()

	return &Server{
		E:   e,
		Db:  db,
		JWT: jwtConfig,

		InboundDomain: inboundDomain,
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
