package ports

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
)

type LeaveRecordWithEmployee struct {
	LeaveRecord  *domain.LeaveRecord
	EmployeeUID  string
	EmployeeName string
}

type ListAllLeaveRecordsFilter struct {
	Search      string
	LeaveTypeID *int64
	StartDate   *time.Time
	EndDate     *time.Time
}

type LeaveRecordRepository interface {
	GetByID(ctx context.Context, q Querier, id int64) (*domain.LeaveRecord, error)
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.LeaveRecord, error)
	Create(ctx context.Context, q Querier, record *domain.LeaveRecord) error
	ListByEmployee(ctx context.Context, q Querier, employeeID int64) ([]*domain.LeaveRecord, error)
	ListByEmployeePaginated(ctx context.Context, q Querier, employeeID int64, limit, offset int) ([]*domain.LeaveRecord, error)
	CountByEmployee(ctx context.Context, q Querier, employeeID int64) (int, error)
	ListByEmployeeAndDateRange(ctx context.Context, q Querier, employeeID int64, start, end time.Time) ([]*domain.LeaveRecord, error)
	ListByEmployeeAndType(ctx context.Context, q Querier, employeeID, leaveTypeID int64) ([]*domain.LeaveRecord, error)
	ListAllPaginated(ctx context.Context, q Querier, filter ListAllLeaveRecordsFilter, limit, offset int) ([]*LeaveRecordWithEmployee, error)
	CountAll(ctx context.Context, q Querier, filter ListAllLeaveRecordsFilter) (int, error)

	// CountOnLeaveToday returns the number of employees on leave for the given date.
	CountOnLeaveToday(ctx context.Context, q Querier, date time.Time) (int, error)
	// HasLeaveOnDate returns whether the employee has any leave covering the given date.
	HasLeaveOnDate(ctx context.Context, q Querier, employeeID int64, date time.Time) (bool, error)
}
