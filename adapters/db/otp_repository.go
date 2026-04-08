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

type OTPRepository struct{}

func NewOTPRepository() *OTPRepository {
	return &OTPRepository{}
}

func (r *OTPRepository) Create(ctx context.Context, q ports.Querier, otp *domain.OTPCode) error {
	query := `
		INSERT INTO otp_codes (phone, code, expires_at, used, created_at)
		VALUES (?, ?, ?, ?, ?)`

	now := time.Now()
	otp.CreatedAt = now

	result, err := q.ExecContext(ctx, query,
		otp.Phone, otp.Code, otp.ExpiresAt, otp.Used, otp.CreatedAt)
	if err != nil {
		slog.Error("otp_repository.Create.exec_query", "error", err)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("otp_repository.Create.last_insert_id", "error", err)
		return err
	}
	otp.ID = id

	return nil
}

func (r *OTPRepository) GetLatestByPhone(ctx context.Context, q ports.Querier, phone string) (*domain.OTPCode, error) {
	query := `
		SELECT id, phone, code, expires_at, used, created_at
		FROM otp_codes
		WHERE phone = ?
		ORDER BY id DESC
		LIMIT 1`

	return r.scanOTP(q.QueryRowContext(ctx, query, phone))
}

func (r *OTPRepository) MarkUsed(ctx context.Context, q ports.Querier, id int64) error {
	query := `UPDATE otp_codes SET used = 1 WHERE id = ?`
	_, err := q.ExecContext(ctx, query, id)
	if err != nil {
		slog.Error("otp_repository.MarkUsed.exec_query", "error", err, "id", id)
	}
	return err
}

func (r *OTPRepository) scanOTP(row *sql.Row) (*domain.OTPCode, error) {
	var otp domain.OTPCode
	var expiresAt, createdAt domain.Time
	var used int
	err := row.Scan(&otp.ID, &otp.Phone, &otp.Code, &expiresAt, &used, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("otp_repository.scanOTP.scan_row", "error", err)
		return nil, err
	}
	otp.ExpiresAt = expiresAt.Time
	otp.CreatedAt = createdAt.Time
	otp.Used = used == 1
	return &otp, nil
}
