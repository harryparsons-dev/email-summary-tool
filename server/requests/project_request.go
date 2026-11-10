package requests

import (
	"errors"
	"strings"
	"unicode/utf8"

	"email-summary-tool/database/models"
)

// maxProjectNameLength matches the projects.title column.
const maxProjectNameLength = 255

type CreateProjectRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Status      models.ProjectStatus `json:"status"`
}

func (r CreateProjectRequest) Validate() error {
	return validateProjectDetails(r.Name, r.Description)
}

type UpdateProjectRequest struct {
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Status      models.ProjectStatus `json:"status"`
}

func (r UpdateProjectRequest) Validate() error {
	return validateProjectDetails(r.Name, r.Description)
}

func validateProjectDetails(name, description string) error {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(description) == "" {
		return errors.New("name and description are required")
	}
	if utf8.RuneCountInString(name) > maxProjectNameLength {
		return errors.New("name is too long")
	}

	return nil
}
