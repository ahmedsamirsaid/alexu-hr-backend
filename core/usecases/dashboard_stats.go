package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetDashboardStatsInput struct {
	ManagedDepartmentUIDs []string
}

type DashboardAttendanceRecordRepository interface {
	ListByDate(ctx context.Context, q ports.Querier, date time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error)
}

// DashboardStatsOutput contains the dashboard statistics.
type DashboardStatsOutput struct {
	TotalEmployees  int
	CheckedInToday  int
	CheckedOutToday int
	LeavesToday     int
	PendingRequests int
}

// GetDashboardStatsUseCase handles retrieving dashboard statistics.
type GetDashboardStatsUseCase struct {
	db               ports.DB
	employeeRepo     ports.EmployeeRepository
	leaveRecordRepo  ports.LeaveRecordRepository
	leaveRequestRepo ports.LeaveRequestRepository
	attendanceRepo   DashboardAttendanceRecordRepository
}

// NewGetDashboardStatsUseCase creates a new get dashboard stats use case.
func NewGetDashboardStatsUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	leaveRecordRepo ports.LeaveRecordRepository,
	leaveRequestRepo ports.LeaveRequestRepository,
	attendanceRepo DashboardAttendanceRecordRepository,
) *GetDashboardStatsUseCase {
	return &GetDashboardStatsUseCase{
		db:               db,
		employeeRepo:     employeeRepo,
		leaveRecordRepo:  leaveRecordRepo,
		leaveRequestRepo: leaveRequestRepo,
		attendanceRepo:   attendanceRepo,
	}
}

// Execute retrieves dashboard statistics.
func (uc *GetDashboardStatsUseCase) Execute(ctx context.Context, input GetDashboardStatsInput) (*DashboardStatsOutput, error) {
	if len(input.ManagedDepartmentUIDs) > 0 {
		return uc.executeScoped(ctx, input.ManagedDepartmentUIDs)
	}

	totalEmployees, err := uc.employeeRepo.Count(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	// Use UTC day boundaries to stay consistent with PostgreSQL date functions and
	// frontend date filters that are based on ISO date strings.
	today := time.Now().UTC()
	leavesToday, err := uc.leaveRecordRepo.CountOnLeaveToday(ctx, uc.db, today)
	if err != nil {
		return nil, err
	}

	checkedInToday, checkedOutToday, err := uc.countAttendanceToday(ctx, today, nil)
	if err != nil {
		return nil, err
	}

	pendingRequests, err := uc.leaveRequestRepo.Count(ctx, uc.db, ports.LeaveRequestListFilter{
		Status: ptrApprovalStatus(domain.ApprovalRequestStatusPending),
	})
	if err != nil {
		return nil, err
	}

	return &DashboardStatsOutput{
		TotalEmployees:  totalEmployees,
		CheckedInToday:  checkedInToday,
		CheckedOutToday: checkedOutToday,
		LeavesToday:     leavesToday,
		PendingRequests: pendingRequests,
	}, nil
}

func (uc *GetDashboardStatsUseCase) executeScoped(ctx context.Context, departmentUIDs []string) (*DashboardStatsOutput, error) {
	allowedDepartments := make(map[string]struct{}, len(departmentUIDs))
	for _, departmentUID := range departmentUIDs {
		allowedDepartments[departmentUID] = struct{}{}
	}

	employees, err := uc.employeeRepo.List(ctx, uc.db, nil)
	if err != nil {
		return nil, err
	}

	filteredEmployees := make([]*domain.Employee, 0)
	allowedEmployeeUIDs := make(map[string]struct{})
	for _, employee := range employees {
		if employee.DepartmentUID == nil {
			continue
		}
		if _, ok := allowedDepartments[*employee.DepartmentUID]; ok {
			filteredEmployees = append(filteredEmployees, employee)
			allowedEmployeeUIDs[employee.UID] = struct{}{}
		}
	}

	// Keep scoped and unscoped dashboard stats on the same day basis.
	today := time.Now().UTC()
	checkedInToday, checkedOutToday, err := uc.countAttendanceToday(ctx, today, allowedEmployeeUIDs)
	if err != nil {
		return nil, err
	}

	leavesToday := 0
	pendingRequests := 0

	for _, employee := range filteredEmployees {
		onLeave, err := uc.leaveRecordRepo.HasLeaveOnDate(ctx, uc.db, employee.ID, today)
		if err != nil {
			return nil, err
		}
		if onLeave {
			leavesToday++
		}

		employeeUID := employee.UID
		pendingCount, err := uc.leaveRequestRepo.Count(ctx, uc.db, ports.LeaveRequestListFilter{
			Status:      ptrApprovalStatus(domain.ApprovalRequestStatusPending),
			EmployeeUID: &employeeUID,
		})
		if err != nil {
			return nil, err
		}
		pendingRequests += pendingCount
	}

	return &DashboardStatsOutput{
		TotalEmployees:  len(filteredEmployees),
		CheckedInToday:  checkedInToday,
		CheckedOutToday: checkedOutToday,
		LeavesToday:     leavesToday,
		PendingRequests: pendingRequests,
	}, nil
}

func (uc *GetDashboardStatsUseCase) countAttendanceToday(ctx context.Context, date time.Time, allowedEmployeeUIDs map[string]struct{}) (int, int, error) {
	if uc.attendanceRepo == nil {
		return 0, 0, nil
	}

	records, err := uc.attendanceRepo.ListByDate(ctx, uc.db, date, nil)
	if err != nil {
		return 0, 0, err
	}

	checkedInEmployees := make(map[string]struct{})
	checkedOutEmployees := make(map[string]struct{})
	for _, record := range records {
		if allowedEmployeeUIDs != nil {
			if _, ok := allowedEmployeeUIDs[record.EmployeeUID]; !ok {
				continue
			}
		}

		switch record.PunchType {
		case domain.AttendancePunchTypeCheckIn:
			checkedInEmployees[record.EmployeeUID] = struct{}{}
		case domain.AttendancePunchTypeCheckOut:
			checkedOutEmployees[record.EmployeeUID] = struct{}{}
		case domain.AttendancePunchTypeUnknown:
			checkedInEmployees[record.EmployeeUID] = struct{}{}
			checkedOutEmployees[record.EmployeeUID] = struct{}{}
		}
	}

	return len(checkedInEmployees), len(checkedOutEmployees), nil
}

func ptrApprovalStatus(status domain.ApprovalRequestStatus) *domain.ApprovalRequestStatus {
	return &status
}
