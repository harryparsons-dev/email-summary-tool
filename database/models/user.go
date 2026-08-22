package models

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/schema"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	ID           string    `bun:"id,pk" json:"id"`
	Email        string    `bun:"email,notnull" json:"email"`
	PasswordHash string    `bun:"password_hash,notnull" json:"password_hash"`
	CreatedAt    time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt    time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeAppendModel assigns the application's UUIDv7 primary key before Bun
// builds an INSERT query.
func (u *User) BeforeAppendModel(_ context.Context, query schema.Query) error {
	if _, ok := query.(*bun.InsertQuery); !ok || u.ID != "" {
		return nil
	}

	id, err := uuid.NewV7()
	if err != nil {
		return fmt.Errorf("generate user ID: %w", err)
	}
	u.ID = id.String()
	return nil
}
