package server

import (
	"email-summary-tool/pkg/jwt"
	appvalidator "email-summary-tool/pkg/validator"
	"email-summary-tool/services"
	"log"
	"log/slog"
	"os"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/uptrace/bun"
)

type Server struct {
	E   *echo.Echo
	Db  *bun.DB
	JWT *jwt.Config

	// Dependencies holds the other shared dependencies, such as the
	// summarizer and background task tracker.
	Dependencies *services.DependencyService
}

func NewServer(jwtConfig *jwt.Config, db *bun.DB, dependencies *services.DependencyService) *Server {

	e := echo.NewWithConfig(echo.Config{NoGroupAutoRegister404Routes: true})
	e.Validator = appvalidator.New()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:      true,
		LogStatus:   true,
		HandleError: true,
		LogValuesFunc: func(c *echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error != nil {
				logger.Error("request failed", "uri", v.URI, "status", v.Status, "error", v.Error)
				return nil
			}
			logger.Info("request", "uri", v.URI, "status", v.Status)
			return nil
		},
	}))

	e.Use(middleware.Recover())
	e.Use(middleware.BodyLimit(1048576)) // 1 MB

	return &Server{
		E:   e,
		Db:  db,
		JWT: jwtConfig,

		Dependencies: dependencies,
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
