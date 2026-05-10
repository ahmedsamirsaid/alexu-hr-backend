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
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	now := time.Now()
	otp.CreatedAt = now

	err := q.QueryRowContext(ctx, query,
		otp.Phone, otp.Code, otp.ExpiresAt, otp.Used, otp.CreatedAt).Scan(&otp.ID)
	if err != nil {
		slog.Error("otp_repository.Create.exec_query", "error", err)
		return err
	}

	return nil
}

func (r *OTPRepository) GetLatestByPhone(ctx context.Context, q ports.Querier, phone string) (*domain.OTPCode, error) {
	query := `
		SELECT id, phone, code, expires_at, used, created_at
		FROM otp_codes
		WHERE phone = $1
		ORDER BY id DESC
		LIMIT 1`

	return r.scanOTP(q.QueryRowContext(ctx, query, phone))
}

func (r *OTPRepository) MarkUsed(ctx context.Context, q ports.Querier, id int64) error {
	query := `UPDATE otp_codes SET used = true WHERE id = $1`
	_, err := q.ExecContext(ctx, query, id)
	if err != nil {
		slog.Error("otp_repository.MarkUsed.exec_query", "error", err, "id", id)
	}
	return err
}

func (r *OTPRepository) scanOTP(row *sql.Row) (*domain.OTPCode, error) {
	var otp domain.OTPCode
	var expiresAt, createdAt domain.Time
	err := row.Scan(&otp.ID, &otp.Phone, &otp.Code, &expiresAt, &otp.Used, &createdAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("otp_repository.scanOTP.scan_row", "error", err)
		return nil, err
	}
	otp.ExpiresAt = expiresAt.Time
	otp.CreatedAt = createdAt.Time
	return &otp, nil
}
