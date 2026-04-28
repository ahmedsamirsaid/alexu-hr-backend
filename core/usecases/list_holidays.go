package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/ports"
)

type HolidayListItem struct {
	UID            string
	Date           time.Time
	NameEN         string
	NameAR         string
	DepartmentUIDs []string
	IsManual       bool
}

type ListHolidaysInput struct {
	StartDate *time.Time
	EndDate   *time.Time
}

type ListHolidaysOutput struct {
	Holidays []HolidayListItem
}

type ListHolidaysUseCase struct {
	db             ports.DB
	holidayDefRepo ports.HolidayDefinitionRepository
}

func NewListHolidaysUseCase(
	db ports.DB,
	holidayDefRepo ports.HolidayDefinitionRepository,
) *ListHolidaysUseCase {
	return &ListHolidaysUseCase{
		db:             db,
		holidayDefRepo: holidayDefRepo,
	}
}

func (uc *ListHolidaysUseCase) Execute(ctx context.Context, input ListHolidaysInput) (*ListHolidaysOutput, error) {
	now := time.Now()
	start := normalizeDateOnly(now.AddDate(-1, 0, 0))
	end := normalizeDateOnly(now.AddDate(1, 0, 0)).Add(24*time.Hour - time.Nanosecond)

	if input.StartDate != nil {
		start = normalizeDateOnly(*input.StartDate)
	}
	if input.EndDate != nil {
		end = normalizeDateOnly(*input.EndDate).Add(24*time.Hour - time.Nanosecond)
	}
	if end.Before(start) {
		start, end = end, start
	}

	definitions, err := uc.holidayDefRepo.ListAllByDateRange(ctx, uc.db, start, end)
	if err != nil {
		return nil, err
	}

	items := make([]HolidayListItem, 0, len(definitions))
	for _, definition := range definitions {
		items = append(items, HolidayListItem{
			UID:            definition.UID,
			Date:           definition.Date,
			NameEN:         definition.NameEN,
			NameAR:         definition.NameAR,
			DepartmentUIDs: append([]string(nil), definition.DepartmentUIDs...),
			IsManual:       definition.IsManual,
		})
	}

	return &ListHolidaysOutput{Holidays: items}, nil
}
