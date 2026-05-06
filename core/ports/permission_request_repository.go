package ports

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
)

type PermissionRequestListFilter struct {
	Status         *domain.ApprovalRequestStatus
	Type           *domain.PermissionType
	EmployeeUID    *string
	DepartmentUIDs []string
	DateFrom       *time.Time
	DateTo         *time.Time
}

type PermissionRequestRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.PermissionRequest, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.PermissionRequest, error)
	GetByApprovalRequestUID(ctx context.Context, q Querier, approvalRequestUID string) (*domain.PermissionRequest, error)
	Create(ctx context.Context, q Querier, request *domain.PermissionRequest) error
	Update(ctx context.Context, q Querier, request *domain.PermissionRequest) error

	// Same-day approved permissions for one employee. Used by the attendance adjuster.
	ListApprovedForEmployeeOnDate(ctx context.Context, q Querier, employeeUID string, date time.Time) ([]*domain.PermissionRequest, error)

	// Bulk variant — returns map[employeeUID][]permission for all approved permissions on the given date.
	ListApprovedForEmployeesOnDate(ctx context.Context, q Querier, employeeUIDs []string, date time.Time) (map[string][]*domain.PermissionRequest, error)

	// HasOverlappingOnDate detects same-day overlap with pending or approved permissions.
	// startTime / endTime are HH:MM strings — every row in the table now carries a resolved
	// window so the comparison is a straight string compare.
	HasOverlappingOnDate(ctx context.Context, q Querier, employeeUID string, date time.Time, startTime, endTime string, excludeUID *string) (bool, error)

	// CountInWeek counts permissions in [weekStart, weekEnd] (inclusive) of any of the given types
	// in any of the given statuses for the employee.
	CountInWeek(ctx context.Context, q Querier, employeeUID string, types []domain.PermissionType, weekStart, weekEnd time.Time, statuses []domain.ApprovalRequestStatus, excludeUID *string) (int, error)

	// List with filters + pagination.
	List(ctx context.Context, q Querier, filter PermissionRequestListFilter, limit, offset int) ([]*domain.PermissionRequest, error)
	Count(ctx context.Context, q Querier, filter PermissionRequestListFilter) (int, error)

	// FindExpiredPending returns pending permission requests whose deadline has passed at asOf.
	// The repository compares `permission_date` and the type-specific cutoff time against asOf.
	FindExpiredPending(ctx context.Context, q Querier, asOf time.Time) ([]*domain.PermissionRequest, error)
}
