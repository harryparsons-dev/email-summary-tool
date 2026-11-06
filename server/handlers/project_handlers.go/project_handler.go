package projecthandlers

import (
	"database/sql"
	"email-summary-tool/database/models"
	"email-summary-tool/pkg/responses"
	"email-summary-tool/server"
	"email-summary-tool/services/inbound"
	"errors"
	"log"
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun/driver/pgdriver"
)

type ProjectHandler struct {
	Server *server.Server
}

func NewProjectHandler(s *server.Server) *ProjectHandler {
	return &ProjectHandler{
		Server: s,
	}
}

// Create makes a project with a newly generated inbound email address.
func (h *ProjectHandler) Create(c *echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return responses.Error(c, http.StatusInternalServerError, "Authenticated user is unavailable")
	}

	var request struct {
		Title       string `json:"title" form:"title" validate:"required,max=255"`
		Description string `json:"description" form:"description"`
	}
	if err := c.Bind(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Invalid request payload")
	}
	if err := c.Validate(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Title is required")
	}

	ctx := c.Request().Context()
	for attempt := 0; attempt < models.ProjectEmailAddressAttempts; attempt++ {
		address, err := inbound.GenerateAddress(h.Server.InboundDomain)
		if err != nil {
			log.Println(err)
			return responses.Error(c, http.StatusInternalServerError, "Error creating project")
		}

		project := &models.Project{
			UserID:       user.ID,
			Title:        request.Title,
			Description:  request.Description,
			EmailAddress: address,
			Status:       models.ProjectStatusActive,
		}
		_, err = h.Server.DB.NewInsert().Model(project).Returning("*").Exec(ctx)
		if isUniqueViolation(err) {
			continue
		}
		if err != nil {
			log.Printf("Error creating project: %v", err)
			return responses.Error(c, http.StatusInternalServerError, "Error creating project")
		}
		return responses.Data(c, http.StatusCreated, project)
	}

	log.Printf("Error creating project: no free email address after %d attempts", models.ProjectEmailAddressAttempts)
	return responses.Error(c, http.StatusInternalServerError, "Error creating project")
}

func (h *ProjectHandler) List(c *echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return responses.Error(c, http.StatusInternalServerError, "Authenticated user is unavailable")
	}

	projects := []models.Project{}
	err := h.Server.DB.NewSelect().
		Model(&projects).
		Where("user_id = ?", user.ID).
		Where("deleted_at IS NULL").
		Order("created_at DESC").
		Scan(c.Request().Context())
	if err != nil {
		log.Printf("Error listing projects: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Error listing projects")
	}
	return responses.Data(c, http.StatusOK, projects)
}

func (h *ProjectHandler) Get(c *echo.Context) error {
	user, ok := c.Get("user").(*models.User)
	if !ok {
		return responses.Error(c, http.StatusInternalServerError, "Authenticated user is unavailable")
	}

	project := &models.Project{}
	err := h.Server.DB.NewSelect().
		Model(project).
		Where("id = ?", c.Param("projectId")).
		Where("user_id = ?", user.ID).
		Where("deleted_at IS NULL").
		Scan(c.Request().Context())
	if errors.Is(err, sql.ErrNoRows) {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}
	if err != nil {
		log.Printf("Error loading project: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Error loading project")
	}
	return responses.Data(c, http.StatusOK, project)
}

func isUniqueViolation(err error) bool {
	var pgErr pgdriver.Error
	return errors.As(err, &pgErr) && pgErr.Field('C') == "23505"
}
