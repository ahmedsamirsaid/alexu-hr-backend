package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var holidayCodeSanitizer = regexp.MustCompile(`[^A-Z0-9]+`)
var holidayDayOffPrefixSanitizer = regexp.MustCompile(`(?i)^day off for\s+`)

var calendarificHolidayLocalizations = map[string]string{
	"coptic-christmas-day":      "عيد الميلاد المجيد",
	"revolution-day-january-25": "ثورة 25 يناير",
	"eid-al-fitr":               "عيد الفطر",
	"spring-festival":           "عيد شم النسيم",
	"sinai-liberation-day":      "عيد تحرير سيناء",
	"labor-day":                 "عيد العمال",
	"arafat-day":                "يوم عرفة",
	"eid-al-adha":               "عيد الأضحى",
	"muharram":                  "رأس السنة الهجرية",
	"june-30-uprising":          "ثورة 30 يونيو",
	"revolution-day-july-23":    "ثورة 23 يوليو",
	"prophet-birthday":          "المولد النبوي الشريف",
	"armed-forces-day":          "عيد القوات المسلحة",
}

type SyncEgyptPublicHolidaysOutput struct {
	CreatedCount        int
	UpdatedCount        int
	DeletedCount        int
	SkippedPastCount    int
	DeletedAbsenceCount int
}

type syncedHoliday struct {
	Date      string `json:"date"`
	LocalName string `json:"localName"`
	Name      string `json:"name"`
	URLID     string `json:"urlId"`
	Global    bool   `json:"global"`
}

type calendarificResponse struct {
	Response struct {
		Holidays []calendarificHoliday `json:"holidays"`
	} `json:"response"`
}

type calendarificHoliday struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	PrimaryType string   `json:"primary_type"`
	URLID       string   `json:"urlid"`
	Type        []string `json:"type"`
	Date        struct {
		ISO string `json:"iso"`
	} `json:"date"`
}

type SyncEgyptPublicHolidaysUseCase struct {
	db             ports.DB
	holidayDefRepo ports.HolidayDefinitionRepository
	endpoint       string
	apiKey         string
	country        string
	httpClient     *http.Client
	location       *time.Location
}

func NewSyncEgyptPublicHolidaysUseCase(
	db ports.DB,
	holidayDefRepo ports.HolidayDefinitionRepository,
	endpoint string,
	apiKey string,
	country string,
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
		apiKey:         strings.TrimSpace(apiKey),
		country:        strings.TrimSpace(country),
		httpClient:     &http.Client{Timeout: 12 * time.Second},
		location:       loc,
	}
}

func (uc *SyncEgyptPublicHolidaysUseCase) Execute(ctx context.Context) (*SyncEgyptPublicHolidaysOutput, error) {
	if strings.TrimSpace(uc.endpoint) == "" || uc.apiKey == "" {
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

		codeSeed := row.Name
		if strings.TrimSpace(row.URLID) != "" {
			codeSeed = stripCalendarificCountryPrefix(row.URLID)
		}
		holidayCode := buildHolidayCode(codeSeed, dateValue)
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

func (uc *SyncEgyptPublicHolidaysUseCase) fetchHolidays(ctx context.Context) ([]syncedHoliday, error) {
	now := time.Now().In(uc.location)
	years := []int{now.Year(), now.Year() + 1}
	rows := make([]syncedHoliday, 0)

	for _, year := range years {
		yearRows, err := uc.fetchHolidaysForYear(ctx, year)
		if err != nil {
			return nil, err
		}
		rows = append(rows, yearRows...)
	}

	return rows, nil
}

func (uc *SyncEgyptPublicHolidaysUseCase) fetchHolidaysForYear(ctx context.Context, year int) ([]syncedHoliday, error) {
	requestURL, err := uc.buildCalendarificURL(year)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
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

	var payload calendarificResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	dayOffExistsByIdentity := make(map[string]struct{})
	for _, holiday := range payload.Response.Holidays {
		if !isNationalHoliday(holiday) {
			continue
		}
		if isDayOffHolidayName(holiday.Name) {
			identity := canonicalHolidayIdentity(holiday.Name)
			if identity != "" {
				dayOffExistsByIdentity[identity] = struct{}{}
			}
		}
	}

	rows := make([]syncedHoliday, 0, len(payload.Response.Holidays))
	for _, holiday := range payload.Response.Holidays {
		isGlobal := isNationalHoliday(holiday)
		if !isGlobal {
			continue
		}

		identity := canonicalHolidayIdentity(holiday.Name)
		if _, exists := dayOffExistsByIdentity[identity]; exists && !isDayOffHolidayName(holiday.Name) {
			// Prefer the observed "Day off for ..." holiday entry over the base holiday entry.
			continue
		}

		rows = append(rows, syncedHoliday{
			Date:      strings.TrimSpace(holiday.Date.ISO),
			Name:      strings.TrimSpace(holiday.Name),
			LocalName: localizeCalendarificHoliday(holiday),
			URLID:     strings.TrimSpace(holiday.URLID),
			Global:    true,
		})
	}

	return rows, nil
}

func (uc *SyncEgyptPublicHolidaysUseCase) buildCalendarificURL(year int) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(uc.endpoint))
	if err != nil {
		return "", err
	}

	query := parsed.Query()
	query.Set("api_key", uc.apiKey)
	country := uc.country
	if country == "" {
		country = "EG"
	}
	query.Set("country", country)
	query.Set("year", fmt.Sprintf("%d", year))
	query.Set("type", "national")
	parsed.RawQuery = query.Encode()

	return parsed.String(), nil
}

