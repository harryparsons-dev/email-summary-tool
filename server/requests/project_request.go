package requests

import "email-summary-tool/database/models"

type CreateProjectRequest struct {
	Name        string               `json:"name" validate:"required"`
	Description string               `json:"description" validate:"required"`
	Status      models.ProjectStatus `json:"status" validate:"required"`
}
