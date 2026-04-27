package ports

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
)

type LeaveRequestListFilter struct {
	Status      *domain.ApprovalRequestStatus
	EmployeeUID *string
	// DepartmentUIDs limits results to requests whose employee belongs to one of these departments when non-nil.
	DepartmentUIDs []string
}

type LeaveRequestRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.LeaveRequest, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.LeaveRequest, error)
	GetByUIDs(ctx context.Context, q Querier, uids []string) ([]*domain.LeaveRequest, error)
	GetByApprovalRequestUID(ctx context.Context, q Querier, approvalRequestUID string) (*domain.LeaveRequest, error)
	Create(ctx context.Context, q Querier, request *domain.LeaveRequest) error
	Update(ctx context.Context, q Querier, request *domain.LeaveRequest) error

	// ListByEmployee returns all leave requests for an employee
	ListByEmployee(ctx context.Context, q Querier, employeeUID string) ([]*domain.LeaveRequest, error)

	// ListByEmployeePaginated returns paginated leave requests for an employee
	ListByEmployeePaginated(ctx context.Context, q Querier, employeeUID string, limit, offset int) ([]*domain.LeaveRequest, error)

	// CountByEmployee returns the count of leave requests for an employee
	CountByEmployee(ctx context.Context, q Querier, employeeUID string) (int, error)

	// HasOverlapping checks if there are overlapping pending or approved requests
	HasOverlapping(ctx context.Context, q Querier, employeeUID string, startDate, endDate time.Time, excludeUID *string) (bool, error)

	// List returns leave requests with optional filtering
	List(ctx context.Context, q Querier, filter LeaveRequestListFilter, limit, offset int) ([]*domain.LeaveRequest, error)

	// Count returns count of leave requests matching filter
	Count(ctx context.Context, q Querier, filter LeaveRequestListFilter) (int, error)

	// FindExpiredPending returns pending leave requests whose start date has passed beyond the grace period
	FindExpiredPending(ctx context.Context, q Querier, graceDays int) ([]*domain.LeaveRequest, error)
}
