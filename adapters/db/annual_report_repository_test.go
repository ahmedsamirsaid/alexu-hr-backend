package db

import (
	"context"
	"testing"

	"github.com/banumusa/backend/core/domain"
)

func TestAnnualReportRepository_CreateAndListByEmployeeUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewAnnualReportRepository()
	ctx := context.Background()
	employee := tdb.SeedEmployee("Report Employee")

	first := &domain.AnnualReport{
		EmployeeUID:    employee.UID,
		ReportYear:     "2024",
		ReportGrade:    "A",
		Notes:          "Strong performance",
		ReportImageURL: "minio://documents/reports/r1.pdf",
	}
	second := &domain.AnnualReport{
		EmployeeUID:    employee.UID,
		ReportYear:     "2025",
		ReportGrade:    "B+",
		Notes:          "Good performance",
		ReportImageURL: "minio://documents/reports/r2.pdf",
	}

	if err := repo.Create(ctx, tdb.SQLiteDB, first); err != nil {
		t.Fatalf("Create() first error = %v", err)
	}
	if err := repo.Create(ctx, tdb.SQLiteDB, second); err != nil {
		t.Fatalf("Create() second error = %v", err)
	}

	reports, err := repo.ListByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID)
	if err != nil {
		t.Fatalf("ListByEmployeeUID() error = %v", err)
	}
	if len(reports) != 2 {
		t.Fatalf("ListByEmployeeUID() len = %d, want 2", len(reports))
	}
	if reports[0].ReportYear != "2025" || reports[1].ReportYear != "2024" {
		t.Fatalf("unexpected report ordering: %+v", reports)
	}
}

func TestAnnualReportRepository_DeleteByEmployeeUID(t *testing.T) {
	tdb := NewTestDB(t)
	defer tdb.Close()

	repo := NewAnnualReportRepository()
	ctx := context.Background()
	employee := tdb.SeedEmployee("Report Employee")

	if err := repo.Create(ctx, tdb.SQLiteDB, &domain.AnnualReport{
		EmployeeUID:    employee.UID,
		ReportYear:     "2025",
		ReportGrade:    "A",
		Notes:          "Strong",
		ReportImageURL: "minio://documents/reports/r1.pdf",
	}); err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.DeleteByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID); err != nil {
		t.Fatalf("DeleteByEmployeeUID() error = %v", err)
	}

	reports, err := repo.ListByEmployeeUID(ctx, tdb.SQLiteDB, employee.UID)
	if err != nil {
		t.Fatalf("ListByEmployeeUID() error = %v", err)
	}
	if len(reports) != 0 {
		t.Fatalf("ListByEmployeeUID() len = %d, want 0", len(reports))
	}
}
