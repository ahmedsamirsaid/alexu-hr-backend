package usecases

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

type ExportEmployeeAttendanceReportInput struct {
	EmployeeUID string
	StartDate   time.Time
	EndDate     time.Time
	Language    string
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

	language := normalizeReportLanguage(input.Language)

	label := formatReportDateRange(start, end, language)
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

	sheetName := employeeAttendanceReportSheetName(language)
	index, err := f.NewSheet(sheetName)
	if err != nil {
		return nil, err
	}
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")
	if isArabicLanguage(language) {
		rightToLeft := true
		if err := f.SetSheetView(sheetName, 0, &excelize.ViewOptions{RightToLeft: &rightToLeft}); err != nil {
			return nil, err
		}
	}

	if err := writeEmployeeAttendanceReportWorkbook(f, sheetName, input.EmployeeUID, start, end, report, language); err != nil {
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

func writeEmployeeAttendanceReportWorkbook(f *excelize.File, sheetName string, employeeUID string, start, end time.Time, report *GetMonthlyAttendanceStatsOutput, language string) error {
	labels := employeeAttendanceReportLabelsForLanguage(language)
	summaryRows := [][]any{
		{labels.SummaryEmployeeUID, employeeUID},
		{labels.SummaryEmployeeName, report.EmployeeName},
		{labels.SummaryDateRange, formatReportDateRange(start, end, language)},
		{labels.SummaryTotalWorkedHours, report.TotalWorkedHours},
		{labels.SummaryLateDays, report.LateDaysCount},
		{labels.SummaryEarlyDepartureDays, report.EarlyDepartureDaysCount},
		{labels.SummaryAbsentDays, report.AbsentDaysCount},
		{labels.SummaryMissingCheckInDays, report.MissingCheckInCount},
		{labels.SummaryMissingCheckOutDays, report.MissingCheckOutCount},
	}
	for rowIndex, row := range summaryRows {
		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex+1), &row); err != nil {
			return err
		}
	}

	headers := []any{
		labels.HeaderDate,
		labels.HeaderCheckIn,
		labels.HeaderCheckOut,
		labels.HeaderWorkedHours,
		labels.HeaderLate,
		labels.HeaderLateMinutes,
		labels.HeaderEarlyDeparture,
		labels.HeaderEarlyDepartureMinutes,
		labels.HeaderAbsent,
		labels.HeaderOnLeave,
		labels.HeaderLeaveType,
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
			formatReportBool(day.IsLate, language),
			day.LateMinutes,
			formatReportBool(day.IsEarlyDeparture, language),
			day.EarlyDepartureMinutes,
			formatReportBool(day.IsAbsent, language),
			formatReportBool(day.IsOnLeave, language),
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

func formatReportBool(value bool, language string) string {
	if value {
		if isArabicLanguage(language) {
			return "نعم"
		}
		return "Yes"
	}
	if isArabicLanguage(language) {
		return "لا"
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

type employeeAttendanceReportLabels struct {
	SummaryEmployeeUID          string
	SummaryEmployeeName         string
	SummaryDateRange            string
	SummaryTotalWorkedHours     string
	SummaryLateDays             string
	SummaryEarlyDepartureDays   string
	SummaryAbsentDays           string
	SummaryMissingCheckInDays   string
	SummaryMissingCheckOutDays  string
	HeaderDate                  string
	HeaderCheckIn               string
	HeaderCheckOut              string
	HeaderWorkedHours           string
	HeaderLate                  string
	HeaderLateMinutes           string
	HeaderEarlyDeparture        string
	HeaderEarlyDepartureMinutes string
	HeaderAbsent                string
	HeaderOnLeave               string
	HeaderLeaveType             string
}

func employeeAttendanceReportLabelsForLanguage(language string) employeeAttendanceReportLabels {
	if isArabicLanguage(language) {
		return employeeAttendanceReportLabels{
			SummaryEmployeeUID:          "معرف الموظف",
			SummaryEmployeeName:         "اسم الموظف",
			SummaryDateRange:            "نطاق التاريخ",
			SummaryTotalWorkedHours:     "إجمالي ساعات العمل",
			SummaryLateDays:             "أيام التأخر",
			SummaryEarlyDepartureDays:   "أيام الانصراف المبكر",
			SummaryAbsentDays:           "أيام الغياب",
			SummaryMissingCheckInDays:   "أيام نسيان الدخول",
			SummaryMissingCheckOutDays:  "أيام نسيان الخروج",
			HeaderDate:                  "التاريخ",
			HeaderCheckIn:               "الدخول",
			HeaderCheckOut:              "الخروج",
			HeaderWorkedHours:           "ساعات العمل",
			HeaderLate:                  "تأخير",
			HeaderLateMinutes:           "دقائق التأخر",
			HeaderEarlyDeparture:        "انصراف مبكر",
			HeaderEarlyDepartureMinutes: "دقائق الانصراف المبكر",
			HeaderAbsent:                "غياب",
			HeaderOnLeave:               "في إجازة",
			HeaderLeaveType:             "نوع الإجازة",
		}
	}

	return employeeAttendanceReportLabels{
		SummaryEmployeeUID:          "Employee UID",
		SummaryEmployeeName:         "Employee Name",
		SummaryDateRange:            "Date Range",
		SummaryTotalWorkedHours:     "Total Worked Hours",
		SummaryLateDays:             "Late Days",
		SummaryEarlyDepartureDays:   "Early Departure Days",
		SummaryAbsentDays:           "Absent Days",
		SummaryMissingCheckInDays:   "Missing Check-In Days",
		SummaryMissingCheckOutDays:  "Missing Check-Out Days",
		HeaderDate:                  "Date",
		HeaderCheckIn:               "Check-In",
		HeaderCheckOut:              "Check-Out",
		HeaderWorkedHours:           "Worked Hours",
		HeaderLate:                  "Late",
		HeaderLateMinutes:           "Late Minutes",
		HeaderEarlyDeparture:        "Early Departure",
		HeaderEarlyDepartureMinutes: "Early Departure Minutes",
		HeaderAbsent:                "Absent",
		HeaderOnLeave:               "On Leave",
		HeaderLeaveType:             "Leave Type",
	}
}

func employeeAttendanceReportSheetName(language string) string {
	if isArabicLanguage(language) {
		return "تقرير الحضور والانصراف"
	}
	return "Employee Report"
}

func formatReportDateRange(start, end time.Time, language string) string {
	if isArabicLanguage(language) {
		return fmt.Sprintf("%s إلى %s", start.Format("2006-01-02"), end.Format("2006-01-02"))
	}
	return fmt.Sprintf("%s to %s", start.Format("2006-01-02"), end.Format("2006-01-02"))
}

func normalizeReportLanguage(language string) string {
	value := strings.ToLower(strings.TrimSpace(language))
	if strings.HasPrefix(value, "ar") {
		return "ar"
	}
	return "en"
}

func isArabicLanguage(language string) bool {
	return normalizeReportLanguage(language) == "ar"
}
