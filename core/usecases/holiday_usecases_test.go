package usecases

import (
	"context"
	"errors"
	"testing"
	"time"

	dbadapter "github.com/banumusa/backend/adapters/db"
	"github.com/banumusa/backend/core/domain"
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

func TestCreateManualHolidayUseCase_AllowsWeekendDate(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()
	seedWeekendDays(t, db, int(time.Friday), int(time.Saturday))

	defRepo := dbadapter.NewHolidayDefinitionRepository()
	weekendRepo := dbadapter.NewWeekendConfigRepository()
	uc := NewCreateManualHolidayUseCase(db, defRepo, weekendRepo)

	friday := nextWeekdayDate(time.Now(), time.Friday)
	created, err := uc.Execute(context.Background(), CreateManualHolidayInput{
		Date:   friday,
		NameEN: "Friday Holiday",
	})
	if err != nil {
		t.Fatalf("expected weekend create to succeed, got %v", err)
	}
	if created == nil {
		t.Fatal("expected created holiday, got nil")
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

func TestCreateManualHolidayUseCase_AllowsDepartmentScopedHolidayOnSameDateAsGlobal(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()

	defRepo := dbadapter.NewHolidayDefinitionRepository()
	createUC := NewCreateManualHolidayUseCase(db, defRepo, nil)
	listUC := NewListHolidaysUseCase(db, defRepo)

	dateValue := nextWeekdayDate(time.Now(), time.Monday)
	if _, err := createUC.Execute(context.Background(), CreateManualHolidayInput{
		Date:   dateValue,
		NameEN: "Company Day",
	}); err != nil {
		t.Fatalf("global create returned error: %v", err)
	}

	departmentHoliday, err := createUC.Execute(context.Background(), CreateManualHolidayInput{
		Date:           dateValue,
		NameEN:         "Finance Day",
		DepartmentUIDs: []string{"dept-finance"},
	})
	if err != nil {
		t.Fatalf("department-scoped create returned error: %v", err)
	}
	if len(departmentHoliday.DepartmentUIDs) != 1 || departmentHoliday.DepartmentUIDs[0] != "dept-finance" {
		t.Fatalf("unexpected scoped departments: %+v", departmentHoliday.DepartmentUIDs)
	}

	start := dateValue.AddDate(0, 0, -1)
	end := dateValue.AddDate(0, 0, 1)
	output, err := listUC.Execute(context.Background(), ListHolidaysInput{StartDate: &start, EndDate: &end})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	if len(output.Holidays) != 2 {
		t.Fatalf("expected 2 holidays, got %d", len(output.Holidays))
	}
}

func TestUpdateHolidayUseCase_RejectsPastDate(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()

	defRepo := dbadapter.NewHolidayDefinitionRepository()
	createUC := NewCreateManualHolidayUseCase(db, defRepo, nil, nil)
	updateUC := NewUpdateHolidayUseCase(db, defRepo, nil, nil)

	futureDate := normalizeDateOnly(time.Now().AddDate(0, 0, 2))
	created, err := createUC.Execute(context.Background(), CreateManualHolidayInput{
		Date:   futureDate,
		NameEN: "Editable Holiday",
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	pastDate := normalizeDateOnly(time.Now().AddDate(0, 0, -1))
	_, err = updateUC.Execute(context.Background(), UpdateHolidayInput{
		UID:  created.UID,
		Date: pastDate,
	})
	if !errors.Is(err, ErrHolidayDateInPast) {
		t.Fatalf("expected ErrHolidayDateInPast, got %v", err)
	}
}

func TestDeleteHolidayUseCase_RejectsPastDate(t *testing.T) {
	db := newUsecaseTestSQLiteDB(t)
	defer db.Close()

	defRepo := dbadapter.NewHolidayDefinitionRepository()
	deleteUC := NewDeleteHolidayUseCase(db, defRepo, nil)

	pastDate := normalizeDateOnly(time.Now().AddDate(0, 0, -1))
	pastHoliday := &domain.HolidayDefinition{
		UID:      "hdef_past_1",
		Code:     "PAST_HOLIDAY",
		NameEN:   "Past Holiday",
		NameAR:   "Past Holiday",
		Date:     pastDate,
		IsManual: true,
	}
	if err := defRepo.Create(context.Background(), db, pastHoliday); err != nil {
		t.Fatalf("failed to create past holiday: %v", err)
	}

	err := deleteUC.Execute(context.Background(), DeleteHolidayInput{UID: pastHoliday.UID})
	if !errors.Is(err, ErrHolidayDateInPast) {
		t.Fatalf("expected ErrHolidayDateInPast, got %v", err)
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
