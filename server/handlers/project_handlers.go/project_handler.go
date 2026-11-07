package projecthandlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"email-summary-tool/database/models"
	"email-summary-tool/pkg"
	"email-summary-tool/pkg/responses"
	"email-summary-tool/server"
	"email-summary-tool/server/requests"
	projectresponses "email-summary-tool/server/responses"
	"email-summary-tool/services/inbound"

	"github.com/google/uuid"
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

func (h *ProjectHandler) List(c *echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get("user").(*models.User)

	page, pageSize, pagniation := pkg.Paginate(c)

	var projects []models.Project
	err := h.Server.Db.NewSelect().
		Model(&projects).
		Where("user_id = ?", user.ID).
		Apply(pagniation).
		Scan(ctx)
	if err != nil {
		log.Printf("Unexpected error listing projects: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to list projects")
	}

	totalCount, err := h.Server.Db.NewSelect().
		Model(&models.Project{}).
		Where("user_id = ?", user.ID).
		Count(ctx)
	if err != nil {
		totalCount = 0
	}

	response := projectresponses.NewPaginatedProjectResponse(projects, page, pageSize, int(totalCount))
	return responses.Data(c, http.StatusOK, response)
}

func (h *ProjectHandler) Get(c *echo.Context) error {
	ctx := c.Request().Context()
	projectID := c.Param("id")
	if _, err := uuid.Parse(projectID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}

	user := c.Get("user").(*models.User)
	project := &models.Project{}
	err := h.Server.Db.NewSelect().
		Model(project).
		Where("id = ?", projectID).
		Where("user_id = ?", user.ID).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}
	if err != nil {
		log.Printf("Unexpected error getting project: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to get project")
	}

	return responses.Data(c, http.StatusOK, projectresponses.NewProjectResponse(project))
}

func (h *ProjectHandler) Create(c *echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get("user").(*models.User)

	request := &requests.CreateProjectRequest{}
	if err := c.Bind(&request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Invalid request body")
	}

	if !models.ValidProjectStatuses[request.Status] {
		return responses.Error(c, http.StatusBadRequest, "Invalid project status")
	}

	// Each project gets a newly generated inbound email address; retry on the
	// unlikely chance that the generated address is already taken.
	for attempt := 0; attempt < models.ProjectEmailAddressAttempts; attempt++ {
		address, err := inbound.GenerateAddress(h.Server.InboundDomain)
		if err != nil {
			log.Printf("Unexpected error generating project email address: %v", err)
			return responses.Error(c, http.StatusInternalServerError, "Failed to create project")
		}

		project := &models.Project{
			Title:        request.Name,
			Description:  request.Description,
			Status:       request.Status,
			EmailAddress: address,
			UserId:       user.ID,
		}

		_, err = h.Server.Db.NewInsert().Model(project).Exec(ctx)
		if isUniqueViolation(err) {
			continue
		}
		if err != nil {
			log.Printf("Unexpected error creating project: %v", err)
			return responses.Error(c, http.StatusInternalServerError, "Failed to create project")
		}

		response := projectresponses.NewProjectResponse(project)
		return responses.Data(c, http.StatusCreated, response)
	}

	log.Printf("Unexpected error creating project: no free email address after %d attempts", models.ProjectEmailAddressAttempts)
	return responses.Error(c, http.StatusInternalServerError, "Failed to create project")
}

func (h *ProjectHandler) Update(c *echo.Context) error {
	ctx := c.Request().Context()
	request := &requests.UpdateProjectRequest{}
	if err := c.Bind(request); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Invalid request body")
	}

	if !models.ValidProjectStatuses[request.Status] {
		return responses.Error(c, http.StatusBadRequest, "Invalid project status")
	}
	if err := request.Validate(); err != nil {
		return responses.Error(c, http.StatusBadRequest, "Invalid request body")
	}

	projectID := c.Param("id")
	if _, err := uuid.Parse(projectID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}

	user := c.Get("user").(*models.User)
	project := &models.Project{}
	err := h.Server.Db.NewSelect().
		Model(project).
		Where("id = ?", projectID).
		Where("user_id = ?", user.ID).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}
	if err != nil {
		log.Printf("Unexpected error getting project for update: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to update project")
	}

	project.Title = request.Name
	project.Description = request.Description
	project.Status = request.Status
	project.UpdatedAt = time.Now()

	_, err = h.Server.Db.NewUpdate().
		Model(project).
		Column("title", "description", "status", "updated_at").
		WherePK().
		Where("user_id = ?", user.ID).
		Exec(ctx)
	if err != nil {
		log.Printf("Unexpected error updating project: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to update project")
	}

	return responses.Data(c, http.StatusOK, projectresponses.NewProjectResponse(project))
}

func (h *ProjectHandler) Delete(c *echo.Context) error {
	ctx := c.Request().Context()
	projectID := c.Param("id")
	if _, err := uuid.Parse(projectID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}

	user := c.Get("user").(*models.User)
	project := &models.Project{}
	err := h.Server.Db.NewSelect().
		Model(project).
		Where("id = ?", projectID).
		Where("user_id = ?", user.ID).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}
	if err != nil {
		log.Printf("Unexpected error getting project for deletion: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to delete project")
	}

	// Project has a soft_delete column, so this sets deleted_at rather than
	// removing the row; the project and its emails are kept.
	_, err = h.Server.Db.NewDelete().
		Model(project).
		WherePK().
		Where("user_id = ?", user.ID).
		Exec(ctx)
	if err != nil {
		log.Printf("Unexpected error deleting project: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to delete project")
	}

	return c.NoContent(http.StatusNoContent)
}

func isUniqueViolation(err error) bool {
	var pgErr pgdriver.Error
	return errors.As(err, &pgErr) && pgErr.Field('C') == "23505"
}
