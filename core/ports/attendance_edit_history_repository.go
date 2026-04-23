package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type AttendanceEditHistoryWithEditor struct {
	History       *domain.AttendanceEditHistory
	EditedByName  string
	EditedByPhone *string
}

type AttendanceEditHistoryRepository interface {
	Create(ctx context.Context, q Querier, history *domain.AttendanceEditHistory) error
	ListByAttendanceRecordUID(ctx context.Context, q Querier, attendanceRecordUID string) ([]*AttendanceEditHistoryWithEditor, error)
}
