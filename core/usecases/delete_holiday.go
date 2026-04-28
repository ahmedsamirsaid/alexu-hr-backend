package usecases

import (
	"context"
	"fmt"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/ports"
)

type DeleteHolidayInput struct {
	UID string
}

type DeleteHolidayUseCase struct {
	db             ports.DB
	holidayDefRepo ports.HolidayDefinitionRepository
	auditor        audit.Auditor
}

func NewDeleteHolidayUseCase(db ports.DB, holidayDefRepo ports.HolidayDefinitionRepository, auditor audit.Auditor) *DeleteHolidayUseCase {
	return &DeleteHolidayUseCase{db: db, holidayDefRepo: holidayDefRepo, auditor: auditor}
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

	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionSentence := fmt.Sprintf(
		"%s deleted holiday '%s' on %s",
		actorName, definition.NameEN, definition.Date.Format("Jan 2, 2006"),
	)

	// Audit log before deletion
	defer uc.auditor.From(ctx).
		Did(audit.ActionDelete).
		On(audit.EntityHoliday, input.UID).
		WithMeta("action", actionSentence).
		WithMeta("date", definition.Date.Format("2006-01-02")).
		WithMeta("name_en", definition.NameEN).
		WithMeta("name_ar", definition.NameAR).
		WithMeta("is_manual", definition.IsManual).
		Save(ctx)

	return uc.holidayDefRepo.Delete(ctx, uc.db, definition.ID)
}
