package projecthandlers

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"email-summary-tool/database/models"
	"email-summary-tool/pkg/responses"
	"email-summary-tool/server"
	projectresponses "email-summary-tool/server/responses"
	summarymanager "email-summary-tool/services/summaryManager"

	"github.com/google/uuid"
	"github.com/labstack/echo/v5"
)

const emailSummaryListLimit = 50

// ProjectEmailSummaryHandler serves the
// /projects/:id/emails/:emailId/summarise resource.
type ProjectEmailSummaryHandler struct {
	Server    *server.Server
	Summaries *summarymanager.SummaryManager
}

func NewProjectEmailSummaryHandler(s *server.Server) *ProjectEmailSummaryHandler {
	return &ProjectEmailSummaryHandler{
		Server:    s,
		Summaries: summarymanager.NewSummaryManager(s),
	}
}

// List returns the email's most recent summaries, newest first.
func (h *ProjectEmailSummaryHandler) List(c *echo.Context) error {
	user := c.Get("user").(*models.User)
	ctx := c.Request().Context()

	projectID := c.Param("id")
	if _, err := uuid.Parse(projectID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}

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
		log.Printf("Error loading project for email summaries: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to list summaries")
	}

	emailID := c.Param("emailId")
	if _, err := uuid.Parse(emailID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Email not found")
	}

	summaries := []models.Summary{}
	err = h.Server.Db.NewSelect().
		Model(&summaries).
		Where("s.project_id = ?", project.ID).
		Where("s.email_id = ?", emailID).
		Where("s.user_id = ?", user.ID).
		OrderExpr("s.created_at DESC").
		Limit(emailSummaryListLimit).
		Scan(ctx)
	if err != nil {
		log.Printf("Error listing email summaries: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to list summaries")
	}

	return responses.Data(c, http.StatusOK, projectresponses.NewSummaryResponses(summaries))
}

// Get returns one summary. Clients poll this until status is no longer
// pending.
func (h *ProjectEmailSummaryHandler) Get(c *echo.Context) error {
	user := c.Get("user").(*models.User)
	ctx := c.Request().Context()

	projectID := c.Param("id")
	if _, err := uuid.Parse(projectID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}

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
		log.Printf("Error loading project for email summary: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to get summary")
	}

	emailID := c.Param("emailId")
	summaryID := c.Param("summaryId")
	if _, err := uuid.Parse(emailID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Summary not found")
	}
	if _, err := uuid.Parse(summaryID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Summary not found")
	}

	summary := &models.Summary{}
	err = h.Server.Db.NewSelect().
		Model(summary).
		Where("s.id = ?", summaryID).
		Where("s.project_id = ?", project.ID).
		Where("s.email_id = ?", emailID).
		Where("s.user_id = ?", user.ID).
		Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return responses.Error(c, http.StatusNotFound, "Summary not found")
	}
	if err != nil {
		log.Printf("Error getting email summary: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to get summary")
	}

	return responses.Data(c, http.StatusOK, projectresponses.NewSummaryResponse(summary))
}

// Create starts summarising the email. The summary is generated in the
// background; the response is the pending summary, which the client polls
// with Get. If the email already has a summary in progress, that one is
// returned instead.
func (h *ProjectEmailSummaryHandler) Create(c *echo.Context) error {
	user := c.Get("user").(*models.User)
	ctx := c.Request().Context()

	projectID := c.Param("id")
	if _, err := uuid.Parse(projectID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Project not found")
	}

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
		log.Printf("Error loading project for email summary: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to create summary")
	}

	emailID := c.Param("emailId")
	if _, err := uuid.Parse(emailID); err != nil {
		return responses.Error(c, http.StatusNotFound, "Email not found")
	}

	exists, err := h.Server.Db.NewSelect().
		Model((*models.Email)(nil)).
		Where("id = ?", emailID).
		Where("project_id = ?", project.ID).
		Exists(ctx)
	if err != nil {
		log.Printf("Error checking email for summary: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to create summary")
	}
	if !exists {
		return responses.Error(c, http.StatusNotFound, "Email not found")
	}

	emailSummary, created, err := h.Summaries.Create(ctx, user.ID, project, &emailID)
	if err != nil {
		log.Printf("Error starting email summary: %v", err)
		return responses.Error(c, http.StatusInternalServerError, "Failed to create summary")
	}
	if !created {
		return responses.Data(c, http.StatusOK, projectresponses.NewSummaryResponse(emailSummary))
	}

	h.Server.Dependencies.WgAdd()
	go func() {
		defer h.Server.Dependencies.WgDone()
		h.Summaries.Generate(emailSummary.ID, project, &emailID)
	}()

	return responses.Data(c, http.StatusAccepted, projectresponses.NewSummaryResponse(emailSummary))
}
