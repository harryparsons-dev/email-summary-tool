package projecthandlers

import (
	"email-summary-tool/database/models"
	"email-summary-tool/pkg"
	"email-summary-tool/server"
	"email-summary-tool/server/requests"
	"email-summary-tool/server/responses"
	"log"
	"net/http"

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
