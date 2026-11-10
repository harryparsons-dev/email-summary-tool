package requests

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

const (
	// maxEmailLength matches the users.email column.
	maxEmailLength    = 100
	minPasswordLength = 8
	// maxPasswordBytes is bcrypt's limit; it rejects anything longer.
	maxPasswordBytes = 72
)

type SignUpRequest struct {
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

// Validate returns an error whose message is safe to show to the user.
func (r SignUpRequest) Validate() error {
	if strings.TrimSpace(r.Email) == "" || r.Password == "" {
		return errors.New("Email and password are required")
	}

	address, err := mail.ParseAddress(r.Email)
	if err != nil || address.Address != r.Email || len(r.Email) > maxEmailLength {
		return errors.New("Enter a valid email address")
	}

	if utf8.RuneCountInString(r.Password) < minPasswordLength {
		return errors.New("Password must be at least 8 characters")
	}
	if len(r.Password) > maxPasswordBytes {
		return errors.New("Password must be at most 72 bytes")
	}

	return nil
}

type LoginRequest struct {
	Email    string `json:"email" form:"email" validate:"required"`
	Password string `json:"password" form:"password" validate:"required"`
}
