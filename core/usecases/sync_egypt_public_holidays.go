package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var holidayCodeSanitizer = regexp.MustCompile(`[^A-Z0-9]+`)

type SyncEgyptPublicHolidaysOutput struct {
	CreatedCount        int
	UpdatedCount        int
	DeletedCount        int
	SkippedPastCount    int
	DeletedAbsenceCount int
}

type nagerHoliday struct {
	Date      string `json:"date"`
	LocalName string `json:"localName"`
	Name      string `json:"name"`
	Global    bool   `json:"global"`
}

type SyncEgyptPublicHolidaysUseCase struct {
	db             ports.DB
	holidayDefRepo ports.HolidayDefinitionRepository
	endpoint       string
	httpClient     *http.Client
	location       *time.Location
}

func NewSyncEgyptPublicHolidaysUseCase(
	db ports.DB,
	holidayDefRepo ports.HolidayDefinitionRepository,
	endpoint string,
	timezone string,
) *SyncEgyptPublicHolidaysUseCase {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.FixedZone("Africa/Cairo", 2*60*60)
	}

	return &SyncEgyptPublicHolidaysUseCase{
		db:             db,
		holidayDefRepo: holidayDefRepo,
		endpoint:       endpoint,
		httpClient:     &http.Client{Timeout: 12 * time.Second},
		location:       loc,
	}
}

func (uc *SyncEgyptPublicHolidaysUseCase) Execute(ctx context.Context) (*SyncEgyptPublicHolidaysOutput, error) {
	if strings.TrimSpace(uc.endpoint) == "" {
		return &SyncEgyptPublicHolidaysOutput{}, nil
	}

	now := time.Now().In(uc.location)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, uc.location)
	output := &SyncEgyptPublicHolidaysOutput{}
	futureHolidayDates := make(map[string]struct{})
	fetchedUpcomingCodes := make(map[string]struct{})

	holidayRows, err := uc.fetchHolidays(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	for _, row := range holidayRows {
		if !row.Global {
			continue
		}

		dateValue, err := time.ParseInLocation("2006-01-02", row.Date, uc.location)
		if err != nil {
			slog.Warn("sync_egypt_public_holidays.Execute.skip_invalid_date", "date", row.Date, "error", err)
			continue
		}

		if dateValue.Before(today) {
			output.SkippedPastCount++
			continue
		}

		holidayCode := buildHolidayCode(row.Name, dateValue)
		fetchedUpcomingCodes[holidayCode] = struct{}{}
		definition, created, err := uc.ensureDefinition(ctx, tx, holidayCode, row, dateValue)
		if err != nil {
			return nil, err
		}

		if created {
			output.CreatedCount++
		} else {
			if definition.IsManual {
				continue
			}
			if definition.Date.Before(today) {
				continue
			}
			output.UpdatedCount++
		}

		futureHolidayDates[dateValue.Format("2006-01-02")] = struct{}{}
	}

	deletedCount, err := deleteAbsenceExceptionsByDates(ctx, tx, futureHolidayDates)
	if err != nil {
		return nil, err
	}
	output.DeletedAbsenceCount = deletedCount

	deletedHolidayCount, err := deleteMissingUpcomingAutoHolidays(ctx, tx, uc.holidayDefRepo, today, fetchedUpcomingCodes)
	if err != nil {
		return nil, err
	}
	output.DeletedCount = deletedHolidayCount

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return output, nil
}

func deleteMissingUpcomingAutoHolidays(
	ctx context.Context,
	q ports.Querier,
	holidayDefRepo ports.HolidayDefinitionRepository,
	today time.Time,
	fetchedUpcomingCodes map[string]struct{},
) (int, error) {
	definitions, err := holidayDefRepo.List(ctx, q)
	if err != nil {
		return 0, err
	}

	deleted := 0
	for _, definition := range definitions {
		if definition == nil || definition.IsManual {
			continue
		}
		if definition.Date.Before(today) {
			continue
		}
		if _, exists := fetchedUpcomingCodes[definition.Code]; exists {
			continue
		}

		result, err := q.ExecContext(ctx, `DELETE FROM holiday_definitions WHERE id = ?`, definition.ID)
		if err != nil {
			return 0, err
		}

		affected, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		deleted += int(affected)
	}

	return deleted, nil
}

func (uc *SyncEgyptPublicHolidaysUseCase) fetchHolidays(ctx context.Context) ([]nagerHoliday, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, uc.endpoint, nil)
	if err != nil {
		return nil, err
	}

	resp, err := uc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return nil, fmt.Errorf("holiday sync failed: status=%d body=%s", resp.StatusCode, string(body))
	}

	var rows []nagerHoliday
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (uc *SyncEgyptPublicHolidaysUseCase) ensureDefinition(ctx context.Context, q ports.Querier, code string, row nagerHoliday, dateValue time.Time) (*domain.HolidayDefinition, bool, error) {
	definition, err := uc.holidayDefRepo.GetByCode(ctx, q, code)
	if err != nil {
		return nil, false, err
	}
	if definition != nil {
		if !definition.IsManual {
			definition.Date = dateValue
			definition.NameEN = strings.TrimSpace(row.Name)
			definition.NameAR = strings.TrimSpace(row.LocalName)
			if definition.NameEN == "" {
				definition.NameEN = code
			}
			if definition.NameAR == "" {
				definition.NameAR = definition.NameEN
			}
			updateQuery := `
				UPDATE holiday_definitions
				SET name_en = ?, name_ar = ?, date = ?, is_manual = ?, updated_at = ?
				WHERE id = ?`
			if _, err := q.ExecContext(ctx, updateQuery, definition.NameEN, definition.NameAR, definition.Date, definition.IsManual, time.Now(), definition.ID); err != nil {
				return nil, false, err
			}
		}

		return definition, false, nil
	}

	definition = &domain.HolidayDefinition{
		UID:      domain.GenerateUID("hdef"),
		Code:     code,
		NameEN:   strings.TrimSpace(row.Name),
		NameAR:   strings.TrimSpace(row.LocalName),
		Date:     dateValue,
		IsManual: false,
	}
	if definition.NameEN == "" {
		definition.NameEN = code
	}
	if definition.NameAR == "" {
		definition.NameAR = definition.NameEN
	}

	if err := uc.holidayDefRepo.Create(ctx, q, definition); err != nil {
		return nil, false, err
	}

	return definition, true, nil
}

func buildHolidayCode(name string, dateValue time.Time) string {
	normalized := strings.ToUpper(strings.TrimSpace(name))
	normalized = holidayCodeSanitizer.ReplaceAllString(normalized, "_")
	normalized = strings.Trim(normalized, "_")
	if normalized == "" {
		normalized = "PUBLIC_HOLIDAY"
	}
	return fmt.Sprintf("NAGER_%s_%d", normalized, dateValue.Year())
}

func deleteAbsenceExceptionsByDates(ctx context.Context, q ports.Querier, dateSet map[string]struct{}) (int, error) {
	if len(dateSet) == 0 {
		return 0, nil
	}

	placeholders := make([]string, 0, len(dateSet))
	args := make([]any, 0, len(dateSet)+1)
	args = append(args, domain.AttendanceExceptionTypeAbsence)
	for dateValue := range dateSet {
		placeholders = append(placeholders, "?")
		args = append(args, dateValue)
	}

	query := fmt.Sprintf(
		"DELETE FROM attendance_exceptions WHERE exception_type = ? AND attendance_date IN (%s)",
		strings.Join(placeholders, ","),
	)

	result, err := q.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(affected), nil
}
