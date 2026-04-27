package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

const (
	leaveTypeCodeCasual  = "CASUAL"
	leaveTypeCodeRegular = "REGULAR"
)

func entitledBalanceTotalDays(employee *domain.Employee, leaveType *domain.LeaveType, asOf time.Time) int {
	if leaveType == nil {
		return 0
	}

	switch leaveType.Code {
	case leaveTypeCodeCasual:
		return 7
	case leaveTypeCodeRegular:
		if employee != nil && completedYearsBetween(employee.HireDate, asOf) > 50 {
			return 45
		}
		if employee != nil && completedYearsBetween(employee.HireDate, asOf) > 10 {
			return 30
		}
		return 21
	default:
		return leaveType.DefaultBalance
	}
}

func ensureLeaveBalance(
	ctx context.Context,
	q ports.Querier,
	leaveBalanceRepo ports.LeaveBalanceRepository,
	employee *domain.Employee,
	leaveType *domain.LeaveType,
	year int,
	asOf time.Time,
) (*domain.LeaveBalance, bool, error) {
	balance, err := leaveBalanceRepo.GetByEmployeeAndTypeAndYear(ctx, q, employee.ID, leaveType.ID, year)
	if err != nil {
		return nil, false, err
	}

	targetTotalDays := entitledBalanceTotalDays(employee, leaveType, asOf)
	if balance == nil {
		balance = domain.NewLeaveBalance(employee.ID, leaveType.ID, year, targetTotalDays)
		if err := leaveBalanceRepo.Create(ctx, q, balance); err != nil {
			return nil, false, err
		}
		return balance, true, nil
	}

	if balance.TotalDays != targetTotalDays {
		balance.TotalDays = targetTotalDays
		if err := leaveBalanceRepo.Update(ctx, q, balance); err != nil {
			return nil, false, err
		}
	}

	return balance, false, nil
}

func completedYearsBetween(start, end time.Time) int {
	if start.IsZero() || end.Before(start) {
		return 0
	}

	years := end.Year() - start.Year()
	if end.Month() < start.Month() || (end.Month() == start.Month() && end.Day() < start.Day()) {
		years--
	}
	if years < 0 {
		return 0
	}

	return years
}
