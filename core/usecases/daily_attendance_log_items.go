package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

func buildDailyAttendanceLogItems(ctx context.Context, db ports.DB, workHoursRepo ports.WorkHoursConfigRepository, groups []*ports.DailyAttendanceGroup) ([]DailyAttendanceLogItem, error) {
	cfg, err := workHoursRepo.Get(ctx, db)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = domain.NewDefaultWorkHoursConfig()
	}

	records := make([]DailyAttendanceLogItem, 0, len(groups))
	for _, group := range groups {
		item := DailyAttendanceLogItem{
			Date:            group.Date,
			EmployeeUID:     group.EmployeeUID,
			EmployeeName:    group.EmployeeName,
			DepartmentUID:   group.DepartmentUID,
			CheckIn:         group.CheckIn,
			CheckOut:        group.CheckOut,
			CheckInDevice:   group.CheckInDevice,
			CheckOutDevice:  group.CheckOutDevice,
			MissingCheckIn:  group.CheckIn == nil,
			MissingCheckOut: group.CheckOut == nil,
		}

		workStart, err := buildTimeOnDate(group.Date, cfg.WorkDayStart)
		if err != nil {
			return nil, ErrInvalidWorkDayStart
		}
		workEnd, err := buildTimeOnDate(group.Date, cfg.WorkDayEnd)
		if err != nil {
			return nil, ErrInvalidWorkDayEnd
		}

		if group.CheckIn != nil && group.CheckOut != nil && group.CheckOut.After(*group.CheckIn) {
			item.WorkedHours = group.CheckOut.Sub(*group.CheckIn).Hours()
		}
		if group.CheckIn != nil {
			lateThreshold := workStart.Add(time.Duration(cfg.LateGraceMinutes) * time.Minute)
			item.LateArrival = group.CheckIn.After(lateThreshold)
		}
		if group.CheckOut != nil {
			earlyThreshold := workEnd.Add(-time.Duration(cfg.EarlyGraceMinutes) * time.Minute)
			item.EarlyDeparture = group.CheckOut.Before(earlyThreshold)
		}

		records = append(records, item)
	}

	return records, nil
}
