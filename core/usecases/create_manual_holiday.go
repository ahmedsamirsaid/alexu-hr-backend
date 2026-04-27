package usecases

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var manualHolidayCodeSanitizer = regexp.MustCompile(`[^A-Z0-9]+`)

type CreateManualHolidayInput struct {
	Date           time.Time
	NameEN         string
	NameAR         *string
	DepartmentUIDs []string
}

type CreateManualHolidayOutput struct {
	UID            string
	Date           time.Time
	NameEN         string
	NameAR         string
	DepartmentUIDs []string
	IsManual       bool
}

type CreateManualHolidayUseCase struct {
	db             ports.DB
	holidayDefRepo ports.HolidayDefinitionRepository
	weekendRepo    ports.WeekendConfigRepository
}

func NewCreateManualHolidayUseCase(
	db ports.DB,
	holidayDefRepo ports.HolidayDefinitionRepository,
	weekendRepo ports.WeekendConfigRepository,
) *CreateManualHolidayUseCase {
	return &CreateManualHolidayUseCase{
		db:             db,
		holidayDefRepo: holidayDefRepo,
		weekendRepo:    weekendRepo,
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

	departmentUIDs := normalizeInputDepartmentUIDs(input.DepartmentUIDs)

	existing, err := uc.holidayDefRepo.GetByDate(ctx, uc.db, dateOnly)
	if err != nil {
		return nil, err
	}
	if holidayScopeExists(existing, departmentUIDs) {
		return nil, ErrHolidayDateAlreadyExists
	}

	code := buildManualHolidayCode(nameEN, dateOnly, departmentUIDs)
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
		UID:            domain.GenerateUID("hdef"),
		Code:           code,
		NameEN:         nameEN,
		NameAR:         nameAR,
		Date:           dateOnly,
		DepartmentUIDs: departmentUIDs,
		IsManual:       true,
	}
	if err := uc.holidayDefRepo.Create(ctx, tx, definition); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &CreateManualHolidayOutput{
		UID:            definition.UID,
		Date:           definition.Date,
		NameEN:         definition.NameEN,
		NameAR:         definition.NameAR,
		DepartmentUIDs: append([]string(nil), definition.DepartmentUIDs...),
		IsManual:       definition.IsManual,
	}, nil
}

func buildManualHolidayCode(name string, dateValue time.Time, departmentUIDs []string) string {
	normalized := strings.ToUpper(strings.TrimSpace(name))
	normalized = manualHolidayCodeSanitizer.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		normalized = "MANUAL_HOLIDAY"
	}
	if len(departmentUIDs) == 0 {
		return fmt.Sprintf("MANUAL_%s_%s_GLOBAL", normalized, dateValue.Format("20060102"))
	}

	deptParts := append([]string(nil), departmentUIDs...)
	sort.Strings(deptParts)
	deptCode := manualHolidayCodeSanitizer.ReplaceAllString(strings.ToUpper(strings.Join(deptParts, "_")), "_")
	deptCode = strings.Trim(deptCode, "_")
	if deptCode == "" {
		deptCode = "DEPARTMENT"
	}

	return fmt.Sprintf("MANUAL_%s_%s_%s", normalized, dateValue.Format("20060102"), deptCode)
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

func normalizeInputDepartmentUIDs(departmentUIDs []string) []string {
	if len(departmentUIDs) == 0 {
		return nil
	}

	seen := make(map[string]struct{}, len(departmentUIDs))
	result := make([]string, 0, len(departmentUIDs))
	for _, departmentUID := range departmentUIDs {
		departmentUID = strings.TrimSpace(departmentUID)
		if departmentUID == "" {
			continue
		}
		if _, ok := seen[departmentUID]; ok {
			continue
		}
		seen[departmentUID] = struct{}{}
		result = append(result, departmentUID)
	}
	sort.Strings(result)
	if len(result) == 0 {
		return nil
	}
	return result
}

func holidayScopeExists(existing []*domain.HolidayDefinition, departmentUIDs []string) bool {
	for _, definition := range existing {
		if sameDepartmentScope(definition.DepartmentUIDs, departmentUIDs) {
			return true
		}
	}
	return false
}

func sameDepartmentScope(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	if len(left) == 0 {
		return true
	}

	leftCopy := append([]string(nil), left...)
	rightCopy := append([]string(nil), right...)
	sort.Strings(leftCopy)
	sort.Strings(rightCopy)
	for i := range leftCopy {
		if leftCopy[i] != rightCopy[i] {
			return false
		}
	}
	return true
}
