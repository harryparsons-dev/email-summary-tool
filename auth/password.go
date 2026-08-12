package auth

import (
	"fmt"
	"net/mail"
	"strings"
	"unicode/utf8"

	"golang.org/x/crypto/bcrypt"
)

const MinPasswordLength = 12

func NormalizeEmail(value string) (string, error) {
	email := strings.ToLower(strings.TrimSpace(value))
	if len(email) == 0 || len(email) > 254 {
		return "", fmt.Errorf("%w: enter a valid email address", ErrInvalidInput)
	}
	parsed, err := mail.ParseAddress(email)
	if err != nil || parsed.Address != email {
		return "", fmt.Errorf("%w: enter a valid email address", ErrInvalidInput)
	}
	return email, nil
}

func ValidatePassword(password string) error {
	if utf8.RuneCountInString(password) < MinPasswordLength {
		return fmt.Errorf("%w: password must be at least %d characters", ErrInvalidInput, MinPasswordLength)
	}
	if len([]byte(password)) > 72 {
		return fmt.Errorf("%w: password must be at most 72 bytes", ErrInvalidInput)
	}
	return nil
}

func HashPassword(password string) (string, error) {
	if err := ValidatePassword(password); err != nil {
		return "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(hash), nil
}

func PasswordMatches(hash, password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}
