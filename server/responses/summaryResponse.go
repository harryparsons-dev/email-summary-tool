package responses

import (
	"time"

	"email-summary-tool/database/models"
)

type SummaryResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	// EmailID is null when the summary covers the whole project.
	EmailID *string `json:"email_id"`
	// Status is pending, completed or failed. Poll until it is not pending.
	Status    string    `json:"status"`
	Summary   string    `json:"summary"`
	Error     string    `json:"error"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func NewSummaryResponse(summary *models.Summary) SummaryResponse {
	return SummaryResponse{
		ID:        summary.ID,
		ProjectID: summary.ProjectId,
		EmailID:   summary.EmailId,
		Status:    string(summary.Status),
		Summary:   summary.Summary,
		Error:     summary.Error,
		CreatedAt: summary.CreatedAt,
		UpdatedAt: summary.UpdatedAt,
	}
}

func NewSummaryResponses(summaries []models.Summary) []SummaryResponse {
	responses := make([]SummaryResponse, len(summaries))
	for i := range summaries {
		responses[i] = NewSummaryResponse(&summaries[i])
	}
	return responses
}
