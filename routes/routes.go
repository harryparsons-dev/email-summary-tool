package routes

import (
	"email-summary-tool/auth"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"
)

func InitializeRoutes(e *echo.Echo, pool *pgxpool.Pool, config auth.Config) error {
	e.GET("/", func(c *echo.Context) error {
		return c.String(http.StatusOK, "email summary server is running")
	})

	tokens, err := auth.NewTokenManager(config.JWTSecret, config.JWTIssuer, config.JWTAudience, config.AccessTTL, config.RefreshTTL)
	if err != nil {
		return fmt.Errorf("configure authentication tokens: %w", err)
	}
	mailer, err := auth.MailerFromEnv(config.AppBaseURL)
	if err != nil {
		return fmt.Errorf("configure password reset mailer: %w", err)
	}
	service, err := auth.NewService(auth.NewPostgresStore(pool), tokens, mailer, config.ResetTTL)
	if err != nil {
		return fmt.Errorf("configure authentication service: %w", err)
	}
	auth.NewHandler(service, tokens, config.CookieSecure, config.RefreshTTL).RegisterRoutes(e)
	return nil
}
