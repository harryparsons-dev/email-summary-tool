package projecthandlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"time"

	"email-summary-tool/database/models"
	"email-summary-tool/pkg"
	"email-summary-tool/server"
	"email-summary-tool/server/requests"
	"email-summary-tool/server/responses"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
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
	h.Server.Db.NewSelect().
		Model(&projects).
		Where("user_id = ?", user.ID).
		Apply(pagniation).
		Scan(ctx)

	totalCount, err := h.Server.Db.NewSelect().
		Model(&models.Project{}).
		Where("user_id = ?", user.ID).
		Count(ctx)
	if err != nil {
		totalCount = 0
	}

	response := responses.NewPaginatedProjectResponse(projects, page, pageSize, int(totalCount))
	return c.JSON(http.StatusOK, response)
}

func (h *ProjectHandler) Get(c *echo.Context) error {
	projectID := c.Param("id")
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusNotFound, "Project not found")
	}

	user := c.Get("user").(*models.User)
	project := &models.Project{}
	err := h.Server.Db.NewSelect().
		Model(project).
		Where("id = ?", projectID).
		Where("user_id = ?", user.ID).
		Scan(c.Request().Context())
	if errors.Is(err, sql.ErrNoRows) {
		return c.JSON(http.StatusNotFound, "Project not found")
	}
	if err != nil {
		log.Printf("Unexpected error getting project: %v", err)
		return c.JSON(http.StatusInternalServerError, "Failed to get project")
	}

	return c.JSON(http.StatusOK, responses.NewProjectResponse(project))
}

func (h *ProjectHandler) Create(c *echo.Context) error {
	ctx := c.Request().Context()
	user := c.Get("user").(*models.User)

	request := &requests.CreateProjectRequest{}
	if err := c.Bind(&request); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	if !models.ValidProjectStatuses[request.Status] {
		return c.JSON(http.StatusBadRequest, "Invalid project status")
	}

	project := &models.Project{
		Title:        request.Name,
		Description:  request.Description,
		Status:       request.Status,
		EmailAddress: "",
		UserId:       user.ID,
	}

	_, err := h.Server.Db.NewInsert().Model(project).Exec(ctx)
	if err != nil {
		log.Printf("Unexpected error creating project: %v", err)
		return c.JSON(http.StatusInternalServerError, "Failed to create project")
	}

	response := responses.NewProjectResponse(project)
	return c.JSON(http.StatusCreated, response)
}

func (h *ProjectHandler) Update(c *echo.Context) error {
	request := &requests.UpdateProjectRequest{}
	if err := c.Bind(request); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	if !models.ValidProjectStatuses[request.Status] {
		return c.JSON(http.StatusBadRequest, "Invalid project status")
	}
	if err := request.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, "Invalid request body")
	}

	projectID := c.Param("id")
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusNotFound, "Project not found")
	}

	user := c.Get("user").(*models.User)
	project := &models.Project{}
	err := h.Server.Db.NewSelect().
		Model(project).
		Where("id = ?", projectID).
		Where("user_id = ?", user.ID).
		Scan(c.Request().Context())
	if errors.Is(err, sql.ErrNoRows) {
		return c.JSON(http.StatusNotFound, "Project not found")
	}
	if err != nil {
		log.Printf("Unexpected error getting project for update: %v", err)
		return c.JSON(http.StatusInternalServerError, "Failed to update project")
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
		Exec(c.Request().Context())
	if err != nil {
		log.Printf("Unexpected error updating project: %v", err)
		return c.JSON(http.StatusInternalServerError, "Failed to update project")
	}

	return c.JSON(http.StatusOK, responses.NewProjectResponse(project))
}

func (h *ProjectHandler) Delete(c *echo.Context) error {
	projectID := c.Param("id")
	if _, err := uuid.Parse(projectID); err != nil {
		return c.JSON(http.StatusNotFound, "Project not found")
	}

	user := c.Get("user").(*models.User)
	project := &models.Project{}
	err := h.Server.Db.NewSelect().
		Model(project).
		Where("id = ?", projectID).
		Where("user_id = ?", user.ID).
		Scan(c.Request().Context())
	if errors.Is(err, sql.ErrNoRows) {
		return c.JSON(http.StatusNotFound, "Project not found")
	}
	if err != nil {
		log.Printf("Unexpected error getting project for deletion: %v", err)
		return c.JSON(http.StatusInternalServerError, "Failed to delete project")
	}

	_, err = h.Server.Db.NewDelete().
		Model(project).
		WherePK().
		Where("user_id = ?", user.ID).
		Exec(c.Request().Context())
	if err != nil {
		log.Printf("Unexpected error deleting project: %v", err)
		return c.JSON(http.StatusInternalServerError, "Failed to delete project")
	}

	return c.NoContent(http.StatusNoContent)
}
