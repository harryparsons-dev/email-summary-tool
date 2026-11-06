package config

import (
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const testJWTSecret = "a-test-secret-that-is-at-least-32-characters-long"

func TestNewJWTConfigRejectsShortSecret(t *testing.T) {
	_, err := NewJWTConfig("too-short")
	if err == nil {
		t.Fatal("NewJWTConfig() accepted a short secret")
	}
}

func TestJWTConfigSignsAndVerifiesAccessToken(t *testing.T) {
	jwtConfig, err := NewJWTConfig(testJWTSecret)
	if err != nil {
		t.Fatalf("NewJWTConfig() error = %v", err)
	}

	token, err := jwtConfig.SignAccessToken("user-123")
	if err != nil {
		t.Fatalf("SignAccessToken() error = %v", err)
	}

	userID, err := jwtConfig.VerifyAccessToken(token)
	if err != nil {
		t.Fatalf("VerifyAccessToken() error = %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("VerifyAccessToken() user ID = %q, want %q", userID, "user-123")
	}
}

func TestJWTConfigKeepsRefreshTokenValidAfterAccessTokenExpires(t *testing.T) {
	jwtConfig, err := NewJWTConfig(testJWTSecret)
	if err != nil {
		t.Fatalf("NewJWTConfig() error = %v", err)
	}
	jwtConfig.AccessLifetime = -time.Minute

	accessToken, err := jwtConfig.SignAccessToken("user-123")
	if err != nil {
		t.Fatalf("SignAccessToken() error = %v", err)
	}
	refreshToken, err := jwtConfig.SignRefreshToken("user-123")
	if err != nil {
		t.Fatalf("SignRefreshToken() error = %v", err)
	}

	if _, err := jwtConfig.VerifyAccessToken(accessToken); err == nil {
		t.Fatal("VerifyAccessToken() accepted an expired access token")
	}
	if _, err := jwtConfig.VerifyRefreshToken(refreshToken); err != nil {
		t.Fatalf("VerifyRefreshToken() rejected a valid refresh token: %v", err)
	}
}

func TestJWTConfigDoesNotAcceptAccessTokenAsRefreshToken(t *testing.T) {
	jwtConfig, err := NewJWTConfig(testJWTSecret)
	if err != nil {
		t.Fatalf("NewJWTConfig() error = %v", err)
	}

	accessToken, err := jwtConfig.SignAccessToken("user-123")
	if err != nil {
		t.Fatalf("SignAccessToken() error = %v", err)
	}

	if _, err := jwtConfig.VerifyRefreshToken(accessToken); err == nil {
		t.Fatal("VerifyRefreshToken() accepted an access token")
	}
}

func TestJWTConfigRejectsTokenSignedWithAnotherKey(t *testing.T) {
	jwtConfig, err := NewJWTConfig(testJWTSecret)
	if err != nil {
		t.Fatalf("NewJWTConfig() error = %v", err)
	}

	otherKey := []byte(strings.Repeat("b", 32))
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user-123",
	}).SignedString(otherKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := jwtConfig.VerifyAccessToken(token); err == nil {
		t.Fatal("VerifyAccessToken() accepted a token signed with another key")
	}
}

func TestJWTConfigRejectsTokenWithoutExpiration(t *testing.T) {
	jwtConfig, err := NewJWTConfig(testJWTSecret)
	if err != nil {
		t.Fatalf("NewJWTConfig() error = %v", err)
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": "user-123",
	}).SignedString(jwtConfig.SigningKey)
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}

	if _, err := jwtConfig.VerifyAccessToken(token); err == nil {
		t.Fatal("VerifyAccessToken() accepted a token without an expiration")
	}
}
