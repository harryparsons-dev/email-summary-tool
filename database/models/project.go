package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

const ProjectStatusActive = "active"

// ProjectEmailAddressAttempts bounds retries when a generated address is
// already taken, which with 60 random bits should practically never happen.
const ProjectEmailAddressAttempts = 5

type Project struct {
	bun.BaseModel `bun:"table:projects,alias:p"`

	ID           string     `bun:"id,pk" json:"id"`
	UserID       string     `bun:"user_id,notnull" json:"user_id"`
	Title        string     `bun:"title,notnull" json:"title"`
	Description  string     `bun:"description" json:"description"`
	EmailAddress string     `bun:"email_address,notnull" json:"email_address"`
	Status       string     `bun:"status,notnull" json:"status"`
	CreatedAt    time.Time  `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time  `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt    *time.Time `bun:"deleted_at" json:"-"`
}

func (p *Project) BeforeAppendModel(_ context.Context, query schema.Query) error {
	if _, ok := query.(*bun.InsertQuery); !ok || p.ID != "" {
		return nil
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate project ID: %w", err)
	}
	p.ID = id.String()
	return nil
}
