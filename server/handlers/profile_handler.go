package handlers

import (
	"email-summary-tool/database/models"
	"email-summary-tool/pkg/responses"
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
		return responses.Error(c, http.StatusInternalServerError, "Authenticated user is unavailable")
	}

	return responses.Data(c, http.StatusOK, profileResponse{
		ID:    user.ID,
		Email: user.Email,
	})
}
