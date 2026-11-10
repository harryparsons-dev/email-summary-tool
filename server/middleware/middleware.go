package middleware

import (
	"database/sql"
	"email-summary-tool/database/models"
	"email-summary-tool/pkg/responses"
	"email-summary-tool/server"
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
)

func Authenticated(server *server.Server) echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			ctx := c.Request().Context()

			cookie, err := c.Request().Cookie("Authorization")
			if err != nil {
				return responses.Error(c, http.StatusUnauthorized, "Authentication required")
			}

			userID, err := server.JWT.VerifyAccessToken(cookie.Value)
			if err != nil {
				return responses.Error(c, http.StatusUnauthorized, "Invalid or expired authentication cookie")
			}

			user := &models.User{}
			err = server.Db.NewSelect().Model(user).Where("id = ?", userID).Scan(ctx)
			if errors.Is(err, sql.ErrNoRows) {
				return responses.Error(c, http.StatusUnauthorized, "User not found")
			}
			if err != nil {
				log.Printf("Error looking up authenticated user: %v", err)
				return responses.Error(c, http.StatusInternalServerError, "Error authenticating request")
			}

			c.Set("user", user)

			return next(c)
		}
	})
}
