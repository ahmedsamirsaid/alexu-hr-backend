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
		VALUES (?, ?, ?, ?, ?)`

	now := time.Now()
	token.CreatedAt = now

	result, err := q.ExecContext(ctx, query,
		token.TokenHash, token.UserID, token.ExpiresAt, token.Revoked, token.CreatedAt)
	if err != nil {
		slog.Error("refresh_token_repository.Create.exec_query", "error", err, "user_id", token.UserID)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("refresh_token_repository.Create.last_insert_id", "error", err, "user_id", token.UserID)
		return err
	}
	token.ID = id

	return nil
}

func (r *RefreshTokenRepository) GetByHash(ctx context.Context, q ports.Querier, tokenHash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, token_hash, user_id, expires_at, revoked, created_at
		FROM refresh_tokens
		WHERE token_hash = ?`

	return r.scanToken(q.QueryRowContext(ctx, query, tokenHash))
}

func (r *RefreshTokenRepository) Revoke(ctx context.Context, q ports.Querier, id int64) error {
	query := `UPDATE refresh_tokens SET revoked = 1 WHERE id = ?`
	_, err := q.ExecContext(ctx, query, id)
	if err != nil {
		slog.Error("refresh_token_repository.Revoke.exec_query", "error", err, "id", id)
	}
	return err
}

func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, q ports.Querier, userID int64) error {
	query := `UPDATE refresh_tokens SET revoked = 1 WHERE user_id = ?`
	_, err := q.ExecContext(ctx, query, userID)
	if err != nil {
		slog.Error("refresh_token_repository.RevokeAllForUser.exec_query", "error", err, "user_id", userID)
	}
	return err
}

func (r *RefreshTokenRepository) scanToken(row *sql.Row) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	var expiresAt, createdAt domain.Time
	var revoked int
	err := row.Scan(&token.ID, &token.TokenHash, &token.UserID, &expiresAt, &revoked, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("refresh_token_repository.scanToken.scan_row", "error", err)
		return nil, err
	}
	token.ExpiresAt = expiresAt.Time
	token.CreatedAt = createdAt.Time
	token.Revoked = revoked == 1
	return &token, nil
}
