package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	accessTokenType  = "access"
	refreshTokenType = "refresh"
)

type Claims struct {
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret     []byte
	issuer     string
	audience   string
	accessTTL  time.Duration
	refreshTTL time.Duration
	now        func() time.Time
}

func NewTokenManager(secret, issuer, audience string, accessTTL, refreshTTL time.Duration) (*TokenManager, error) {
	if len(secret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must contain at least 32 characters")
	}
	if issuer == "" || audience == "" {
		return nil, fmt.Errorf("JWT issuer and audience must not be empty")
	}
	if accessTTL <= 0 || refreshTTL <= 0 {
		return nil, fmt.Errorf("token lifetimes must be positive")
	}
	return &TokenManager{
		secret: []byte(secret), issuer: issuer, audience: audience,
		accessTTL: accessTTL, refreshTTL: refreshTTL, now: time.Now,
	}, nil
}

func (m *TokenManager) NewPair(userID string) (access string, refresh string, stored StoredRefreshToken, err error) {
	now := m.now().UTC()
	accessID, err := randomUUID()
	if err != nil {
		return "", "", stored, err
	}
	refreshID, err := randomUUID()
	if err != nil {
		return "", "", stored, err
	}

	access, err = m.sign(userID, accessID, accessTokenType, now, m.accessTTL)
	if err != nil {
		return "", "", stored, err
	}
	refresh, err = m.sign(userID, refreshID, refreshTokenType, now, m.refreshTTL)
	if err != nil {
		return "", "", stored, err
	}
	stored = StoredRefreshToken{
		Hash: HashToken(refresh), ID: refreshID, UserID: userID, ExpiresAt: now.Add(m.refreshTTL),
	}
	return access, refresh, stored, nil
}

func (m *TokenManager) sign(userID, tokenID, tokenType string, now time.Time, ttl time.Duration) (string, error) {
	claims := Claims{
		TokenType: tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer: m.issuer, Subject: userID,
			Audience:  jwt.ClaimStrings{m.audience},
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			NotBefore: jwt.NewNumericDate(now), IssuedAt: jwt.NewNumericDate(now), ID: tokenID,
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
}

func (m *TokenManager) ParseRefresh(raw string) (*Claims, error) {
	return m.parse(raw, refreshTokenType)
}

func (m *TokenManager) ParseAccess(raw string) (*Claims, error) {
	return m.parse(raw, accessTokenType)
}

func (m *TokenManager) parse(raw, expectedType string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		return m.secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithIssuer(m.issuer), jwt.WithAudience(m.audience), jwt.WithLeeway(5*time.Second))
	if err != nil || !token.Valid || claims.TokenType != expectedType || claims.Subject == "" || claims.ID == "" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *TokenManager) AccessTTLSeconds() int64   { return int64(m.accessTTL.Seconds()) }
func (m *TokenManager) RefreshTTL() time.Duration { return m.refreshTTL }

func HashToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

func NewOpaqueToken() (raw string, hash []byte, err error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil, fmt.Errorf("generate secure token: %w", err)
	}
	raw = base64.RawURLEncoding.EncodeToString(b)
	return raw, HashToken(raw), nil
}

func randomUUID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate UUID: %w", err)
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	dst := make([]byte, 36)
	hex.Encode(dst[0:8], b[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], b[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], b[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], b[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:36], b[10:16])
	return string(dst), nil
}
