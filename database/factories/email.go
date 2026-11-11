package factories

import (
	"context"
	"fmt"

	"mailai/database/models"

	"github.com/uptrace/bun"
)

func NewEmail(ctx context.Context, db *bun.DB, projectID, subject string) (*models.Email, error) {
	email := &models.Email{
		ProjectID:   projectID,
		EnvelopeTo:  "project@example.com",
		FromAddress: "sender@example.com",
		FromName:    "Test Sender",
		ToAddresses: "project@example.com",
		Subject:     subject,
		TextBody:    "This is a test email.",
		Attachments: []models.EmailAttachment{},
		RawMessage:  []byte("Subject: " + subject + "\r\n\r\nThis is a test email."),
	}

	if _, err := db.NewInsert().Model(email).Exec(ctx); err != nil {
		return nil, fmt.Errorf("insert email fixture: %w", err)
	}

	return email, nil
}
