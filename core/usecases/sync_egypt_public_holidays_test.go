package usecases

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	dbadapter "github.com/banumusa/backend/adapters/db"
	"github.com/banumusa/backend/core/domain"
)

func TestSyncEgyptPublicHolidaysUseCase_FiltersAndPreservesManual(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()

	defRepo := dbadapter.NewHolidayDefinitionRepository()

	today := normalizeDateOnly(time.Now().UTC())
	upcomingDate := today.AddDate(0, 0, 2)
	dayOffDate := today.AddDate(0, 0, 3)
	manualDate := today.AddDate(0, 0, 4)
	pastDate := today.AddDate(0, 0, -1)
	nextYearDate := time.Date(today.Year()+1, 5, 1, 0, 0, 0, 0, time.UTC)

	manualCode := buildHolidayCode("Manual Keep", manualDate)
	manualDef := &domain.HolidayDefinition{
		UID:      domain.GenerateUID("hdef"),
		Code:     manualCode,
		NameEN:   "Manual Keep",
		NameAR:   "Manual Keep",
		Date:     manualDate,
		IsManual: true,
	}
	if err := defRepo.Create(context.Background(), db, manualDef); err != nil {
		t.Fatalf("failed to seed manual definition: %v", err)
	}

	uid := domain.GenerateUID("emp")
	if _, err := db.ExecContext(context.Background(), `INSERT INTO employees(uid) VALUES (?)`, uid); err != nil {
		t.Fatalf("failed to seed employee: %v", err)
	}
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := db.ExecContext(context.Background(), `
		INSERT INTO attendance_exceptions (uid, employee_uid, attendance_date, exception_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, domain.GenerateUID("aex"), uid, dayOffDate.Format("2006-01-02"), domain.AttendanceExceptionTypeAbsence, now, now)
	if err != nil {
		t.Fatalf("failed to seed attendance exception: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("api_key"); got != "test-api-key" {
			t.Fatalf("expected api_key=test-api-key got %q", got)
		}
		if got := strings.ToUpper(r.URL.Query().Get("country")); got != "EG" {
			t.Fatalf("expected country=EG got %q", got)
		}
		if got := strings.ToLower(r.URL.Query().Get("type")); got != "national" {
			t.Fatalf("expected type=national got %q", got)
		}

		currentYear := today.Year()
		nextYear := currentYear + 1
		year := r.URL.Query().Get("year")

		response := calendarificResponse{}
		switch year {
		case fmt.Sprintf("%d", currentYear):
			response.Response.Holidays = []calendarificHoliday{
				{
					Name:        "Coptic Christmas Day",
					Description: "Coptic Christmas Day is a national holiday in Egypt",
					PrimaryType: "National holiday",
					Type:        []string{"National holiday"},
					URLID:       "egypt/coptic-christmas-day",
					Date: struct {
						ISO string "json:\"iso\""
					}{ISO: upcomingDate.Format("2006-01-02")},
				},
				{
					Name:        "Day off for Coptic Christmas Day",
					Description: "Coptic Christmas Day is a national holiday in Egypt",
					PrimaryType: "National holiday",
					Type:        []string{"National holiday"},
					URLID:       "egypt/coptic-christmas-day",
					Date: struct {
						ISO string "json:\"iso\""
					}{ISO: dayOffDate.Format("2006-01-02")},
				},
				{
					Name:        "Manual Keep",
					Description: "عطلة يدوية",
					PrimaryType: "National holiday",
					Type:        []string{"National holiday"},
					Date: struct {
						ISO string "json:\"iso\""
					}{ISO: manualDate.Format("2006-01-02")},
				},
				{
					Name:        "Revolution Day July 23",
					Description: "Revolution Day July 23 is a national holiday in Egypt",
					PrimaryType: "National holiday",
					Type:        []string{"National holiday"},
					URLID:       "egypt/revolution-day-july-23",
					Date: struct {
						ISO string "json:\"iso\""
					}{ISO: today.AddDate(0, 0, 5).Format("2006-01-02")},
				},
				{
					Name:        "Past Global",
					Description: "عطلة سابقة",
					PrimaryType: "National holiday",
					Type:        []string{"National holiday"},
					Date: struct {
						ISO string "json:\"iso\""
					}{ISO: pastDate.Format("2006-01-02")},
				},
				{
					Name:        "Non Global",
					Description: "Regional",
					PrimaryType: "Regional holiday",
					Type:        []string{"Regional holiday"},
					Date: struct {
						ISO string "json:\"iso\""
					}{ISO: today.AddDate(0, 0, 4).Format("2006-01-02")},
				},
			}
		case fmt.Sprintf("%d", nextYear):
			response.Response.Holidays = []calendarificHoliday{
				{
					Name:        "Labor Day",
					Description: "Labor Day, International Workers' Day, and May Day, is a day off for workers in many countries around the world.",
					PrimaryType: "National holiday",
					Type:        []string{"National holiday"},
					URLID:       "egypt/labor-day",
					Date: struct {
						ISO string "json:\"iso\""
					}{ISO: nextYearDate.Format("2006-01-02")},
				},
			}
		}

		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	uc := NewSyncEgyptPublicHolidaysUseCase(db, defRepo, server.URL, "test-api-key", "EG", "UTC")
	output, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("sync returned error: %v", err)
	}

	if output.CreatedCount != 3 {
		t.Fatalf("expected created=3 got %d", output.CreatedCount)
	}
	if output.DeletedCount != 0 {
		t.Fatalf("expected deleted=0 got %d", output.DeletedCount)
	}
	if output.SkippedPastCount < 1 {
		t.Fatalf("expected skipped past at least 1 got %d", output.SkippedPastCount)
	}
	if output.DeletedAbsenceCount != 1 {
		t.Fatalf("expected deleted absence=1 got %d", output.DeletedAbsenceCount)
	}

	upcomingCode := buildHolidayCode("coptic-christmas-day", dayOffDate)
	upcomingDefinition, err := defRepo.GetByCode(context.Background(), db, upcomingCode)
	if err != nil {
		t.Fatalf("failed to get synced holiday: %v", err)
	}
	if upcomingDefinition == nil {
		t.Fatalf("expected synced upcoming holiday with code %s", upcomingCode)
	}
	if upcomingDefinition.IsManual {
		t.Fatal("expected synced upcoming holiday to be non-manual")
	}
	if upcomingDefinition.Date.Format("2006-01-02") != dayOffDate.Format("2006-01-02") {
		t.Fatalf("expected synced holiday date=%s got %s", dayOffDate.Format("2006-01-02"), upcomingDefinition.Date.Format("2006-01-02"))
	}
	if upcomingDefinition.NameAR != "إجازة عيد الميلاد المجيد" {
		t.Fatalf("expected localized arabic holiday name, got %q", upcomingDefinition.NameAR)
	}

	laborDayCode := buildHolidayCode("labor-day", nextYearDate)
	laborDayDefinition, err := defRepo.GetByCode(context.Background(), db, laborDayCode)
	if err != nil {
		t.Fatalf("failed to get labor day holiday: %v", err)
	}
	if laborDayDefinition == nil {
		t.Fatalf("expected labor day holiday with code %s", laborDayCode)
	}
	if laborDayDefinition.NameAR != "عيد العمال" {
		t.Fatalf("expected labor day arabic name, got %q", laborDayDefinition.NameAR)
	}

	july23Code := buildHolidayCode("revolution-day-july-23", today.AddDate(0, 0, 5))
	july23Definition, err := defRepo.GetByCode(context.Background(), db, july23Code)
	if err != nil {
		t.Fatalf("failed to get july 23 holiday: %v", err)
	}
	if july23Definition == nil {
		t.Fatalf("expected july 23 holiday with code %s", july23Code)
	}
	if july23Definition.NameAR != "ثورة 23 يوليو" {
		t.Fatalf("expected july 23 arabic name, got %q", july23Definition.NameAR)
	}

	manualUpdated, err := defRepo.GetByCode(context.Background(), db, manualCode)
	if err != nil {
		t.Fatalf("failed to get manual holiday: %v", err)
	}
	if manualUpdated == nil || !manualUpdated.IsManual {
		t.Fatalf("manual holiday should stay unchanged, got %+v", manualUpdated)
	}

	var remaining int
	row := db.QueryRowContext(context.Background(), `SELECT COUNT(*) FROM attendance_exceptions WHERE attendance_date = ? AND exception_type = ?`, dayOffDate.Format("2006-01-02"), domain.AttendanceExceptionTypeAbsence)
	if err := row.Scan(&remaining); err != nil {
		t.Fatalf("failed to count attendance exceptions: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected no remaining absence exceptions for upcoming holiday, got %d", remaining)
	}
}

func TestSyncEgyptPublicHolidaysUseCase_ReconcilesUpcomingAutoOnly(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()

	defRepo := dbadapter.NewHolidayDefinitionRepository()

	today := normalizeDateOnly(time.Now().UTC())
	keptFetchedDate := today.AddDate(0, 0, 2)
	missingUpcomingAutoDate := today.AddDate(0, 0, 3)
	manualUpcomingDate := today.AddDate(0, 0, 4)
	pastAutoDate := today.AddDate(0, 0, -2)

	keptFetchedCode := buildHolidayCode("labor-day", keptFetchedDate)
	missingUpcomingAutoCode := buildHolidayCode("Will Be Removed", missingUpcomingAutoDate)
	manualUpcomingCode := buildHolidayCode("Manual Upcoming", manualUpcomingDate)
	pastAutoCode := buildHolidayCode("Past Auto", pastAutoDate)

	seed := []*domain.HolidayDefinition{
		{
			UID:      domain.GenerateUID("hdef"),
			Code:     keptFetchedCode,
			NameEN:   "Kept Old Name",
			NameAR:   "Kept Old Name",
			Date:     keptFetchedDate,
			IsManual: false,
		},
		{
			UID:      domain.GenerateUID("hdef"),
			Code:     missingUpcomingAutoCode,
			NameEN:   "Will Be Removed",
			NameAR:   "Will Be Removed",
			Date:     missingUpcomingAutoDate,
			IsManual: false,
		},
		{
			UID:      domain.GenerateUID("hdef"),
			Code:     manualUpcomingCode,
			NameEN:   "Manual Upcoming",
			NameAR:   "Manual Upcoming",
			Date:     manualUpcomingDate,
			IsManual: true,
		},
		{
			UID:      domain.GenerateUID("hdef"),
			Code:     pastAutoCode,
			NameEN:   "Past Auto",
			NameAR:   "Past Auto",
			Date:     pastAutoDate,
			IsManual: false,
		},
	}

	for _, item := range seed {
		if err := defRepo.Create(context.Background(), db, item); err != nil {
			t.Fatalf("failed to seed holiday %s: %v", item.Code, err)
		}
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := strings.ToLower(r.URL.Query().Get("type")); got != "national" {
			t.Fatalf("expected type=national got %q", got)
		}

		response := calendarificResponse{}
		response.Response.Holidays = []calendarificHoliday{
			{
				Name:        "Labor Day",
				Description: "Labor Day, International Workers' Day, and May Day, is a day off for workers in many countries around the world.",
				PrimaryType: "National holiday",
				Type:        []string{"National holiday"},
				URLID:       "egypt/labor-day",
				Date: struct {
					ISO string "json:\"iso\""
				}{ISO: keptFetchedDate.Format("2006-01-02")},
			},
			{
				Name:        "Manual Upcoming",
				Description: "يدوي",
				PrimaryType: "National holiday",
				Type:        []string{"National holiday"},
				URLID:       "egypt/manual-upcoming",
				Date: struct {
					ISO string "json:\"iso\""
				}{ISO: manualUpcomingDate.Format("2006-01-02")},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	uc := NewSyncEgyptPublicHolidaysUseCase(db, defRepo, server.URL, "test-api-key", "EG", "UTC")
	output, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("sync returned error: %v", err)
	}

	if output.DeletedCount != 1 {
		t.Fatalf("expected deleted=1 got %d", output.DeletedCount)
	}

	keptFetched, err := defRepo.GetByCode(context.Background(), db, keptFetchedCode)
	if err != nil {
		t.Fatalf("failed to load kept fetched code: %v", err)
	}
	if keptFetched == nil {
		t.Fatalf("expected fetched upcoming auto holiday to remain")
	}
	if keptFetched.NameAR != "عيد العمال" {
		t.Fatalf("expected kept fetched holiday to be localized, got %q", keptFetched.NameAR)
	}

	removedAuto, err := defRepo.GetByCode(context.Background(), db, missingUpcomingAutoCode)
	if err != nil {
		t.Fatalf("failed to load removed code: %v", err)
	}
	if removedAuto != nil {
		t.Fatalf("expected missing upcoming auto holiday to be removed")
	}

	manualUpcoming, err := defRepo.GetByCode(context.Background(), db, manualUpcomingCode)
	if err != nil {
		t.Fatalf("failed to load manual code: %v", err)
	}
	if manualUpcoming == nil || !manualUpcoming.IsManual {
		t.Fatalf("expected manual upcoming holiday to remain untouched")
	}

	pastAuto, err := defRepo.GetByCode(context.Background(), db, pastAutoCode)
	if err != nil {
		t.Fatalf("failed to load past auto code: %v", err)
	}
	if pastAuto == nil {
		t.Fatalf("expected past auto holiday to remain untouched")
	}
}
