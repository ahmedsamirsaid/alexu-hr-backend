package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type AttendanceReminderRepository interface {
	Create(ctx context.Context, q Querier, reminder *domain.AttendanceReminder) error
	GetByEmployeeDateAndType(ctx context.Context, q Querier, employeeUID, attendanceDate string, reminderType domain.AttendanceReminderType) (*domain.AttendanceReminder, error)
}
