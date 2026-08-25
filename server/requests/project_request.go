package requests

import (
	"errors"
	"strings"

	"email-summary-tool/database/models"
)

type CreateProjectRequest struct {
	Name        string               `json:"name" validate:"required"`
	Description string               `json:"description" validate:"required"`
	Status      models.ProjectStatus `json:"status" validate:"required"`
}

type UpdateProjectRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Status      models.ProjectStatus `json:"status"`
}

func (r UpdateProjectRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" || strings.TrimSpace(r.Description) == "" {
		return errors.New("name and description are required")
	}

	return nil
}
