package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/audit"
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
	auditor        audit.Auditor
}

func NewUpdateHolidayUseCase(
	db ports.DB,
	holidayDefRepo ports.HolidayDefinitionRepository,
	weekendRepo ports.WeekendConfigRepository,
	auditor audit.Auditor,
) *UpdateHolidayUseCase {
	return &UpdateHolidayUseCase{db: db, holidayDefRepo: holidayDefRepo, weekendRepo: weekendRepo, auditor: auditor}
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


	oldDate := definition.Date
	oldNameEN := definition.NameEN
	oldNameAR := definition.NameAR

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

	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionParams := map[string]interface{}{
		"Actor": actorName,
		"Name":  definition.NameEN,
	}

	// Build audit metadata with field changes
	auditBuilder := uc.auditor.From(ctx).
		Did(audit.ActionUpdate).
		On(audit.EntityHoliday, definition.UID).
		WithMeta("action_key", "audit.sentence.update_holiday").
		WithMeta("action_params", actionParams)

	if !oldDate.Equal(dateOnly) {
		auditBuilder.WithMeta("old_date", oldDate.Format("2006-01-02")).
			WithMeta("new_date", dateOnly.Format("2006-01-02"))
	}
	if oldNameEN != nameEN {
		auditBuilder.WithMeta("old_name_en", oldNameEN).
			WithMeta("new_name_en", nameEN)
	}
	if oldNameAR != nameAR {
		auditBuilder.WithMeta("old_name_ar", oldNameAR).
			WithMeta("new_name_ar", nameAR)
	}

	defer auditBuilder.Save(ctx)

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
