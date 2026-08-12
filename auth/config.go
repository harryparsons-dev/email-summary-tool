package auth

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	JWTSecret    string
	JWTIssuer    string
	JWTAudience  string
	AccessTTL    time.Duration
	RefreshTTL   time.Duration
	ResetTTL     time.Duration
	CookieSecure bool
	AppBaseURL   string
}

func ConfigFromEnv() (Config, error) {
	accessTTL, err := durationFromEnv("AUTH_ACCESS_TOKEN_TTL", 15*time.Minute)
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := durationFromEnv("AUTH_REFRESH_TOKEN_TTL", 30*24*time.Hour)
	if err != nil {
		return Config{}, err
	}
	resetTTL, err := durationFromEnv("AUTH_PASSWORD_RESET_TTL", time.Hour)
	if err != nil {
		return Config{}, err
	}
	secure, err := strconv.ParseBool(valueOrDefault("AUTH_COOKIE_SECURE", "false"))
	if err != nil {
		return Config{}, fmt.Errorf("parse AUTH_COOKIE_SECURE: %w", err)
	}
	config := Config{
		JWTSecret: os.Getenv("JWT_SECRET"), JWTIssuer: valueOrDefault("JWT_ISSUER", "email-summary-tool"),
		JWTAudience: valueOrDefault("JWT_AUDIENCE", "email-summary-tool-web"), AccessTTL: accessTTL,
		RefreshTTL: refreshTTL, ResetTTL: resetTTL, CookieSecure: secure,
		AppBaseURL: valueOrDefault("APP_BASE_URL", "http://localhost:5173"),
	}
	if len(config.JWTSecret) < 32 {
		return Config{}, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}
	if strings.EqualFold(os.Getenv("APP_ENV"), "production") && !config.CookieSecure {
		return Config{}, fmt.Errorf("AUTH_COOKIE_SECURE must be true in production")
	}
	return config, nil
}

func durationFromEnv(name string, fallback time.Duration) (time.Duration, error) {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive Go duration", name)
	}
	return value, nil
}

func valueOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
