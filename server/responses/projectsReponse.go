package responses

import "email-summary-tool/database/models"

type ProjectResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

type PaginatedProjectResponse struct {
	Projects []ProjectResponse `json:"projects"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
	Total    int               `json:"total"`
}

func NewProjectResponse(project *models.Project) ProjectResponse {
	return ProjectResponse{
		ID:          project.ID,
		Name:        project.Title,
		Description: project.Description,
		Status:      string(project.Status),
	}
}

func NewProjectResponses(projects []models.Project) []ProjectResponse {
	responses := make([]ProjectResponse, len(projects))
	for i, project := range projects {
		responses[i] = NewProjectResponse(&project)
	}

	return responses
}

func NewPaginatedProjectResponse(projects []models.Project, page, pageSize, total int) PaginatedProjectResponse {
	return PaginatedProjectResponse{
		Projects: NewProjectResponses(projects),
		Page:     page,
		PageSize: pageSize,
		Total:    total,
	}
}
