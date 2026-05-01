package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

func calculateLeaveDays(
	ctx context.Context,
	q ports.Querier,
	workingDaysCalc *WorkingDaysCalculator,
	leaveType *domain.LeaveType,
	startDate time.Time,
	endDate time.Time,
) (int, error) {
	if leaveType != nil {
		switch leaveType.Code {
		case leaveTypeCodeCasual, leaveTypeCodeRegular:
			return workingDaysCalc.CalculateWorkingDays(ctx, q, startDate, endDate)
		}
	}

	return calendarDaysInclusive(startDate, endDate), nil
}

func calendarDaysInclusive(startDate, endDate time.Time) int {
	if endDate.Before(startDate) {
		return 0
	}

	start := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, startDate.Location())
	end := time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 0, 0, 0, 0, endDate.Location())

	return int(end.Sub(start).Hours()/24) + 1
}
