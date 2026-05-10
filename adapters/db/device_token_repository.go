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

type DeviceTokenRepository struct{}

func NewDeviceTokenRepository() *DeviceTokenRepository {
	return &DeviceTokenRepository{}
}

func (r *DeviceTokenRepository) Create(ctx context.Context, q ports.Querier, token *domain.DeviceToken) error {
	// Use INSERT OR REPLACE to handle duplicate tokens (same FCM token for potentially different user)
	query := `
		INSERT INTO device_tokens (uid, user_uid, token, platform, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT(token) DO UPDATE SET
			user_uid = excluded.user_uid,
			updated_at = excluded.updated_at
		RETURNING id`

	now := time.Now()
	token.CreatedAt = now
	token.UpdatedAt = now

	err := q.QueryRowContext(ctx, query,
		token.UID, token.UserUID, token.Token, token.Platform, token.CreatedAt, token.UpdatedAt).Scan(&token.ID)
	if err != nil {
		slog.Error("device_token_repository.Create.exec_query", "error", err, "user_uid", token.UserUID)
		return err
	}

	return nil
}

func (r *DeviceTokenRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.DeviceToken, error) {
	query := `
		SELECT id, uid, user_uid, token, platform, created_at, updated_at
		FROM device_tokens
		WHERE uid = $1`

	return r.scanToken(q.QueryRowContext(ctx, query, uid))
}

func (r *DeviceTokenRepository) GetByToken(ctx context.Context, q ports.Querier, token string) (*domain.DeviceToken, error) {
	query := `
		SELECT id, uid, user_uid, token, platform, created_at, updated_at
		FROM device_tokens
		WHERE token = $1`

	return r.scanToken(q.QueryRowContext(ctx, query, token))
}

func (r *DeviceTokenRepository) GetByUserUID(ctx context.Context, q ports.Querier, userUID string) ([]*domain.DeviceToken, error) {
	query := `
		SELECT id, uid, user_uid, token, platform, created_at, updated_at
		FROM device_tokens
		WHERE user_uid = $1
		ORDER BY created_at DESC`

	rows, err := q.QueryContext(ctx, query, userUID)
	if err != nil {
		slog.Error("device_token_repository.GetByUserUID.query", "error", err, "user_uid", userUID)
		return nil, err
	}
	defer rows.Close()

	var tokens []*domain.DeviceToken
	for rows.Next() {
		token, err := r.scanTokenRow(rows)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}

	if err := rows.Err(); err != nil {
		slog.Error("device_token_repository.GetByUserUID.rows_err", "error", err, "user_uid", userUID)
		return nil, err
	}

	return tokens, nil
}

func (r *DeviceTokenRepository) Delete(ctx context.Context, q ports.Querier, uid string) error {
	query := `DELETE FROM device_tokens WHERE uid = $1`
	_, err := q.ExecContext(ctx, query, uid)
	if err != nil {
		slog.Error("device_token_repository.Delete.exec_query", "error", err, "uid", uid)
	}
	return err
}

func (r *DeviceTokenRepository) DeleteByToken(ctx context.Context, q ports.Querier, token string) error {
	query := `DELETE FROM device_tokens WHERE token = $1`
	_, err := q.ExecContext(ctx, query, token)
	if err != nil {
		slog.Error("device_token_repository.DeleteByToken.exec_query", "error", err)
	}
	return err
}

func (r *DeviceTokenRepository) DeleteAllForUser(ctx context.Context, q ports.Querier, userUID string) error {
	query := `DELETE FROM device_tokens WHERE user_uid = $1`
	_, err := q.ExecContext(ctx, query, userUID)
	if err != nil {
		slog.Error("device_token_repository.DeleteAllForUser.exec_query", "error", err, "user_uid", userUID)
	}
	return err
}

func (r *DeviceTokenRepository) scanToken(row *sql.Row) (*domain.DeviceToken, error) {
	var token domain.DeviceToken
	var createdAt, updatedAt domain.Time

	err := row.Scan(&token.ID, &token.UID, &token.UserUID, &token.Token, &token.Platform, &createdAt, &updatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("device_token_repository.scanToken.scan_row", "error", err)
		return nil, err
	}

	token.CreatedAt = createdAt.Time
	token.UpdatedAt = updatedAt.Time
	return &token, nil
}

func (r *DeviceTokenRepository) scanTokenRow(rows *sql.Rows) (*domain.DeviceToken, error) {
	var token domain.DeviceToken
	var createdAt, updatedAt domain.Time

	err := rows.Scan(&token.ID, &token.UID, &token.UserUID, &token.Token, &token.Platform, &createdAt, &updatedAt)
	if err != nil {
		slog.Error("device_token_repository.scanTokenRow.scan_row", "error", err)
		return nil, err
	}

	token.CreatedAt = createdAt.Time
	token.UpdatedAt = updatedAt.Time
	return &token, nil
}
