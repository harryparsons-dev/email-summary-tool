package factories

import (
	"context"
	"fmt"

	"email-summary-tool/database/models"

	"github.com/uptrace/bun"
)

func NewProject(ctx context.Context, db *bun.DB, userID string) (*models.Project, error) {
	project := &models.Project{
		Title:        "Test project",
		Description:  "A project created for testing",
		Status:       models.ProjectStatus("pending"),
		EmailAddress: "project@example.com",
		UserId:       userID,
	}

	if _, err := db.NewInsert().Model(project).Exec(ctx); err != nil {
		return nil, fmt.Errorf("insert project fixture: %w", err)
	}

	return project, nil
}
