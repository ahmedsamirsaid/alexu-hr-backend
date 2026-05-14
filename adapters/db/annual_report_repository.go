package db

import (
	"context"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type AnnualReportRepository struct{}

func NewAnnualReportRepository() *AnnualReportRepository {
	return &AnnualReportRepository{}
}

func (r *AnnualReportRepository) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) ([]*domain.AnnualReport, error) {
	query := `
		SELECT id, employee_uid, report_year, report_grade, notes, report_image_url, created_at, updated_at
		FROM annual_reports
		WHERE employee_uid = $1
		ORDER BY report_year DESC, id DESC`

	rows, err := q.QueryContext(ctx, query, employeeUID)
	if err != nil {
		slog.Error("annual_report_repository.ListByEmployeeUID.query", "error", err, "employee_uid", employeeUID)
		return nil, err
	}
	defer rows.Close()

	reports := make([]*domain.AnnualReport, 0)
	for rows.Next() {
		var report domain.AnnualReport
		if err := rows.Scan(
			&report.ID,
			&report.EmployeeUID,
			&report.ReportYear,
			&report.ReportGrade,
			&report.Notes,
			&report.ReportImageURL,
			&report.CreatedAt,
			&report.UpdatedAt,
		); err != nil {
			slog.Error("annual_report_repository.ListByEmployeeUID.scan", "error", err, "employee_uid", employeeUID)
			return nil, err
		}
		reports = append(reports, &report)
	}

	if err := rows.Err(); err != nil {
		slog.Error("annual_report_repository.ListByEmployeeUID.rows", "error", err, "employee_uid", employeeUID)
		return nil, err
	}

	return reports, nil
}

func (r *AnnualReportRepository) Create(ctx context.Context, q ports.Querier, report *domain.AnnualReport) error {
	query := `
		INSERT INTO annual_reports (
			employee_uid, report_year, report_grade, notes, report_image_url, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`

	now := time.Now()
	report.CreatedAt = now
	report.UpdatedAt = now

	var id int64
	err := q.QueryRowContext(
		ctx, query,
		report.EmployeeUID,
		report.ReportYear,
		report.ReportGrade,
		report.Notes,
		report.ReportImageURL,
		report.CreatedAt,
		report.UpdatedAt,
	).Scan(&id)
	if err != nil {
		slog.Error("annual_report_repository.Create.query_row", "error", err, "employee_uid", report.EmployeeUID)
		return err
	}
	report.ID = id
	return nil
}

func (r *AnnualReportRepository) DeleteByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) error {
	query := `DELETE FROM annual_reports WHERE employee_uid = $1`
	if _, err := q.ExecContext(ctx, query, employeeUID); err != nil {
		slog.Error("annual_report_repository.DeleteByEmployeeUID.exec_query", "error", err, "employee_uid", employeeUID)
		return err
	}
	return nil
}
