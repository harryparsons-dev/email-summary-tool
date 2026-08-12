package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type ResetMailer interface {
	SendPasswordReset(context.Context, string, string) error
}

type Service struct {
	store     Store
	tokens    *TokenManager
	mailer    ResetMailer
	resetTTL  time.Duration
	dummyHash string
	now       func() time.Time
}

func NewService(store Store, tokens *TokenManager, mailer ResetMailer, resetTTL time.Duration) (*Service, error) {
	if store == nil || tokens == nil || mailer == nil || resetTTL <= 0 {
		return nil, fmt.Errorf("invalid authentication service configuration")
	}
	dummy, err := bcrypt.GenerateFromPassword([]byte("not-a-real-password"), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("initialize password verifier: %w", err)
	}
	return &Service{store: store, tokens: tokens, mailer: mailer, resetTTL: resetTTL, dummyHash: string(dummy), now: time.Now}, nil
}

func (s *Service) Register(ctx context.Context, email, password string) (Session, error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return Session{}, err
	}
	hash, err := HashPassword(password)
	if err != nil {
		return Session{}, err
	}
	userID, err := randomUUID()
	if err != nil {
		return Session{}, err
	}
	access, refresh, stored, err := s.tokens.NewPair(userID)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	user, err := s.store.CreateUserWithRefresh(ctx, User{ID: userID, Email: normalized, PasswordHash: hash}, stored)
	if err != nil {
		return Session{}, err
	}
	return s.session(user, access, refresh), nil
}

func (s *Service) Login(ctx context.Context, email, password string) (Session, error) {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return Session{}, ErrInvalidCredential
	}
	user, err := s.store.FindUserByEmail(ctx, normalized)
	if errors.Is(err, errNotFound) {
		_ = bcrypt.CompareHashAndPassword([]byte(s.dummyHash), []byte(password))
		return Session{}, ErrInvalidCredential
	}
	if err != nil {
		return Session{}, err
	}
	if !PasswordMatches(user.PasswordHash, password) {
		return Session{}, ErrInvalidCredential
	}
	access, refresh, stored, err := s.tokens.NewPair(user.ID)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	if err := s.store.StoreRefresh(ctx, stored); err != nil {
		return Session{}, err
	}
	return s.session(user, access, refresh), nil
}

func (s *Service) Refresh(ctx context.Context, raw string) (Session, error) {
	claims, err := s.tokens.ParseRefresh(raw)
	if err != nil {
		return Session{}, ErrInvalidToken
	}
	access, refresh, next, err := s.tokens.NewPair(claims.Subject)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	user, err := s.store.RotateRefresh(ctx, HashToken(raw), claims.Subject, claims.ID, next)
	if err != nil {
		return Session{}, err
	}
	return s.session(user, access, refresh), nil
}

func (s *Service) Logout(ctx context.Context, raw string) error {
	if raw == "" {
		return nil
	}
	return s.store.RevokeRefresh(ctx, HashToken(raw))
}

func (s *Service) User(ctx context.Context, id string) (User, error) {
	user, err := s.store.FindUserByID(ctx, id)
	if errors.Is(err, errNotFound) {
		return User{}, ErrInvalidToken
	}
	return user, err
}

func (s *Service) RequestPasswordReset(ctx context.Context, email string) error {
	normalized, err := NormalizeEmail(email)
	if err != nil {
		return nil
	}
	user, err := s.store.FindUserByEmail(ctx, normalized)
	if errors.Is(err, errNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	raw, hash, err := NewOpaqueToken()
	if err != nil {
		return err
	}
	if err := s.store.StorePasswordReset(ctx, user.ID, hash, s.now().UTC().Add(s.resetTTL)); err != nil {
		return err
	}
	if err := s.mailer.SendPasswordReset(ctx, user.Email, raw); err != nil {
		return fmt.Errorf("send password reset: %w", err)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, raw, password string) error {
	if raw == "" {
		return ErrInvalidToken
	}
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}
	return s.store.ConsumePasswordReset(ctx, HashToken(raw), hash)
}

func (s *Service) session(user User, access, refresh string) Session {
	user.PasswordHash = ""
	return Session{User: user, AccessToken: access, RefreshToken: refresh, ExpiresIn: s.tokens.AccessTTLSeconds()}
}
