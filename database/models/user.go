package models

import (
	"github.com/uptrace/bun"
)

type User struct {
	bun.BaseModel `bun:"table:users,alias:u"`

	UuidBase

	Email        string `json:"email"`
	PasswordHash string `json:"password_hash"`
}
