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

type AttendanceReminderRepository struct{}

func NewAttendanceReminderRepository() *AttendanceReminderRepository {
	return &AttendanceReminderRepository{}
}

func (r *AttendanceReminderRepository) Create(ctx context.Context, q ports.Querier, reminder *domain.AttendanceReminder) error {
	query := `
		INSERT INTO attendance_reminders (uid, employee_uid, attendance_date, reminder_type, sent_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	if reminder.SentAt.IsZero() {
		reminder.SentAt = now
	}
	if reminder.CreatedAt.IsZero() {
		reminder.CreatedAt = now
	}
	reminder.UpdatedAt = now

	_, err := q.ExecContext(ctx, query,
		reminder.UID,
		reminder.EmployeeUID,
		reminder.AttendanceDate,
		reminder.ReminderType,
		reminder.SentAt.Format(time.RFC3339Nano),
		reminder.CreatedAt,
		reminder.UpdatedAt,
	)
	if err != nil {
		slog.Error("attendance_reminder_repository.Create.exec_query", "error", err, "employee_uid", reminder.EmployeeUID)
	}

	return err
}

func (r *AttendanceReminderRepository) GetByEmployeeDateAndType(ctx context.Context, q ports.Querier, employeeUID, attendanceDate string, reminderType domain.AttendanceReminderType) (*domain.AttendanceReminder, error) {
	query := `
		SELECT id, uid, employee_uid, attendance_date, reminder_type, sent_at, created_at, updated_at
		FROM attendance_reminders
		WHERE employee_uid = ? AND attendance_date = ? AND reminder_type = ?
		LIMIT 1`

	return r.scanReminder(q.QueryRowContext(ctx, query, employeeUID, attendanceDate, reminderType))
}

func (r *AttendanceReminderRepository) scanReminder(row *sql.Row) (*domain.AttendanceReminder, error) {
	var reminder domain.AttendanceReminder
	var sentAt, createdAt, updatedAt domain.Time

	err := row.Scan(
		&reminder.ID,
		&reminder.UID,
		&reminder.EmployeeUID,
		&reminder.AttendanceDate,
		&reminder.ReminderType,
		&sentAt,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("attendance_reminder_repository.scanReminder.scan_row", "error", err)
		return nil, err
	}

	reminder.SentAt = sentAt.Time
	reminder.CreatedAt = createdAt.Time
	reminder.UpdatedAt = updatedAt.Time
	return &reminder, nil
}
