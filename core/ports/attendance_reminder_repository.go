package ports

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
)

type AttendanceReminderRepository interface {
	Create(ctx context.Context, q Querier, reminder *domain.AttendanceReminder) error
	GetByEmployeeDateAndType(ctx context.Context, q Querier, employeeUID string, attendanceDate time.Time, reminderType domain.AttendanceReminderType) (*domain.AttendanceReminder, error)
}
