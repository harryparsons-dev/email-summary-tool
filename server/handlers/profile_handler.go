package handlers

import (
	"email-summary-tool/database/models"
	"net/http"

	"github.com/labstack/echo/v5"
)

type profileResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

func GetProfile(c *echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Authenticated user is unavailable",
		})
	}

	return c.JSON(http.StatusOK, profileResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}
