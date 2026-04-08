package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/ports"
)

// DashboardStatsOutput contains the dashboard statistics.
type DashboardStatsOutput struct {
	TotalEmployees  int
	LeavesToday     int
	PendingRequests int
}

// GetDashboardStatsUseCase handles retrieving dashboard statistics.
type GetDashboardStatsUseCase struct {
	db              ports.DB
	employeeRepo    ports.EmployeeRepository
	leaveRecordRepo ports.LeaveRecordRepository
}

// NewGetDashboardStatsUseCase creates a new get dashboard stats use case.
func NewGetDashboardStatsUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	leaveRecordRepo ports.LeaveRecordRepository,
) *GetDashboardStatsUseCase {
	return &GetDashboardStatsUseCase{
		db:              db,
		employeeRepo:    employeeRepo,
		leaveRecordRepo: leaveRecordRepo,
	}
}

// Execute retrieves dashboard statistics.
func (uc *GetDashboardStatsUseCase) Execute(ctx context.Context) (*DashboardStatsOutput, error) {
	totalEmployees, err := uc.employeeRepo.Count(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	today := time.Now()
	leavesToday, err := uc.leaveRecordRepo.CountOnLeaveToday(ctx, uc.db, today)
	if err != nil {
		return nil, err
	}

	return &DashboardStatsOutput{
		TotalEmployees:  totalEmployees,
		LeavesToday:     leavesToday,
		PendingRequests: 0, // Placeholder for future approval workflow
	}, nil
}
