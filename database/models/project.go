package models

type ProjectStatus string

const (
	projectStatusPending    ProjectStatus = "pending"
	projectStatusInProgress ProjectStatus = "in_progress"
	projectStatusCompleted  ProjectStatus = "completed"
	projectStatusArchived   ProjectStatus = "archived"
)

var ValidProjectStatuses = map[ProjectStatus]bool{
	projectStatusPending:    true,
	projectStatusInProgress: true,
	projectStatusCompleted:  true,
	projectStatusArchived:   true,
}

type Project struct {
	UuidBase

	Title        string        `json:"title"`
	Description  string        `json:"description"`
	Status       ProjectStatus `json:"status"`
	EmailAddress string        `json:"email_address"`

	UserId string `json:"user_id"`
}
