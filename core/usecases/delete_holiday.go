package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/ports"
)

type DeleteHolidayInput struct {
	UID string
}

type DeleteHolidayUseCase struct {
	db             ports.DB
	holidayDefRepo ports.HolidayDefinitionRepository
}

func NewDeleteHolidayUseCase(db ports.DB, holidayDefRepo ports.HolidayDefinitionRepository) *DeleteHolidayUseCase {
	return &DeleteHolidayUseCase{db: db, holidayDefRepo: holidayDefRepo}
}

func (uc *DeleteHolidayUseCase) Execute(ctx context.Context, input DeleteHolidayInput) error {
	definitions, err := uc.holidayDefRepo.List(ctx, uc.db)
	if err != nil {
		return err
	}
	definition := findHolidayDefinitionByUID(definitions, input.UID)
	if definition == nil {
		return ErrHolidayNotFound
	}

	today := normalizeDateOnly(time.Now())
	if normalizeDateOnly(definition.Date).Before(today) {
		return ErrHolidayDateInPast
	}

	return uc.holidayDefRepo.Delete(ctx, uc.db, definition.ID)
}
