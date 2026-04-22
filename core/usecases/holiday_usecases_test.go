package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	dbadapter "github.com/banumusa/backend/adapters/db"
)

func TestCreateManualHolidayUseCase_RejectsExistingHolidayDate(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()

	defRepo := dbadapter.NewHolidayDefinitionRepository()
	uc := NewCreateManualHolidayUseCase(db, defRepo, nil)

	dateValue := nextWeekdayDate(time.Now(), time.Monday)
	_, err := uc.Execute(context.Background(), CreateManualHolidayInput{
		Date:   dateValue,
		NameEN: "Company Day",
	})
	if err != nil {
		t.Fatalf("first create returned error: %v", err)
	}

	_, err = uc.Execute(context.Background(), CreateManualHolidayInput{
		Date:   dateValue,
		NameEN: "Another Holiday",
	})
	if !errors.Is(err, ErrHolidayDateAlreadyExists) {
		t.Fatalf("expected ErrHolidayDateAlreadyExists, got %v", err)
	}
}

func TestCreateManualHolidayUseCase_RejectsWeekendDate(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()
	seedWeekendDays(t, db, int(time.Friday), int(time.Saturday))

	defRepo := dbadapter.NewHolidayDefinitionRepository()
	weekendRepo := dbadapter.NewWeekendConfigRepository()
	uc := NewCreateManualHolidayUseCase(db, defRepo, weekendRepo)

	friday := nextWeekdayDate(time.Now(), time.Friday)
	_, err := uc.Execute(context.Background(), CreateManualHolidayInput{
		Date:   friday,
		NameEN: "Friday Holiday",
	})
	if !errors.Is(err, ErrHolidayDateFallsOnWeekend) {
		t.Fatalf("expected ErrHolidayDateFallsOnWeekend, got %v", err)
	}
}

func TestListHolidaysUseCase_ReturnsNamesAndMetadata(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()

	defRepo := dbadapter.NewHolidayDefinitionRepository()
	createUC := NewCreateManualHolidayUseCase(db, defRepo, nil)
	listUC := NewListHolidaysUseCase(db, defRepo)

	dateValue := nextWeekdayDate(time.Now(), time.Monday)
	created, err := createUC.Execute(context.Background(), CreateManualHolidayInput{
		Date:   dateValue,
		NameEN: "Bridge Holiday",
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	start := dateValue.AddDate(0, 0, -1)
	end := dateValue.AddDate(0, 0, 1)
	output, err := listUC.Execute(context.Background(), ListHolidaysInput{StartDate: &start, EndDate: &end})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	if len(output.Holidays) != 1 {
		t.Fatalf("expected 1 holiday, got %d", len(output.Holidays))
	}
	if output.Holidays[0].UID != created.UID {
		t.Fatalf("unexpected uid: got %s want %s", output.Holidays[0].UID, created.UID)
	}
	if output.Holidays[0].NameEN != "Bridge Holiday" {
		t.Fatalf("unexpected nameEn: %s", output.Holidays[0].NameEN)
	}
	if !output.Holidays[0].IsManual {
		t.Fatalf("unexpected metadata: manual=%v", output.Holidays[0].IsManual)
	}
}

func TestNonWorkingDateChecker_FallbackWeekendAndHoliday(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()
	seedWeekendDays(t, db, int(time.Friday), int(time.Saturday))

	defRepo := dbadapter.NewHolidayDefinitionRepository()
	weekendRepo := dbadapter.NewWeekendConfigRepository()
	createUC := NewCreateManualHolidayUseCase(db, defRepo, weekendRepo)

	holidayDate := nextWeekdayDate(time.Now(), time.Monday)
	if _, err := createUC.Execute(context.Background(), CreateManualHolidayInput{Date: holidayDate, NameEN: "Test Holiday"}); err != nil {
		t.Fatalf("failed to create holiday: %v", err)
	}

	start := holidayDate.AddDate(0, 0, -1)
	end := holidayDate.AddDate(0, 0, 1)
	checker, err := newNonWorkingDateChecker(context.Background(), db, weekendRepo, defRepo, &start, &end)
	if err != nil {
		t.Fatalf("failed to create checker: %v", err)
	}

	nonWorkingHoliday, err := checker.IsNonWorking(holidayDate)
	if err != nil {
		t.Fatalf("checker holiday error: %v", err)
	}
	if !nonWorkingHoliday {
		t.Fatal("expected holiday date to be non-working")
	}

	friday := nextWeekdayDate(time.Now(), time.Friday)
	nonWorkingWeekend, err := checker.IsNonWorking(friday)
	if err != nil {
		t.Fatalf("checker weekend error: %v", err)
	}
	if !nonWorkingWeekend {
		t.Fatal("expected friday to be non-working")
	}

	saturday := nextWeekdayDate(time.Now(), time.Saturday)
	nonWorkingSaturday, err := checker.IsNonWorking(saturday)
	if err != nil {
		t.Fatalf("checker saturday error: %v", err)
	}
	if !nonWorkingSaturday {
		t.Fatal("expected saturday to be non-working")
	}
}

func nextWeekdayDate(from time.Time, target time.Weekday) time.Time {
	dateValue := normalizeDateOnly(from)
	for dateValue.Weekday() != target {
		dateValue = dateValue.AddDate(0, 0, 1)
	}
	return dateValue
}
