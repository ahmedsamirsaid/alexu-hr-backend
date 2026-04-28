package usecases

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var manualHolidayCodeSanitizer = regexp.MustCompile(`[^A-Z0-9]+`)

type CreateManualHolidayInput struct {
	Date   time.Time
	NameEN string
	NameAR *string
}

type CreateManualHolidayOutput struct {
	UID      string
	Date     time.Time
	NameEN   string
	NameAR   string
	IsManual bool
}

type CreateManualHolidayUseCase struct {
	db             ports.DB
	holidayDefRepo ports.HolidayDefinitionRepository
	weekendRepo    ports.WeekendConfigRepository
	auditor        audit.Auditor
}

func NewCreateManualHolidayUseCase(
	db ports.DB,
	holidayDefRepo ports.HolidayDefinitionRepository,
	weekendRepo ports.WeekendConfigRepository,
	auditor audit.Auditor,
) *CreateManualHolidayUseCase {
	return &CreateManualHolidayUseCase{
		db:             db,
		holidayDefRepo: holidayDefRepo,
		weekendRepo:    weekendRepo,
		auditor:        auditor,
	}
}

func (uc *CreateManualHolidayUseCase) Execute(ctx context.Context, input CreateManualHolidayInput) (*CreateManualHolidayOutput, error) {
	nameEN := strings.TrimSpace(input.NameEN)
	if nameEN == "" {
		nameEN = "Manual Holiday"
	}

	nameAR := nameEN
	if input.NameAR != nil && strings.TrimSpace(*input.NameAR) != "" {
		nameAR = strings.TrimSpace(*input.NameAR)
	}

	dateOnly := normalizeDateOnly(input.Date)
	isWeekend, err := isWeekendDate(ctx, uc.db, uc.weekendRepo, dateOnly)
	if err != nil {
		return nil, err
	}
	if isWeekend {
		return nil, ErrHolidayDateFallsOnWeekend
	}

	existing, err := uc.holidayDefRepo.GetByDate(ctx, uc.db, dateOnly)
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return nil, ErrHolidayDateAlreadyExists
	}

	code := buildManualHolidayCode(nameEN, dateOnly)
	for i := 0; i < 5; i++ {
		def, err := uc.holidayDefRepo.GetByCode(ctx, uc.db, code)
		if err != nil {
			return nil, err
		}
		if def == nil {
			break
		}
		code = fmt.Sprintf("%s_%d", code, i+1)
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	definition := &domain.HolidayDefinition{
		UID:      domain.GenerateUID("hdef"),
		Code:     code,
		NameEN:   nameEN,
		NameAR:   nameAR,
		Date:     dateOnly,
		IsManual: true,
	}
	
	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionSentence := fmt.Sprintf(
		"%s created manual holiday '%s' on %s",
		actorName, nameEN, dateOnly.Format("Jan 2, 2006"),
	)
	
	// Audit log will fire after successful transaction commit
	defer uc.auditor.From(ctx).
		Did(audit.ActionCreate).
		On(audit.EntityHoliday, definition.UID).
		WithMeta("action", actionSentence).
		WithMeta("date", definition.Date.Format("2006-01-02")).
		WithMeta("name_en", definition.NameEN).
		WithMeta("name_ar", definition.NameAR).
		WithMeta("is_manual", true).
		Save(ctx)
	
	if err := uc.holidayDefRepo.Create(ctx, tx, definition); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateManualHolidayOutput{
		UID:      definition.UID,
		Date:     definition.Date,
		NameEN:   definition.NameEN,
		NameAR:   definition.NameAR,
		IsManual: definition.IsManual,
	}, nil
}

func buildManualHolidayCode(name string, dateValue time.Time) string {
	normalized := strings.ToUpper(strings.TrimSpace(name))
	normalized = manualHolidayCodeSanitizer.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		normalized = "MANUAL_HOLIDAY"
	}
	return fmt.Sprintf("MANUAL_%s_%s", normalized, dateValue.Format("20060102"))
}

func isWeekendDate(ctx context.Context, db ports.DB, weekendRepo ports.WeekendConfigRepository, dateValue time.Time) (bool, error) {
	if weekendRepo == nil {
		return false, nil
	}

	weekendDays, err := weekendRepo.GetWeekendDays(ctx, db)
	if err != nil {
		return false, err
	}

	weekendSet := make(map[int]struct{}, len(weekendDays))
	for _, day := range weekendDays {
		weekendSet[day] = struct{}{}
	}

	_, isWeekend := weekendSet[int(dateValue.Weekday())]
	return isWeekend, nil
}
