package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type UuidBase struct {
	ID        string    `bun:",pk" json:"id"`
	CreatedAt time.Time `bun:",nullzero,default:current_timestamp" json:"created_at"`
	UpdatedAt time.Time `bun:",nullzero,default:current_timestamp" json:"updated_at"`
}

// BeforeAppendModel assigns the application's UUIDv7 primary key before Bun
// builds an INSERT query.
func (u *UuidBase) BeforeAppendModel(_ context.Context, query schema.Query) error {
	if _, ok := query.(*bun.InsertQuery); !ok || u.ID != "" {
		return nil
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate model ID: %w", err)
	}
	u.ID = id.String()
	return nil
}
