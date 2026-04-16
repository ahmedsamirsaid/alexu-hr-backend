package usecases

import (
	"context"
	"errors"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var ErrShiftNotFound = errors.New("shift not found")

type ListShiftsOutput struct {
	Shifts []*domain.Shift
}

type ListShiftsUseCase struct {
	db        ports.DB
	shiftRepo ports.ShiftRepository
}

func NewListShiftsUseCase(db ports.DB, shiftRepo ports.ShiftRepository) *ListShiftsUseCase {
	return &ListShiftsUseCase{db: db, shiftRepo: shiftRepo}
}

func (uc *ListShiftsUseCase) Execute(ctx context.Context) (*ListShiftsOutput, error) {
	shifts, err := uc.shiftRepo.List(ctx, uc.db)
	if err != nil {
		return nil, err
	}
	return &ListShiftsOutput{Shifts: shifts}, nil
}

type GetShiftOutput struct {
	Shift *domain.Shift
}

type GetShiftUseCase struct {
	db        ports.DB
	shiftRepo ports.ShiftRepository
}

func NewGetShiftUseCase(db ports.DB, shiftRepo ports.ShiftRepository) *GetShiftUseCase {
	return &GetShiftUseCase{db: db, shiftRepo: shiftRepo}
}

func (uc *GetShiftUseCase) Execute(ctx context.Context, uid string) (*GetShiftOutput, error) {
	shift, err := uc.shiftRepo.GetByUID(ctx, uc.db, uid)
	if err != nil {
		return nil, err
	}
	if shift == nil {
		return nil, ErrShiftNotFound
	}
	return &GetShiftOutput{Shift: shift}, nil
}

type CreateShiftInput struct {
	StartTime    string
	EndTime      string
	GraceMinutes int
}

type CreateShiftOutput struct {
	Shift *domain.Shift
}

type CreateShiftUseCase struct {
	db        ports.DB
	shiftRepo ports.ShiftRepository
}

func NewCreateShiftUseCase(db ports.DB, shiftRepo ports.ShiftRepository) *CreateShiftUseCase {
	return &CreateShiftUseCase{db: db, shiftRepo: shiftRepo}
}

func (uc *CreateShiftUseCase) Execute(ctx context.Context, input CreateShiftInput) (*CreateShiftOutput, error) {
	if input.GraceMinutes < 0 {
		return nil, ErrInvalidGraceMinutes
	}

	start, err := time.Parse("15:04", input.StartTime)
	if err != nil {
		return nil, ErrInvalidWorkDayStart
	}
	end, err := time.Parse("15:04", input.EndTime)
	if err != nil {
		return nil, ErrInvalidWorkDayEnd
	}
	if !end.After(start) {
		return nil, ErrInvalidWorkHoursRange
	}

	shift := &domain.Shift{
		UID:          domain.GenerateUID("shf"),
		StartTime:    input.StartTime,
		EndTime:      input.EndTime,
		GraceMinutes: input.GraceMinutes,
	}
	if err := uc.shiftRepo.Upsert(ctx, uc.db, shift); err != nil {
		return nil, err
	}

	return &CreateShiftOutput{Shift: shift}, nil
}

type UpdateShiftInput struct {
	UID          string
	StartTime    *string
	EndTime      *string
	GraceMinutes *int
}

type UpdateShiftOutput struct {
	Shift *domain.Shift
}

type UpdateShiftUseCase struct {
	db        ports.DB
	shiftRepo ports.ShiftRepository
}

func NewUpdateShiftUseCase(db ports.DB, shiftRepo ports.ShiftRepository) *UpdateShiftUseCase {
	return &UpdateShiftUseCase{db: db, shiftRepo: shiftRepo}
}

func (uc *UpdateShiftUseCase) Execute(ctx context.Context, input UpdateShiftInput) (*UpdateShiftOutput, error) {
	shift, err := uc.shiftRepo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return nil, err
	}
	if shift == nil {
		return nil, ErrShiftNotFound
	}

	if input.StartTime != nil {
		shift.StartTime = *input.StartTime
	}
	if input.EndTime != nil {
		shift.EndTime = *input.EndTime
	}
	if input.GraceMinutes != nil {
		shift.GraceMinutes = *input.GraceMinutes
	}

	start, err := time.Parse("15:04", shift.StartTime)
	if err != nil {
		return nil, ErrInvalidWorkDayStart
	}
	end, err := time.Parse("15:04", shift.EndTime)
	if err != nil {
		return nil, ErrInvalidWorkDayEnd
	}
	if !end.After(start) {
		return nil, ErrInvalidWorkHoursRange
	}
	if shift.GraceMinutes < 0 {
		return nil, ErrInvalidGraceMinutes
	}

	if err := uc.shiftRepo.Upsert(ctx, uc.db, shift); err != nil {
		return nil, err
	}

	return &UpdateShiftOutput{Shift: shift}, nil
}
