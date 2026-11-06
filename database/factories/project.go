package factories

import (
	"context"
	"fmt"

	"email-summary-tool/database/models"
	"email-summary-tool/services/inbound"

	"github.com/uptrace/bun"
)

func NewProject(ctx context.Context, db *bun.DB, userID string) (*models.Project, error) {
	// Project email addresses are unique, so each fixture needs its own.
	emailAddress, err := inbound.GenerateAddress("example.com")
	if err != nil {
		return nil, fmt.Errorf("generate project fixture email address: %w", err)
	}

	project := &models.Project{
		Title:        "Test project",
		Description:  "A project created for testing",
		Status:       models.ProjectStatus("pending"),
		EmailAddress: emailAddress,
		UserId:       userID,
	}

	if _, err := db.NewInsert().Model(project).Exec(ctx); err != nil {
		return nil, fmt.Errorf("insert project fixture: %w", err)
	}

	return project, nil
}
