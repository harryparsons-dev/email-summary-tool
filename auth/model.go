package auth

import (
	"errors"
	"time"
)

var (
	ErrEmailInUse        = errors.New("email address is already registered")
	ErrInvalidCredential = errors.New("invalid email or password")
	ErrInvalidToken      = errors.New("token is invalid or expired")
	ErrInvalidInput      = errors.New("invalid input")
)

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}

type Session struct {
	User         User
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

type StoredRefreshToken struct {
	Hash      []byte
	ID        string
	UserID    string
	ExpiresAt time.Time
}
