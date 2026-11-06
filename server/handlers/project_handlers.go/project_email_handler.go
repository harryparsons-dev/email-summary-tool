package projecthandlers

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

// ProjectEmailHandler serves the /projects/:projectId/emails resource.
type ProjectEmailHandler struct {
	Server *server.Server
}

func NewProjectEmailHandler(s *server.Server) *ProjectEmailHandler {
	return &ProjectEmailHandler{
		Server: s,
	}
}

// List returns the emails received by one of the user's projects, newest
// first.
func (h *ProjectEmailHandler) List(c *echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return responses.Error(c, http.StatusInternalServerError, "Authenticated user is unavailable")
	}

	ctx := c.Request().Context()
	project := &models.Project{}
	err := h.Server.DB.NewSelect().
		Model(project).
		Where("id = ?", c.Param("projectId")).
		Where("user_id = ?", user.ID).
		Where("deleted_at IS NULL").
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}
	if err != nil {
		log.Printf("Error loading project: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Error listing emails")
	}

	emails := []models.Email{}
	err = h.Server.DB.NewSelect().
		Model(&emails).
		ExcludeColumn("raw_message").
		Where("project_id = ?", project.ID).
		Order("received_at DESC").
		Limit(100).
		Scan(ctx)
	if err != nil {
		log.Printf("Error listing emails: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Error listing emails")
	}
	return responses.Data(c, http.StatusOK, emails)
}