func isNationalHoliday(holiday calendarificHoliday) bool {
	if strings.EqualFold(strings.TrimSpace(holiday.PrimaryType), "National holiday") {
		return true
	}
	for _, holidayType := range holiday.Type {
		if strings.EqualFold(strings.TrimSpace(holidayType), "National holiday") {
			return true
		}
	}
	return false
}

func isDayOffHolidayName(name string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(name)), "day off for ")
}

func canonicalHolidayIdentity(name string) string {
	canonical := holidayDayOffPrefixSanitizer.ReplaceAllString(strings.TrimSpace(name), "")
	return strings.ToLower(strings.TrimSpace(canonical))
}

func stripCalendarificCountryPrefix(urlID string) string {
	trimmed := strings.TrimSpace(urlID)
	if idx := strings.Index(trimmed, "/"); idx >= 0 && idx+1 < len(trimmed) {
		trimmed = trimmed[idx+1:]
	}
	return trimmed
}

func localizeCalendarificHoliday(holiday calendarificHoliday) string {
	baseKey := calendarificHolidayLocalizationKey(holiday)
	baseName := calendarificHolidayLocalizations[baseKey]
	if baseName == "" {
		return ""
	}

	normalizedName := strings.ToLower(strings.TrimSpace(holiday.Name))
	switch {
	case strings.HasPrefix(normalizedName, "day off for "):
		return "إجازة " + baseName
	case strings.HasSuffix(normalizedName, " holiday"):
		return "عطلة " + baseName
	default:
		return baseName
	}
}

func calendarificHolidayLocalizationKey(holiday calendarificHoliday) string {
	baseKey := stripCalendarificCountryPrefix(holiday.URLID)
	if baseKey != "" {
		if _, exists := calendarificHolidayLocalizations[baseKey]; exists {
			return baseKey
		}
		if idx := strings.LastIndex(baseKey, "-"); idx >= 0 {
			suffix := baseKey[idx+1:]
			if suffix != "" {
				if _, err := strconv.Atoi(suffix); err == nil {
					trimmedBase := baseKey[:idx]
					if _, exists := calendarificHolidayLocalizations[trimmedBase]; exists {
						return trimmedBase
					}
				}
			}
		}
	}

	if slugFromName := calendarificHolidaySlugFromName(holiday.Name); slugFromName != "" {
		if _, exists := calendarificHolidayLocalizations[slugFromName]; exists {
			return slugFromName
		}
	}

	return baseKey
}

func calendarificHolidaySlugFromName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = holidayDayOffPrefixSanitizer.ReplaceAllString(normalized, "")
	normalized = strings.TrimSuffix(normalized, " holiday")
	normalized = strings.TrimSpace(normalized)
	normalized = strings.ReplaceAll(normalized, "'", "")
	normalized = strings.ReplaceAll(normalized, ".", "")
	normalized = strings.ReplaceAll(normalized, "/", "-")
	normalized = strings.ReplaceAll(normalized, "&", "and")
	normalized = strings.ReplaceAll(normalized, " ", "-")
	for strings.Contains(normalized, "--") {
		normalized = strings.ReplaceAll(normalized, "--", "-")
	}
	return strings.Trim(normalized, "-")
}

func (uc *SyncEgyptPublicHolidaysUseCase) ensureDefinition(ctx context.Context, q ports.Querier, code string, row syncedHoliday, dateValue time.Time) (*domain.HolidayDefinition, bool, error) {
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
