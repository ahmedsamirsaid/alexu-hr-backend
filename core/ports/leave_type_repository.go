package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type LeaveTypeRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.LeaveType, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.LeaveType, error)
	GetByCode(ctx context.Context, q Querier, code string) (*domain.LeaveType, error)
	GetSubLeaveTypeByUID(ctx context.Context, q Querier, uid string) (*domain.SubLeaveType, error)
	List(ctx context.Context, q Querier, activeOnly bool) ([]*domain.LeaveType, error)
	ListSubLeaveTypesByLeaveTypeUID(ctx context.Context, q Querier, leaveTypeUID string) ([]*domain.SubLeaveType, error)
	SetActive(ctx context.Context, q Querier, uid string, isActive bool) error
}
