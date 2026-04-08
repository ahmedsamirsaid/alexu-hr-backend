package legacy

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type NoopLeaveSyncAdapter struct{}

func NewNoopLeaveSyncAdapter() *NoopLeaveSyncAdapter {
	return &NoopLeaveSyncAdapter{}
}

func (a *NoopLeaveSyncAdapter) SyncLeaveRecord(ctx context.Context, record *domain.LeaveRecord, employee *domain.Employee, leaveType *domain.LeaveType) error {
	return nil
}
