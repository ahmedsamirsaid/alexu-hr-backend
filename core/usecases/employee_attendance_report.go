package usecases

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/xuri/excelize/v2"
)

type ExportEmployeeAttendanceReportInput struct {
	EmployeeUID string
	StartDate   time.Time
	EndDate     time.Time
}

type ExportEmployeeAttendanceReportOutput struct {
	Data        []byte
	Filename    string
	ContentType string
}

type ExportEmployeeAttendanceReportUseCase struct {
	statsUC *GetMonthlyAttendanceStatsUseCase
}

func NewExportEmployeeAttendanceReportUseCase(statsUC *GetMonthlyAttendanceStatsUseCase) *ExportEmployeeAttendanceReportUseCase {
	return &ExportEmployeeAttendanceReportUseCase{statsUC: statsUC}
}

func (uc *ExportEmployeeAttendanceReportUseCase) Execute(ctx context.Context, input ExportEmployeeAttendanceReportInput) (*ExportEmployeeAttendanceReportOutput, error) {
	start, end, err := normalizeAttendanceReportRange(input.StartDate, input.EndDate)
	if err != nil {
		return nil, err
	}

	label := fmt.Sprintf("%s to %s", start.Format("2006-01-02"), end.Format("2006-01-02"))
	report, err := uc.statsUC.Execute(ctx, GetMonthlyAttendanceStatsInput{
		EmployeeUID: input.EmployeeUID,
		StartDate:   &start,
		EndDate:     &end,
		PeriodLabel: &label,
	})
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Employee Report"
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")

	if err := writeEmployeeAttendanceReportWorkbook(f, sheetName, input.EmployeeUID, start, end, report); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return &ExportEmployeeAttendanceReportOutput{
		Data:        buf.Bytes(),
		Filename:    employeeAttendanceReportFilename(input.EmployeeUID, start, end),
		ContentType: departmentAttendanceReportContentType,
	}, nil
}

func writeEmployeeAttendanceReportWorkbook(f *excelize.File, sheetName string, employeeUID string, start, end time.Time, report *GetMonthlyAttendanceStatsOutput) error {
	summaryRows := [][]any{
		{"Employee UID", employeeUID},
		{"Employee Name", report.EmployeeName},
		{"Date Range", fmt.Sprintf("%s to %s", start.Format("2006-01-02"), end.Format("2006-01-02"))},
		{"Total Worked Hours", report.TotalWorkedHours},
		{"Late Days", report.LateDaysCount},
		{"Early Departure Days", report.EarlyDepartureDaysCount},
		{"Absent Days", report.AbsentDaysCount},
		{"Missing Check-In Days", report.MissingCheckInCount},
		{"Missing Check-Out Days", report.MissingCheckOutCount},
	}
	for rowIndex, row := range summaryRows {
		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex+1), &row); err != nil {
			return err
		}
	}

	headers := []any{
		"Date",
		"Check-In",
		"Check-Out",
		"Worked Hours",
		"Late",
		"Late Minutes",
		"Early Departure",
		"Early Departure Minutes",
		"Absent",
		"On Leave",
		"Leave Type",
	}
	headerRow := 11
	if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", headerRow), &headers); err != nil {
		return err
	}

	for index, day := range report.DaysBreakdown {
		rowIndex := headerRow + index + 1
		row := []any{
			day.Date.Format("2006-01-02"),
			formatOptionalReportDateTime(day.CheckIn),
			formatOptionalReportDateTime(day.CheckOut),
			day.WorkedHours,
			formatReportBool(day.IsLate),
			day.LateMinutes,
			formatReportBool(day.IsEarlyDeparture),
			day.EarlyDepartureMinutes,
			formatReportBool(day.IsAbsent),
			formatReportBool(day.IsOnLeave),
			formatOptionalReportString(day.LeaveTypeName),
		}
		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex), &row); err != nil {
			return err
		}
	}

	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	if err != nil {
		return err
	}
	if err := f.SetRowStyle(sheetName, headerRow, headerRow, headerStyle); err != nil {
		return err
	}

	numberStyle, err := f.NewStyle(&excelize.Style{NumFmt: 2})
	if err != nil {
		return err
	}
	if len(report.DaysBreakdown) > 0 {
		firstDataRow := headerRow + 1
		lastDataRow := headerRow + len(report.DaysBreakdown)
		if err := f.SetCellStyle(sheetName, fmt.Sprintf("D%d", firstDataRow), fmt.Sprintf("D%d", lastDataRow), numberStyle); err != nil {
			return err
		}
	}
	if err := f.SetCellStyle(sheetName, "B4", "B4", numberStyle); err != nil {
		return err
	}

	widths := []float64{14, 24, 24, 14, 12, 14, 18, 24, 12, 12, 24}
	for index, width := range widths {
		column, err := excelize.ColumnNumberToName(index + 1)
		if err != nil {
			return err
		}
		if err := f.SetColWidth(sheetName, column, column, width); err != nil {
			return err
		}
	}

	return nil
}

func formatOptionalReportDateTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func formatOptionalReportString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func formatReportBool(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}

func employeeAttendanceReportFilename(employeeUID string, start, end time.Time) string {
	return fmt.Sprintf(
		"employee_attendance_report_%s_%s_to_%s.xlsx",
		employeeUID,
		start.Format("2006-01-02"),
		end.Format("2006-01-02"),
	)
}
