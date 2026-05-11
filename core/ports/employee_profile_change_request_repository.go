package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type EmployeeProfileChangeRequestRepository interface {
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.EmployeeProfileChangeRequest, error)
	GetByApprovalRequestUID(ctx context.Context, q Querier, approvalRequestUID string) (*domain.EmployeeProfileChangeRequest, error)
	Create(ctx context.Context, q Querier, request *domain.EmployeeProfileChangeRequest) error
	ListByEmployeeUID(ctx context.Context, q Querier, employeeUID string) ([]*domain.EmployeeProfileChangeRequest, error)
	ListPending(ctx context.Context, q Querier) ([]*domain.EmployeeProfileChangeRequest, error)
	HasPendingForEmployee(ctx context.Context, q Querier, employeeUID string) (bool, error)
}
