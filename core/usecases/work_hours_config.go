package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetWorkHoursConfigOutput struct {
	Config *domain.WorkHoursConfig
}

type GetWorkHoursConfigUseCase struct {
	db         ports.DB
	configRepo ports.WorkHoursConfigRepository
}

func NewGetWorkHoursConfigUseCase(
	db ports.DB,
	configRepo ports.WorkHoursConfigRepository,
) *GetWorkHoursConfigUseCase {
	return &GetWorkHoursConfigUseCase{db: db, configRepo: configRepo}
}

func (uc *GetWorkHoursConfigUseCase) Execute(ctx context.Context) (*GetWorkHoursConfigOutput, error) {
	cfg, err := uc.configRepo.Get(ctx, uc.db)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = domain.NewDefaultWorkHoursConfig()
	}
	return &GetWorkHoursConfigOutput{Config: cfg}, nil
}

type SetWorkHoursConfigInput struct {
	WorkDayStart      string
	WorkDayEnd        string
	LateGraceMinutes  int
	EarlyGraceMinutes int
}

type SetWorkHoursConfigUseCase struct {
	db         ports.DB
	configRepo ports.WorkHoursConfigRepository
}

func NewSetWorkHoursConfigUseCase(
	db ports.DB,
	configRepo ports.WorkHoursConfigRepository,
) *SetWorkHoursConfigUseCase {
	return &SetWorkHoursConfigUseCase{db: db, configRepo: configRepo}
}

func (uc *SetWorkHoursConfigUseCase) Execute(ctx context.Context, input SetWorkHoursConfigInput) (*domain.WorkHoursConfig, error) {
	workStart, err := time.Parse("15:04", input.WorkDayStart)
	if err != nil {
		return nil, ErrInvalidWorkDayStart
	}

	workEnd, err := time.Parse("15:04", input.WorkDayEnd)
	if err != nil {
		return nil, ErrInvalidWorkDayEnd
	}

	if !workEnd.After(workStart) {
		return nil, ErrInvalidWorkHoursRange
	}

	if input.LateGraceMinutes < 0 || input.EarlyGraceMinutes < 0 {
		return nil, ErrInvalidGraceMinutes
	}

	cfg := &domain.WorkHoursConfig{
		ID:                1,
		WorkDayStart:      input.WorkDayStart,
		WorkDayEnd:        input.WorkDayEnd,
		LateGraceMinutes:  input.LateGraceMinutes,
		EarlyGraceMinutes: input.EarlyGraceMinutes,
	}

	if err := uc.configRepo.Upsert(ctx, uc.db, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
