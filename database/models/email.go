package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type EmailAttachment struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type Email struct {
	bun.BaseModel `bun:"table:emails,alias:e"`

	ID           string            `bun:"id,pk" json:"id"`
	ProjectID    string            `bun:"project_id,notnull" json:"project_id"`
	EnvelopeFrom string            `bun:"envelope_from,notnull" json:"envelope_from"`
	EnvelopeTo   string            `bun:"envelope_to,notnull" json:"envelope_to"`
	MessageID    string            `bun:"message_id,notnull" json:"message_id"`
	FromAddress  string            `bun:"from_address,notnull" json:"from_address"`
	FromName     string            `bun:"from_name,notnull" json:"from_name"`
	ToAddresses  string            `bun:"to_addresses,notnull" json:"to_addresses"`
	Subject      string            `bun:"subject,notnull" json:"subject"`
	SentAt       *time.Time        `bun:"sent_at" json:"sent_at"`
	TextBody     string            `bun:"text_body,notnull" json:"text_body"`
	HTMLBody     string            `bun:"html_body,notnull" json:"html_body"`
	Attachments  []EmailAttachment `bun:"attachments,type:jsonb,notnull" json:"attachments"`
	RawMessage   []byte            `bun:"raw_message,type:bytea,notnull" json:"-"`
	ReceivedAt   time.Time         `bun:"received_at,nullzero,notnull,default:current_timestamp" json:"received_at"`
}

func (e *Email) BeforeAppendModel(_ context.Context, query schema.Query) error {
	if _, ok := query.(*bun.InsertQuery); !ok || e.ID != "" {
		return nil
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate email ID: %w", err)
	}
	e.ID = id.String()
	return nil
}
