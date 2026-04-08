package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/ports"
)

type WorkingDaysCalculator struct {
	weekendRepo  ports.WeekendConfigRepository
	holidayRepo  ports.HolidayInstanceRepository
}

func NewWorkingDaysCalculator(weekendRepo ports.WeekendConfigRepository, holidayRepo ports.HolidayInstanceRepository) *WorkingDaysCalculator {
	return &WorkingDaysCalculator{
		weekendRepo:  weekendRepo,
		holidayRepo:  holidayRepo,
	}
}

func (c *WorkingDaysCalculator) CalculateWorkingDays(ctx context.Context, q ports.Querier, start, end time.Time) (int, error) {
	weekendDays, err := c.weekendRepo.GetWeekendDays(ctx, q)
	if err != nil {
		return 0, err
	}

	weekendSet := make(map[int]bool)
	for _, day := range weekendDays {
		weekendSet[day] = true
	}

	holidays, err := c.holidayRepo.ListByDateRange(ctx, q, start, end)
	if err != nil {
		return 0, err
	}

	holidaySet := make(map[string]bool)
	for _, h := range holidays {
		holidaySet[h.ObservedDate.Format("2006-01-02")] = true
	}

	workingDays := 0
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		if weekendSet[int(d.Weekday())] {
			continue
		}

		if holidaySet[d.Format("2006-01-02")] {
			continue
		}

		workingDays++
	}

	return workingDays, nil
}
