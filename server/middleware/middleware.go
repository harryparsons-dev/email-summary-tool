package middleware

import (
	"email-summary-tool/database/models"
	"email-summary-tool/pkg/responses"
	"email-summary-tool/server"
	"net/http"

	"github.com/labstack/echo/v5"
)

func RequireAuth(server *server.Server) echo.MiddlewareFunc {
	return echo.MiddlewareFunc(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {

			cookie, err := c.Request().Cookie("Authorization")
			if err != nil {
				return responses.Error(c, http.StatusUnauthorized, "Authentication required")
			}

			userID, err := server.JWT.VerifyAccessToken(cookie.Value)
			if err != nil {
				return responses.Error(c, http.StatusUnauthorized, "Invalid or expired authentication cookie")
			}

			// get the user
			user := &models.User{}
			server.Db.NewSelect().Model(user).Where("id = ?", userID).Scan(c.Request().Context())
			if user.ID == "" {
				return responses.Error(c, http.StatusUnauthorized, "User not found")
			}

			c.Set("user", user)

			return next(c)
		}
	})
}
