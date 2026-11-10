package jwt

import (
	"errors"
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

const (
	minimumJWTSecretLength = 32
	defaultAccessLifetime  = 15 * time.Minute
	defaultRefreshLifetime = 7 * 24 * time.Hour
	accessTokenType        = "access"
	refreshTokenType       = "refresh"
)

// Config owns the JWT signing key and token settings. Sign and verify tokens
// through its methods rather than reading SigningKey directly.
type Config struct {
	SigningKey      []byte
	AccessLifetime  time.Duration
	RefreshLifetime time.Duration
}

type jwtClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	gojwt.RegisteredClaims
}

func NewConfig(secret string) (*Config, error) {
	if len(secret) < minimumJWTSecretLength {
		return nil, fmt.Errorf("SECRET_KEY must be set to a random value of at least %d characters", minimumJWTSecretLength)
	}

	return &Config{
		SigningKey:      []byte(secret),
		AccessLifetime:  defaultAccessLifetime,
		RefreshLifetime: defaultRefreshLifetime,
	}, nil
}

func (c *Config) SignAccessToken(userID string) (string, error) {
	return c.signToken(userID, accessTokenType, c.AccessLifetime)
}

func (c *Config) SignRefreshToken(userID string) (string, error) {
	return c.signToken(userID, refreshTokenType, c.RefreshLifetime)
}

func (c *Config) signToken(userID, tokenType string, lifetime time.Duration) (string, error) {
	if userID == "" {
		return "", errors.New("cannot sign a token without a user ID")
	}

	now := time.Now()
	token := gojwt.NewWithClaims(gojwt.SigningMethodHS256, jwtClaims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: gojwt.RegisteredClaims{
			ExpiresAt: gojwt.NewNumericDate(now.Add(lifetime)),
			IssuedAt:  gojwt.NewNumericDate(now),
		},
	})

	return token.SignedString(c.SigningKey)
}

func (c *Config) VerifyAccessToken(tokenString string) (string, error) {
	return c.verifyToken(tokenString, accessTokenType)
}

func (c *Config) VerifyRefreshToken(tokenString string) (string, error) {
	return c.verifyToken(tokenString, refreshTokenType)
}

func (c *Config) verifyToken(tokenString, expectedType string) (string, error) {
	claims := new(jwtClaims)
	token, err := gojwt.ParseWithClaims(
		tokenString,
		claims,
		func(_ *gojwt.Token) (any, error) {
			return c.SigningKey, nil
		},
		gojwt.WithValidMethods([]string{gojwt.SigningMethodHS256.Alg()}),
		gojwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", err
	}
	if !token.Valid || claims.UserID == "" || claims.TokenType != expectedType {
		return "", errors.New("invalid authentication token")
	}

	return claims.UserID, nil
}

func (c *Config) AccessCookieMaxAge() int {
	return int(c.AccessLifetime / time.Second)
}

func (c *Config) RefreshCookieMaxAge() int {
	return int(c.RefreshLifetime / time.Second)
}
