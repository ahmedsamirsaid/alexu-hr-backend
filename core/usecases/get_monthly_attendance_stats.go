package usecases

import (
	"context"
	"sort"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type WorkingHoursByDay struct {
	Date        time.Time
	WorkedHours float64
}

type GetMonthlyAttendanceStatsInput struct {
	EmployeeUID string
}

type GetMonthlyAttendanceStatsOutput struct {
	Month                string
	TotalWorkedHours     float64
	AverageCheckInTime   *time.Time
	MissingCheckOutCount int
	WorkingHoursByDay    []WorkingHoursByDay
}

type GetMonthlyAttendanceStatsUseCase struct {
	db         ports.DB
	recordRepo ports.AttendanceRecordRepository
	now        func() time.Time
}

func NewGetMonthlyAttendanceStatsUseCase(
	db ports.DB,
	recordRepo ports.AttendanceRecordRepository,
) *GetMonthlyAttendanceStatsUseCase {
	return &GetMonthlyAttendanceStatsUseCase{
		db:         db,
		recordRepo: recordRepo,
		now:        time.Now,
	}
}

func (uc *GetMonthlyAttendanceStatsUseCase) Execute(ctx context.Context, input GetMonthlyAttendanceStatsInput) (*GetMonthlyAttendanceStatsOutput, error) {
	current := uc.now()
	monthStart := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
	monthEnd := time.Date(current.Year(), current.Month(), current.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), current.Location())

	employeeUID := input.EmployeeUID
	records, err := uc.recordRepo.ListByDateRange(ctx, uc.db, monthStart, monthEnd, &employeeUID)
	if err != nil {
		return nil, err
	}

	type daySummary struct {
		date       time.Time
		checkIn    *time.Time
		checkOut   *time.Time
		employeeID string
	}

	byEmployeeDay := make(map[string]*daySummary)
	for _, rec := range records {
		day := time.Date(rec.PunchedAt.Year(), rec.PunchedAt.Month(), rec.PunchedAt.Day(), 0, 0, 0, 0, rec.PunchedAt.Location())
		key := rec.EmployeeUID + "|" + day.Format("2006-01-02")

		item, ok := byEmployeeDay[key]
		if !ok {
			item = &daySummary{
				date:       day,
				employeeID: rec.EmployeeUID,
			}
			byEmployeeDay[key] = item
		}

		switch rec.PunchType {
		case domain.AttendancePunchTypeCheckIn:
			if item.checkIn == nil || rec.PunchedAt.Before(*item.checkIn) {
				t := rec.PunchedAt
				item.checkIn = &t
			}
		case domain.AttendancePunchTypeCheckOut:
			if item.checkOut == nil || rec.PunchedAt.After(*item.checkOut) {
				t := rec.PunchedAt
				item.checkOut = &t
			}
		case domain.AttendancePunchTypeUnknown:
			if item.checkIn == nil || rec.PunchedAt.Before(*item.checkIn) {
				t := rec.PunchedAt
				item.checkIn = &t
			}
			if item.checkOut == nil || rec.PunchedAt.After(*item.checkOut) {
				t := rec.PunchedAt
				item.checkOut = &t
			}
		}
	}

	totalWorkedHours := 0.0
	missingCheckOutCount := 0
	totalCheckInSeconds := 0
	checkInCount := 0
	workedHoursByDay := make(map[string]float64)

	for _, item := range byEmployeeDay {
		if item.checkIn != nil {
			totalCheckInSeconds += (item.checkIn.Hour() * 3600) + (item.checkIn.Minute() * 60) + item.checkIn.Second()
			checkInCount++
		}
		if item.checkOut == nil {
			missingCheckOutCount++
		}
		if item.checkIn != nil && item.checkOut != nil && item.checkOut.After(*item.checkIn) {
			worked := item.checkOut.Sub(*item.checkIn).Hours()
			totalWorkedHours += worked
			workedHoursByDay[item.date.Format("2006-01-02")] += worked
		}
	}

	var averageCheckInTime *time.Time
	if checkInCount > 0 {
		avgSeconds := totalCheckInSeconds / checkInCount
		t := time.Date(monthStart.Year(), monthStart.Month(), monthStart.Day(), avgSeconds/3600, (avgSeconds%3600)/60, avgSeconds%60, 0, monthStart.Location())
		averageCheckInTime = &t
	}

	daily := make([]WorkingHoursByDay, 0, current.Day())
	for d := monthStart; !d.After(current); d = d.AddDate(0, 0, 1) {
		day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		daily = append(daily, WorkingHoursByDay{
			Date:        day,
			WorkedHours: workedHoursByDay[day.Format("2006-01-02")],
		})
	}

	sort.Slice(daily, func(i, j int) bool {
		return daily[i].Date.Before(daily[j].Date)
	})

	return &GetMonthlyAttendanceStatsOutput{
		Month:                monthStart.Format("2006-01"),
		TotalWorkedHours:     totalWorkedHours,
		AverageCheckInTime:   averageCheckInTime,
		MissingCheckOutCount: missingCheckOutCount,
		WorkingHoursByDay:    daily,
	}, nil
}
