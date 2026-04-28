package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type UpdateHolidayInput struct {
	UID            string
	Date           time.Time
	NameEN         string
	NameAR         *string
	DepartmentUIDs []string
	IsManual       bool
}

type UpdateHolidayOutput struct {
	UID            string
	Date           time.Time
	NameEN         string
	NameAR         string
	DepartmentUIDs []string
	IsManual       bool
}

type UpdateHolidayUseCase struct {
	db             ports.DB
	holidayDefRepo ports.HolidayDefinitionRepository
	weekendRepo    ports.WeekendConfigRepository
}

func NewUpdateHolidayUseCase(
	db ports.DB,
	holidayDefRepo ports.HolidayDefinitionRepository,
	weekendRepo ports.WeekendConfigRepository,
) *UpdateHolidayUseCase {
	return &UpdateHolidayUseCase{db: db, holidayDefRepo: holidayDefRepo, weekendRepo: weekendRepo}
}

func (uc *UpdateHolidayUseCase) Execute(ctx context.Context, input UpdateHolidayInput) (*UpdateHolidayOutput, error) {
	definitions, err := uc.holidayDefRepo.List(ctx, uc.db)
	if err != nil {
		return nil, err
	}
	definition := findHolidayDefinitionByUID(definitions, input.UID)
	if definition == nil {
		return nil, ErrHolidayNotFound
	}

	nameEN := input.NameEN
	if nameEN == "" {
		nameEN = definition.NameEN
	}

	nameAR := definition.NameAR
	if input.NameAR != nil {
		nameAR = *input.NameAR
	}

	dateOnly := normalizeDateOnly(input.Date)
	departmentUIDs := normalizeInputDepartmentUIDs(input.DepartmentUIDs)
	today := normalizeDateOnly(time.Now())
	if dateOnly.Before(today) {
		return nil, ErrHolidayDateInPast
	}

	existing, err := uc.holidayDefRepo.GetByDate(ctx, uc.db, dateOnly)
	if err != nil {
		return nil, err
	}
	for _, other := range existing {
		if other.ID == definition.ID {
			continue
		}
		if sameDepartmentScope(other.DepartmentUIDs, departmentUIDs) {
			return nil, ErrHolidayDateAlreadyExists
		}
	}

	definition.Date = dateOnly
	definition.NameEN = nameEN
	definition.NameAR = nameAR
	definition.DepartmentUIDs = departmentUIDs

	if err := uc.holidayDefRepo.Update(ctx, uc.db, definition); err != nil {
		return nil, err
	}

	return &UpdateHolidayOutput{
		UID:            definition.UID,
		Date:           definition.Date,
		NameEN:         definition.NameEN,
		NameAR:         definition.NameAR,
		DepartmentUIDs: append([]string(nil), definition.DepartmentUIDs...),
		IsManual:       definition.IsManual,
	}, nil
}

func findHolidayDefinitionByUID(definitions []*domain.HolidayDefinition, uid string) *domain.HolidayDefinition {
	for _, definition := range definitions {
		if definition != nil && definition.UID == uid {
			return definition
		}
	}
	return nil
}
