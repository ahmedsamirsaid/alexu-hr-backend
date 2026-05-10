package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type RefreshTokenRepository struct{}

func NewRefreshTokenRepository() *RefreshTokenRepository {
	return &RefreshTokenRepository{}
}

func (r *RefreshTokenRepository) Create(ctx context.Context, q ports.Querier, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (token_hash, user_id, expires_at, revoked, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	token.CreatedAt = now

	err := q.QueryRowContext(ctx, query,
		token.TokenHash, token.UserID, token.ExpiresAt, token.Revoked, token.CreatedAt).Scan(&token.ID)
	if err != nil {
		slog.Error("refresh_token_repository.Create.exec_query", "error", err, "user_id", token.UserID)
		return err
	}

	return nil
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, q ports.Querier, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, token_hash, user_id, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE token_hash = $1`

	return r.scanToken(q.QueryRowContext(ctx, query, tokenHash))
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, q ports.Querier, id int64) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE id = $1`
	_, err := q.ExecContext(ctx, query, id)
	if err != nil {
		slog.Error("refresh_token_repository.Revoke.exec_query", "error", err, "id", id)
	}
	return err
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, q ports.Querier, userID int64) error {
	query := `UPDATE refresh_tokens SET revoked = true WHERE user_id = $1`
	_, err := q.ExecContext(ctx, query, userID)
	if err != nil {
		slog.Error("refresh_token_repository.RevokeAllForUser.exec_query", "error", err, "user_id", userID)
	}
	return err
}

func (r *RefreshTokenRepository) scanToken(row *sql.Row) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	var expiresAt, createdAt domain.Time
	err := row.Scan(&token.ID, &token.TokenHash, &token.UserID, &expiresAt, &token.Revoked, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("refresh_token_repository.scanToken.scan_row", "error", err)
		return nil, err
	}
	token.ExpiresAt = expiresAt.Time
	token.CreatedAt = createdAt.Time
	return &token, nil
}
