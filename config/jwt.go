package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	minimumJWTSecretLength = 32
	defaultAccessLifetime  = 15 * time.Minute
	defaultRefreshLifetime = 7 * 24 * time.Hour
	accessTokenType        = "access"
	refreshTokenType       = "refresh"
)

// JWTConfig owns the JWT signing key and token settings. Keeping the key
// private prevents handlers and middleware from retaining their own direct
// references to it.
type JWTConfig struct {
	signingKey      []byte
	accessLifetime  time.Duration
	refreshLifetime time.Duration
}

type jwtClaims struct {
	UserID    string `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func NewJWTConfig(secret string) (*JWTConfig, error) {
	if len(secret) < minimumJWTSecretLength {
		return nil, fmt.Errorf("SECRET_KEY must be set to a random value of at least %d characters", minimumJWTSecretLength)
	}

	return &JWTConfig{
		signingKey:      []byte(secret),
		accessLifetime:  defaultAccessLifetime,
		refreshLifetime: defaultRefreshLifetime,
	}, nil
}

func (c *JWTConfig) SignAccessToken(userID string) (string, error) {
	return c.signToken(userID, accessTokenType, c.accessLifetime)
}

func (c *JWTConfig) SignRefreshToken(userID string) (string, error) {
	return c.signToken(userID, refreshTokenType, c.refreshLifetime)
}

func (c *JWTConfig) signToken(userID, tokenType string, lifetime time.Duration) (string, error) {
	if userID == "" {
		return "", errors.New("cannot sign a token without a user ID")
	}

	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims{
		UserID:    userID,
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(lifetime)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	})

	return token.SignedString(c.signingKey)
}

func (c *JWTConfig) VerifyAccessToken(tokenString string) (string, error) {
	return c.verifyToken(tokenString, accessTokenType)
}

func (c *JWTConfig) VerifyRefreshToken(tokenString string) (string, error) {
	return c.verifyToken(tokenString, refreshTokenType)
}

func (c *JWTConfig) verifyToken(tokenString, expectedType string) (string, error) {
	claims := new(jwtClaims)
	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(_ *jwt.Token) (any, error) {
			return c.signingKey, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return "", err
	}
	if !token.Valid || claims.UserID == "" || claims.TokenType != expectedType {
		return "", errors.New("invalid authentication token")
	}

	return claims.UserID, nil
}

func (c *JWTConfig) AccessCookieMaxAge() int {
	return int(c.accessLifetime / time.Second)
}

func (c *JWTConfig) RefreshCookieMaxAge() int {
	return int(c.refreshLifetime / time.Second)
}
