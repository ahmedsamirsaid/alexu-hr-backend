package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type ListWeekendDaysOutput struct {
	WeekendDays []int
}

type ListWeekendDaysUseCase struct {
	db          ports.DB
	weekendRepo ports.WeekendConfigRepository
}

func NewListWeekendDaysUseCase(db ports.DB, weekendRepo ports.WeekendConfigRepository) *ListWeekendDaysUseCase {
	return &ListWeekendDaysUseCase{db: db, weekendRepo: weekendRepo}
}

func (uc *ListWeekendDaysUseCase) Execute(ctx context.Context) (*ListWeekendDaysOutput, error) {
	days, err := uc.weekendRepo.GetWeekendDays(ctx, uc.db)
	if err != nil {
		return nil, err
	}
	return &ListWeekendDaysOutput{WeekendDays: days}, nil
}
