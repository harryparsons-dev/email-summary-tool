package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errNotFound = errors.New("record not found")

type Store interface {
	CreateUserWithRefresh(context.Context, User, StoredRefreshToken) (User, error)
	FindUserByEmail(context.Context, string) (User, error)
	FindUserByID(context.Context, string) (User, error)
	StoreRefresh(context.Context, StoredRefreshToken) error
	RotateRefresh(context.Context, []byte, string, string, StoredRefreshToken) (User, error)
	RevokeRefresh(context.Context, []byte) error
	StorePasswordReset(context.Context, string, []byte, time.Time) error
	ConsumePasswordReset(context.Context, []byte, string) error
}

type PostgresStore struct{ pool *pgxpool.Pool }

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore { return &PostgresStore{pool: pool} }

func (s *PostgresStore) CreateUserWithRefresh(ctx context.Context, user User, token StoredRefreshToken) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin registration: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	err = tx.QueryRow(ctx, `
		INSERT INTO users (id, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING created_at`, user.ID, user.Email, user.PasswordHash).Scan(&user.CreatedAt)
	if isUniqueViolation(err) {
		return User{}, ErrEmailInUse
	}
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	if err := insertRefresh(ctx, tx, token); err != nil {
		return User{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit registration: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) FindUserByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users WHERE email = $1`, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, errNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) FindUserByID(ctx context.Context, id string) (User, error) {
	var user User
	err := s.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM users WHERE id = $1`, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, errNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("find user by ID: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) StoreRefresh(ctx context.Context, token StoredRefreshToken) error {
	return insertRefresh(ctx, s.pool, token)
}

type queryer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func insertRefresh(ctx context.Context, q queryer, token StoredRefreshToken) error {
	_, err := q.Exec(ctx, `
		INSERT INTO refresh_tokens (token_hash, user_id, token_id, expires_at)
		VALUES ($1, $2, $3, $4)`, token.Hash, token.UserID, token.ID, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("store refresh token: %w", err)
	}
	return nil
}

func (s *PostgresStore) RotateRefresh(ctx context.Context, oldHash []byte, userID, tokenID string, next StoredRefreshToken) (User, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return User{}, fmt.Errorf("begin token rotation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	result, err := tx.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = NOW()
		WHERE token_hash = $1 AND user_id = $2 AND token_id = $3
		  AND revoked_at IS NULL AND expires_at > NOW()`, oldHash, userID, tokenID)
	if err != nil {
		return User{}, fmt.Errorf("revoke refresh token: %w", err)
	}
	if result.RowsAffected() != 1 {
		return User{}, ErrInvalidToken
	}
	if err := insertRefresh(ctx, tx, next); err != nil {
		return User{}, err
	}

	var user User
	err = tx.QueryRow(ctx, `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`, userID).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrInvalidToken
	}
	if err != nil {
		return User{}, fmt.Errorf("load session user: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return User{}, fmt.Errorf("commit token rotation: %w", err)
	}
	return user, nil
}

func (s *PostgresStore) RevokeRefresh(ctx context.Context, hash []byte) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE refresh_tokens SET revoked_at = COALESCE(revoked_at, NOW())
		WHERE token_hash = $1`, hash)
	if err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}
	return nil
}

func (s *PostgresStore) StorePasswordReset(ctx context.Context, userID string, hash []byte, expiresAt time.Time) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password reset: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if _, err := tx.Exec(ctx, `
		UPDATE password_reset_tokens SET used_at = NOW()
		WHERE user_id = $1 AND used_at IS NULL`, userID); err != nil {
		return fmt.Errorf("invalidate old password reset tokens: %w", err)
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO password_reset_tokens (token_hash, user_id, expires_at)
		VALUES ($1, $2, $3)`, hash, userID, expiresAt); err != nil {
		return fmt.Errorf("store password reset token: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit password reset: %w", err)
	}
	return nil
}

func (s *PostgresStore) ConsumePasswordReset(ctx context.Context, hash []byte, passwordHash string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin password update: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var userID string
	err = tx.QueryRow(ctx, `
		SELECT user_id FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > NOW()
		FOR UPDATE`, hash).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrInvalidToken
	}
	if err != nil {
		return fmt.Errorf("find password reset token: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE id = $2`, passwordHash, userID); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE user_id = $1 AND used_at IS NULL`, userID); err != nil {
		return fmt.Errorf("consume password reset tokens: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE refresh_tokens SET revoked_at = NOW() WHERE user_id = $1 AND revoked_at IS NULL`, userID); err != nil {
		return fmt.Errorf("revoke sessions after password reset: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit password update: %w", err)
	}
	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
