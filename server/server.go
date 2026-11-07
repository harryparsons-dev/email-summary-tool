package server

import (
	"email-summary-tool/config"
	"email-summary-tool/services/background"
	"email-summary-tool/services/summary"
	appvalidator "email-summary-tool/validator"
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
	JWT *config.JWTConfig

	// InboundDomain is the domain used for generated project email addresses.
	InboundDomain string

	// Summarizer writes the AI summaries. Tests replace it with a fake.
	Summarizer summary.Summarizer

	// Background tracks work handlers run in background goroutines.
	Background *background.Tasks
}

func NewServer(jwtConfig *config.JWTConfig, db *bun.DB, inboundDomain string, summarizer summary.Summarizer) *Server {

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

	return &Server{
		E:   e,
		Db:  db,
		JWT: jwtConfig,

		InboundDomain: inboundDomain,

		Summarizer: summarizer,
		Background: background.NewTasks(),
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
