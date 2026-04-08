package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type LeaveSyncPort interface {
	SyncLeaveRecord(ctx context.Context, record *domain.LeaveRecord, employee *domain.Employee, leaveType *domain.LeaveType) error
}
