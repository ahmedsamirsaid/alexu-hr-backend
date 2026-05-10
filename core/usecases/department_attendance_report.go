package usecases

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
	"github.com/xuri/excelize/v2"
)

const departmentAttendanceReportContentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"

type DepartmentAttendanceReportEmployee struct {
	EmployeeUID         string
	EmployeeName        string
	TotalWorkingDays    int
	DaysPresent         int
	DaysAbsent          int
	LateDays            int
	EarlyDepartureDays  int
	MissingCheckInDays  int
	MissingCheckOutDays int
	TotalWorkedHours    float64
	AverageCheckInTime  *time.Time
	AverageCheckOutTime *time.Time
}

type GetDepartmentAttendanceReportInput struct {
	DepartmentUID string
	StartDate     time.Time
	EndDate       time.Time
	Language      string
}

type DepartmentAttendanceReportOutput struct {
	DepartmentUID          string
	StartDate              time.Time
	EndDate                time.Time
	TotalHeadcount         int
	AverageAttendanceRate  float64
	AveragePunctualityRate float64
	OverviewTotalRecords   int
	OverviewAttendanceRate float64
	OverviewLateArrivalRate float64
	OverviewMissingPunchRate float64
	OverviewTotalExceptions int
	OverviewBestAttendanceDay *DepartmentAttendanceOverviewDay
	OverviewWorstAttendanceDay *DepartmentAttendanceOverviewDay
	OverviewTopExceptions []DepartmentAttendanceOverviewException
	OverviewDailyTrend []DepartmentAttendanceOverviewTrendPoint
	Employees              []DepartmentAttendanceReportEmployee
}

type DepartmentAttendanceOverviewDay struct {
	Date           time.Time
	AttendanceRate float64
}

type DepartmentAttendanceOverviewException struct {
	Type  domain.AttendanceExceptionType
	Count int
}

type DepartmentAttendanceOverviewTrendPoint struct {
	Date             time.Time
	Records          int
	AttendanceRate   float64
	AbsenceRate      float64
	LateArrivalRate  float64
	MissingPunchRate float64
	ExceptionsCount  int
}

type GetDepartmentAttendanceReportUseCase struct {
	db           ports.DB
	deptRepo     ports.DepartmentRepository
	recordRepo   ports.AttendanceRecordRepository
	employeeRepo ports.EmployeeRepository
	shiftRepo    ports.ShiftRepository
	weekendRepo  ports.WeekendConfigRepository
	holidayRepo  ports.HolidayDefinitionRepository
}

func NewGetDepartmentAttendanceReportUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	recordRepo ports.AttendanceRecordRepository,
	employeeRepo ports.EmployeeRepository,
	shiftRepo ports.ShiftRepository,
	weekendRepo ports.WeekendConfigRepository,
	holidayRepo ports.HolidayDefinitionRepository,
) *GetDepartmentAttendanceReportUseCase {
	return &GetDepartmentAttendanceReportUseCase{
		db:           db,
		deptRepo:     deptRepo,
		recordRepo:   recordRepo,
		employeeRepo: employeeRepo,
		shiftRepo:    shiftRepo,
		weekendRepo:  weekendRepo,
		holidayRepo:  holidayRepo,
	}
}

func (uc *GetDepartmentAttendanceReportUseCase) Execute(ctx context.Context, input GetDepartmentAttendanceReportInput) (*DepartmentAttendanceReportOutput, error) {
	start, end, err := normalizeAttendanceReportRange(input.StartDate, input.EndDate)
	if err != nil {
		return nil, err
	}

	department, err := uc.deptRepo.GetByUID(ctx, uc.db, input.DepartmentUID)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, ErrDepartmentNotFound
	}

	employees, err := listDepartmentEmployeesForAttendance(ctx, uc.db, input.DepartmentUID, ports.DepartmentAttendanceLogsFilter{})
	if err != nil {
		return nil, err
	}

	checker, err := newNonWorkingDateChecker(ctx, uc.db, uc.weekendRepo, uc.holidayRepo, &start, &end)
	if err != nil {
		return nil, err
	}

	accumulators := make(map[string]*attendanceReportAccumulator, len(employees))
	for _, employee := range employees {
		workingDates, err := employeeWorkingDateKeys(checker, employee, start, end)
		if err != nil {
			return nil, err
		}
		accumulators[employee.UID] = newAttendanceReportAccumulator(employee, workingDates)
	}

	filter := ports.DepartmentAttendanceLogsFilter{
		StartDate: &start,
		EndDate:   &end,
	}
	params := ports.ListParams{
		Page:      1,
		PageSize:  100000,
		SortBy:    "date",
		SortOrder: ports.SortOrderAsc,
	}

	groups, err := uc.recordRepo.ListDailyByDepartmentUID(ctx, uc.db, input.DepartmentUID, filter, params)
	if err != nil {
		return nil, err
	}

	dailyItems, err := buildDailyAttendanceLogItems(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, groups)
	if err != nil {
		return nil, err
	}

	overviewRecords, err := appendDepartmentAbsenceItems(
		ctx,
		uc.db,
		nil,
		uc.weekendRepo,
		uc.holidayRepo,
		append([]DailyAttendanceLogItem(nil), dailyItems...),
		employees,
		input.DepartmentUID,
		filter,
		start,
		end,
	)
	if err != nil {
		return nil, err
	}

	for _, item := range dailyItems {
		accumulator, ok := accumulators[item.EmployeeUID]
		if !ok {
			continue
		}
		accumulator.addDailyItem(item)
	}

	exceptions, err := uc.listDepartmentAttendanceReportExceptions(ctx, input.DepartmentUID, start, end)
	if err != nil {
		return nil, err
	}
	for _, exception := range exceptions {
		accumulator, ok := accumulators[exception.EmployeeUID]
		if !ok {
			continue
		}
		accumulator.addException(exception)
	}

	output := &DepartmentAttendanceReportOutput{
		DepartmentUID:  input.DepartmentUID,
		StartDate:      start,
		EndDate:        end,
		TotalHeadcount: len(employees),
		Employees:      make([]DepartmentAttendanceReportEmployee, 0, len(employees)),
	}

	overviewMetrics := buildDepartmentAttendanceOverviewMetrics(overviewRecords)
	output.OverviewTotalRecords = overviewMetrics.TotalRecords
	output.OverviewAttendanceRate = overviewMetrics.AttendanceRate
	output.OverviewLateArrivalRate = overviewMetrics.LateArrivalRate
	output.OverviewMissingPunchRate = overviewMetrics.MissingPunchRate
	output.OverviewTotalExceptions = overviewMetrics.TotalExceptions
	output.OverviewBestAttendanceDay = overviewMetrics.BestAttendanceDay
	output.OverviewWorstAttendanceDay = overviewMetrics.WorstAttendanceDay
	output.OverviewTopExceptions = overviewMetrics.TopExceptions
	output.OverviewDailyTrend = overviewMetrics.DailyTrend

	totalWorkingDays := 0
	totalPresentDays := 0
	totalPunctualDays := 0

	for _, employee := range employees {
		reportEmployee := accumulators[employee.UID].buildEmployee(start)
		output.Employees = append(output.Employees, reportEmployee)

		totalWorkingDays += reportEmployee.TotalWorkingDays
		totalPresentDays += reportEmployee.DaysPresent
		latePresentDays := accumulators[employee.UID].countLatePresentDays()
		punctualDays := reportEmployee.DaysPresent - latePresentDays
		if punctualDays < 0 {
			punctualDays = 0
		}
		totalPunctualDays += punctualDays
	}

	output.AverageAttendanceRate = percentage(totalPresentDays, totalWorkingDays)
	output.AveragePunctualityRate = percentage(totalPunctualDays, totalPresentDays)

	return output, nil
}

type ExportDepartmentAttendanceReportOutput struct {
	Data        []byte
	Filename    string
	ContentType string
}

type ExportDepartmentAttendanceReportUseCase struct {
	reportUC *GetDepartmentAttendanceReportUseCase
}

func NewExportDepartmentAttendanceReportUseCase(reportUC *GetDepartmentAttendanceReportUseCase) *ExportDepartmentAttendanceReportUseCase {
	return &ExportDepartmentAttendanceReportUseCase{reportUC: reportUC}
}

func (uc *ExportDepartmentAttendanceReportUseCase) Execute(ctx context.Context, input GetDepartmentAttendanceReportInput) (*ExportDepartmentAttendanceReportOutput, error) {
	report, err := uc.reportUC.Execute(ctx, input)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()

	language := normalizeReportLanguage(input.Language)
	sheetName := departmentAttendanceReportSheetName(language)
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

	if err := writeDepartmentAttendanceReportWorkbook(f, sheetName, report, language); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		return nil, err
	}

	return &ExportDepartmentAttendanceReportOutput{
		Data:        buf.Bytes(),
		Filename:    departmentAttendanceReportFilename(report),
		ContentType: departmentAttendanceReportContentType,
	}, nil
}

type attendanceReportExceptionRow struct {
	EmployeeUID string
	Date        time.Time
	Type        domain.AttendanceExceptionType
}

func (uc *GetDepartmentAttendanceReportUseCase) listDepartmentAttendanceReportExceptions(ctx context.Context, departmentUID string, start, end time.Time) ([]attendanceReportExceptionRow, error) {
	rows, err := uc.db.QueryContext(ctx, `
		SELECT ae.employee_uid, ae.attendance_date, ae.exception_type
		FROM attendance_exceptions ae
		INNER JOIN employees e ON e.uid = ae.employee_uid
		WHERE e.department_uid = $1
			AND e.status = $2
			AND ae.attendance_date BETWEEN $3 AND $4
		ORDER BY ae.attendance_date ASC, e.name ASC, e.uid ASC
	`, departmentUID, domain.EmployeeStatusActive, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make([]attendanceReportExceptionRow, 0)
	for rows.Next() {
		var employeeUID string
		var dateValue string
		var exceptionType domain.AttendanceExceptionType

		if err := rows.Scan(&employeeUID, &dateValue, &exceptionType); err != nil {
			return nil, err
		}

		parsedDate, err := parseDateOrTimestamp(dateValue)
		if err != nil {
			return nil, err
		}

		result = append(result, attendanceReportExceptionRow{
			EmployeeUID: employeeUID,
			Date:        parsedDate,
			Type:        exceptionType,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

type attendanceReportAccumulator struct {
	employeeUID          string
	employeeName         string
	workingDates         map[string]struct{}
	presentDates         map[string]struct{}
	absentDates          map[string]struct{}
	lateDates            map[string]struct{}
	earlyDates           map[string]struct{}
	missingCheckInDates  map[string]struct{}
	missingCheckOutDates map[string]struct{}
	totalWorkedHours     float64
	totalCheckInSeconds  int
	checkInCount         int
	totalCheckOutSeconds int
	checkOutCount        int
}

func newAttendanceReportAccumulator(employee departmentEmployeeForAttendance, workingDates map[string]struct{}) *attendanceReportAccumulator {
	return &attendanceReportAccumulator{
		employeeUID:          employee.UID,
		employeeName:         employee.Name,
		workingDates:         workingDates,
		presentDates:         make(map[string]struct{}),
		absentDates:          make(map[string]struct{}),
		lateDates:            make(map[string]struct{}),
		earlyDates:           make(map[string]struct{}),
		missingCheckInDates:  make(map[string]struct{}),
		missingCheckOutDates: make(map[string]struct{}),
	}
}

func (a *attendanceReportAccumulator) addDailyItem(item DailyAttendanceLogItem) {
	dateKey := normalizeDateOnly(item.Date).Format("2006-01-02")
	if !a.isWorkingDate(dateKey) {
		return
	}

	if item.CheckIn != nil || item.CheckOut != nil {
		a.presentDates[dateKey] = struct{}{}
	}
	if item.CheckIn != nil {
		a.totalCheckInSeconds += secondsSinceMidnight(*item.CheckIn)
		a.checkInCount++
	}
	if item.CheckOut != nil {
		a.totalCheckOutSeconds += secondsSinceMidnight(*item.CheckOut)
		a.checkOutCount++
	}
	if item.WorkedHours > 0 {
		a.totalWorkedHours += item.WorkedHours
	}
	if item.LateArrival {
		a.lateDates[dateKey] = struct{}{}
	}
	if item.EarlyDeparture {
		a.earlyDates[dateKey] = struct{}{}
	}
	if item.MissingCheckIn {
		a.missingCheckInDates[dateKey] = struct{}{}
	}
	if item.MissingCheckOut {
		a.missingCheckOutDates[dateKey] = struct{}{}
	}
}

func (a *attendanceReportAccumulator) addException(row attendanceReportExceptionRow) {
	dateKey := normalizeDateOnly(row.Date).Format("2006-01-02")
	if !a.isWorkingDate(dateKey) {
		return
	}

	switch row.Type {
	case domain.AttendanceExceptionTypeAbsence:
		a.absentDates[dateKey] = struct{}{}
	case domain.AttendanceExceptionTypeLateArrival:
		a.lateDates[dateKey] = struct{}{}
	case domain.AttendanceExceptionTypeEarlyDeparture:
		a.earlyDates[dateKey] = struct{}{}
	case domain.AttendanceExceptionTypeMissedPunchIn:
		a.missingCheckInDates[dateKey] = struct{}{}
	case domain.AttendanceExceptionTypeMissedPunchOut:
		a.missingCheckOutDates[dateKey] = struct{}{}
	}
}

func (a *attendanceReportAccumulator) buildEmployee(anchor time.Time) DepartmentAttendanceReportEmployee {
	return DepartmentAttendanceReportEmployee{
		EmployeeUID:         a.employeeUID,
		EmployeeName:        a.employeeName,
		TotalWorkingDays:    len(a.workingDates),
		DaysPresent:         len(a.presentDates),
		DaysAbsent:          countDatesNotIn(a.absentDates, a.presentDates),
		LateDays:            len(a.lateDates),
		EarlyDepartureDays:  len(a.earlyDates),
		MissingCheckInDays:  len(a.missingCheckInDates),
		MissingCheckOutDays: len(a.missingCheckOutDates),
		TotalWorkedHours:    a.totalWorkedHours,
		AverageCheckInTime:  averageTimeOfDay(anchor, a.totalCheckInSeconds, a.checkInCount),
		AverageCheckOutTime: averageTimeOfDay(anchor, a.totalCheckOutSeconds, a.checkOutCount),
	}
}

func (a *attendanceReportAccumulator) countLatePresentDays() int {
	count := 0
	for dateKey := range a.lateDates {
		if _, ok := a.presentDates[dateKey]; ok {
			count++
		}
	}
	return count
}

func (a *attendanceReportAccumulator) isWorkingDate(dateKey string) bool {
	_, ok := a.workingDates[dateKey]
	return ok
}

func normalizeAttendanceReportRange(startDate, endDate time.Time) (time.Time, time.Time, error) {
	if startDate.IsZero() || endDate.IsZero() {
		return time.Time{}, time.Time{}, ErrInvalidReportDateRange
	}

	start := normalizeDateOnly(startDate)
	end := normalizeDateOnly(endDate)
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidReportDateRange
	}
	if countDaysInclusive(start, end) > maxAbsenceRangeDays {
		return time.Time{}, time.Time{}, ErrInvalidReportDateRange
	}

	return start, end, nil
}

func employeeWorkingDateKeys(checker *nonWorkingDateChecker, employee departmentEmployeeForAttendance, start, end time.Time) (map[string]struct{}, error) {
	result := make(map[string]struct{})

	effectiveStart := start
	hireDate := normalizeDateOnly(employee.HireDate)
	if hireDate.After(effectiveStart) {
		effectiveStart = hireDate
	}
	if effectiveStart.After(end) {
		return result, nil
	}

	for day := effectiveStart; !day.After(end); day = day.AddDate(0, 0, 1) {
		nonWorking, err := checker.IsNonWorking(day)
		if err != nil {
			return nil, err
		}
		if nonWorking {
			continue
		}
		result[day.Format("2006-01-02")] = struct{}{}
	}

	return result, nil
}

func secondsSinceMidnight(value time.Time) int {
	localValue := value.In(time.Local)
	return localValue.Hour()*3600 + localValue.Minute()*60 + localValue.Second()
}

func averageTimeOfDay(anchor time.Time, totalSeconds, count int) *time.Time {
	if count == 0 {
		return nil
	}

	avgSeconds := totalSeconds / count
	value := time.Date(
		anchor.Year(),
		anchor.Month(),
		anchor.Day(),
		avgSeconds/3600,
		(avgSeconds%3600)/60,
		avgSeconds%60,
		0,
		time.Local,
	)
	return &value
}

func countDatesNotIn(values, excluded map[string]struct{}) int {
	count := 0
	for dateKey := range values {
		if _, ok := excluded[dateKey]; ok {
			continue
		}
		count++
	}
	return count
}

func percentage(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return (float64(numerator) / float64(denominator)) * 100
}

type departmentAttendanceOverviewMetrics struct {
	TotalRecords      int
	AttendanceRate    float64
	LateArrivalRate   float64
	MissingPunchRate  float64
	TotalExceptions   int
	BestAttendanceDay *DepartmentAttendanceOverviewDay
	WorstAttendanceDay *DepartmentAttendanceOverviewDay
	TopExceptions     []DepartmentAttendanceOverviewException
	DailyTrend        []DepartmentAttendanceOverviewTrendPoint
}

type departmentAttendanceOverviewDayBucket struct {
	Date           time.Time
	Total          int
	AttendanceDays int
	LateArrivalDays int
	MissingPunchDays int
	ExceptionsCount int
}

func buildDepartmentAttendanceOverviewMetrics(records []DailyAttendanceLogItem) departmentAttendanceOverviewMetrics {
	totalRecords := len(records)
	if totalRecords == 0 {
		return departmentAttendanceOverviewMetrics{}
	}

	attendanceDays := 0
	lateArrivalDays := 0
	missingPunchDays := 0
	totalExceptions := 0
	exceptionCounts := make(map[domain.AttendanceExceptionType]int)
	dailyBuckets := make(map[string]*departmentAttendanceOverviewDayBucket)

	for _, record := range records {
		dateKey := normalizeDateOnly(record.Date).Format("2006-01-02")
		bucket, ok := dailyBuckets[dateKey]
		if !ok {
			bucket = &departmentAttendanceOverviewDayBucket{Date: normalizeDateOnly(record.Date)}
			dailyBuckets[dateKey] = bucket
		}

		hasAbsence := false
		hasLateArrival := false
		hasMissingPunch := false

		for _, exceptionType := range record.Exceptions {
			exceptionCounts[exceptionType]++
			totalExceptions++

			switch exceptionType {
			case domain.AttendanceExceptionTypeAbsence:
				hasAbsence = true
			case domain.AttendanceExceptionTypeLateArrival:
				hasLateArrival = true
			case domain.AttendanceExceptionTypeMissedPunchIn, domain.AttendanceExceptionTypeMissedPunchOut:
				hasMissingPunch = true
			}
		}

		if !hasAbsence {
			attendanceDays++
			bucket.AttendanceDays++
		}
		if hasLateArrival {
			lateArrivalDays++
			bucket.LateArrivalDays++
		}
		if hasMissingPunch {
			missingPunchDays++
			bucket.MissingPunchDays++
		}

		bucket.ExceptionsCount += len(record.Exceptions)

		bucket.Total++
	}

	dailyTrend := make([]DepartmentAttendanceOverviewDay, 0, len(dailyBuckets))
	dailyTrendExport := make([]DepartmentAttendanceOverviewTrendPoint, 0, len(dailyBuckets))
	for _, bucket := range dailyBuckets {
		dailyTrend = append(dailyTrend, DepartmentAttendanceOverviewDay{
			Date:           bucket.Date,
			AttendanceRate: ratio(bucket.AttendanceDays, bucket.Total),
		})
		dailyTrendExport = append(dailyTrendExport, DepartmentAttendanceOverviewTrendPoint{
			Date:             bucket.Date,
			Records:          bucket.Total,
			AttendanceRate:   ratio(bucket.AttendanceDays, bucket.Total),
			AbsenceRate:      ratio(bucket.Total-bucket.AttendanceDays, bucket.Total),
			LateArrivalRate:  ratio(bucket.LateArrivalDays, bucket.Total),
			MissingPunchRate: ratio(bucket.MissingPunchDays, bucket.Total),
			ExceptionsCount:  bucket.ExceptionsCount,
		})
	}
	sort.Slice(dailyTrend, func(i, j int) bool {
		return dailyTrend[i].Date.Before(dailyTrend[j].Date)
	})
	sort.Slice(dailyTrendExport, func(i, j int) bool {
		return dailyTrendExport[i].Date.Before(dailyTrendExport[j].Date)
	})

	var bestAttendanceDay *DepartmentAttendanceOverviewDay
	var worstAttendanceDay *DepartmentAttendanceOverviewDay
	for i := range dailyTrend {
		point := dailyTrend[i]
		if bestAttendanceDay == nil || point.AttendanceRate > bestAttendanceDay.AttendanceRate {
			copied := point
			bestAttendanceDay = &copied
		}
		if worstAttendanceDay == nil || point.AttendanceRate < worstAttendanceDay.AttendanceRate {
			copied := point
			worstAttendanceDay = &copied
		}
	}

	topExceptions := make([]DepartmentAttendanceOverviewException, 0, len(exceptionCounts))
	for exceptionType, count := range exceptionCounts {
		topExceptions = append(topExceptions, DepartmentAttendanceOverviewException{Type: exceptionType, Count: count})
	}
	sort.Slice(topExceptions, func(i, j int) bool {
		if topExceptions[i].Count == topExceptions[j].Count {
			return string(topExceptions[i].Type) < string(topExceptions[j].Type)
		}
		return topExceptions[i].Count > topExceptions[j].Count
	})
	if len(topExceptions) > 5 {
		topExceptions = topExceptions[:5]
	}

	return departmentAttendanceOverviewMetrics{
		TotalRecords:       totalRecords,
		AttendanceRate:     ratio(attendanceDays, totalRecords),
		LateArrivalRate:    ratio(lateArrivalDays, totalRecords),
		MissingPunchRate:   ratio(missingPunchDays, totalRecords),
		TotalExceptions:    totalExceptions,
		BestAttendanceDay:  bestAttendanceDay,
		WorstAttendanceDay: worstAttendanceDay,
		TopExceptions:      topExceptions,
		DailyTrend:         dailyTrendExport,
	}
}

func ratio(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func writeDepartmentAttendanceReportWorkbook(f *excelize.File, sheetName string, report *DepartmentAttendanceReportOutput, language string) error {
	labels := departmentAttendanceReportLabelsForLanguage(language)
	summaryRows := [][]any{
		{labels.OverviewTitle, ""},
		{labels.DateRange, formatReportDateRange(report.StartDate, report.EndDate, language)},
		{labels.AttendanceRate, report.OverviewAttendanceRate},
		{labels.LateArrivalRate, report.OverviewLateArrivalRate},
		{labels.MissingPunchRate, report.OverviewMissingPunchRate},
		{labels.TotalRecords, report.OverviewTotalRecords},
		{labels.RecordsAnalyzed, report.OverviewTotalRecords},
		{labels.DaysCovered, len(report.OverviewDailyTrend)},
		{labels.TotalExceptions, report.OverviewTotalExceptions},
		{labels.BestAttendanceDay, formatOverviewDay(report.OverviewBestAttendanceDay, language)},
		{labels.WorstAttendanceDay, formatOverviewDay(report.OverviewWorstAttendanceDay, language)},
	}

	if len(report.OverviewTopExceptions) == 0 {
		summaryRows = append(summaryRows, []any{labels.TopExceptions, labels.NoExceptions})
	} else {
		for index, exception := range report.OverviewTopExceptions {
			label := labels.TopExceptions
			if index > 0 {
				label = ""
			}
			summaryRows = append(summaryRows, []any{label, fmt.Sprintf("%s: %d", formatOverviewExceptionType(exception.Type, language), exception.Count)})
		}
	}

	for rowIndex, row := range summaryRows {
		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex+1), &row); err != nil {
			return err
		}
	}

	trendTitleRow := len(summaryRows) + 2
	trendTitle := []any{labels.AttendanceTrendTitle}
	if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", trendTitleRow), &trendTitle); err != nil {
		return err
	}

	trendHeaderRow := trendTitleRow + 1
	trendHeaders := []any{
		labels.TrendDate,
		labels.TrendAttendanceRate,
		labels.TrendAbsenceRate,
		labels.TrendLateArrivalRate,
		labels.TrendMissingPunchRate,
		labels.TrendExceptions,
		labels.TrendRecords,
	}
	if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", trendHeaderRow), &trendHeaders); err != nil {
		return err
	}

	trendDataStartRow := trendHeaderRow + 1
	if len(report.OverviewDailyTrend) == 0 {
		emptyTrend := []any{labels.NoTrendData}
		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", trendDataStartRow), &emptyTrend); err != nil {
			return err
		}
	}
	for index, point := range report.OverviewDailyTrend {
		rowIndex := trendDataStartRow + index
		row := []any{
			point.Date.Format("2006-01-02"),
			point.AttendanceRate,
			point.AbsenceRate,
			point.LateArrivalRate,
			point.MissingPunchRate,
			point.ExceptionsCount,
			point.Records,
		}
		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", rowIndex), &row); err != nil {
			return err
		}
	}

	headers := []any{
		labels.Employee,
		labels.WorkingDays,
		labels.Present,
		labels.Absent,
		labels.Late,
		labels.EarlyDeparture,
		labels.MissingInOut,
		labels.TotalHours,
		labels.AverageCheckIn,
		labels.AverageCheckOut,
	}
	trendRowsCount := len(report.OverviewDailyTrend)
	if trendRowsCount == 0 {
		trendRowsCount = 1
	}
	employeeTitleRow := trendDataStartRow + trendRowsCount + 2
	employeeTitle := []any{labels.EmployeeBreakdownTitle}
	if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", employeeTitleRow), &employeeTitle); err != nil {
		return err
	}

	headerRow := employeeTitleRow + 1
	if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", headerRow), &headers); err != nil {
		return err
	}

	for index, employee := range report.Employees {
		rowIndex := headerRow + index + 1
		row := []any{
			employee.EmployeeName,
			employee.TotalWorkingDays,
			employee.DaysPresent,
			employee.DaysAbsent,
			employee.LateDays,
			employee.EarlyDepartureDays,
			fmt.Sprintf("%d / %d", employee.MissingCheckInDays, employee.MissingCheckOutDays),
			employee.TotalWorkedHours,
			formatOptionalReportTime(employee.AverageCheckInTime),
			formatOptionalReportTime(employee.AverageCheckOutTime),
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
	if err := f.SetRowStyle(sheetName, trendHeaderRow, trendHeaderRow, headerStyle); err != nil {
		return err
	}

	numberStyle, err := f.NewStyle(&excelize.Style{NumFmt: 2})
	if err != nil {
		return err
	}
	percentStyle, err := f.NewStyle(&excelize.Style{NumFmt: 10})
	if err != nil {
		return err
	}
	if len(report.Employees) > 0 {
		firstDataRow := headerRow + 1
		lastDataRow := headerRow + len(report.Employees)
		if err := f.SetCellStyle(sheetName, fmt.Sprintf("H%d", firstDataRow), fmt.Sprintf("H%d", lastDataRow), numberStyle); err != nil {
			return err
		}
	}
	if len(report.OverviewDailyTrend) > 0 {
		trendLastRow := trendDataStartRow + len(report.OverviewDailyTrend) - 1
		if err := f.SetCellStyle(sheetName, fmt.Sprintf("B%d", trendDataStartRow), fmt.Sprintf("E%d", trendLastRow), percentStyle); err != nil {
			return err
		}
	}
	if err := f.SetCellStyle(sheetName, "B3", "B5", percentStyle); err != nil {
		return err
	}
	if err := f.SetCellStyle(sheetName, "B6", "B9", numberStyle); err != nil {
		return err
	}

	widths := []float64{30, 16, 12, 12, 12, 18, 18, 14, 18, 18}
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

func formatOverviewDay(value *DepartmentAttendanceOverviewDay, language string) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprintf("%s (%.2f%%)", value.Date.Format("2006-01-02"), value.AttendanceRate*100)
}

func formatOverviewExceptionType(value domain.AttendanceExceptionType, language string) string {
	if isArabicLanguage(language) {
		switch value {
		case domain.AttendanceExceptionTypeAbsence:
			return "غياب"
		case domain.AttendanceExceptionTypeLateArrival:
			return "تأخر في الحضور"
		case domain.AttendanceExceptionTypeEarlyDeparture:
			return "انصراف مبكر"
		case domain.AttendanceExceptionTypeMissedPunchIn:
			return "بصمة دخول مفقودة"
		case domain.AttendanceExceptionTypeMissedPunchOut:
			return "بصمة خروج مفقودة"
		default:
			return string(value)
		}
	}

	switch value {
	case domain.AttendanceExceptionTypeAbsence:
		return "Absence"
	case domain.AttendanceExceptionTypeLateArrival:
		return "Late arrival"
	case domain.AttendanceExceptionTypeEarlyDeparture:
		return "Early departure"
	case domain.AttendanceExceptionTypeMissedPunchIn:
		return "Missed punch in"
	case domain.AttendanceExceptionTypeMissedPunchOut:
		return "Missed punch out"
	default:
		return string(value)
	}
}

func formatOptionalReportTime(value *time.Time) string {
	if value == nil {
		return ""
	}
	return value.Format("15:04:05")
}

func departmentAttendanceReportFilename(report *DepartmentAttendanceReportOutput) string {
	return fmt.Sprintf(
		"department_attendance_report_%s_%s_to_%s.xlsx",
		report.DepartmentUID,
		report.StartDate.Format("2006-01-02"),
		report.EndDate.Format("2006-01-02"),
	)
}

type departmentAttendanceReportLabels struct {
	OverviewTitle         string
	DateRange             string
	AttendanceRate        string
	LateArrivalRate       string
	MissingPunchRate      string
	TotalRecords          string
	RecordsAnalyzed       string
	DaysCovered           string
	TotalExceptions       string
	BestAttendanceDay     string
	WorstAttendanceDay    string
	TopExceptions         string
	NoExceptions          string
	AttendanceTrendTitle  string
	NoTrendData           string
	TrendDate             string
	TrendAttendanceRate   string
	TrendAbsenceRate      string
	TrendLateArrivalRate  string
	TrendMissingPunchRate string
	TrendExceptions       string
	TrendRecords          string
	EmployeeBreakdownTitle string
	Employee              string
	WorkingDays           string
	Present               string
	Absent                string
	Late                  string
	EarlyDeparture        string
	MissingInOut          string
	TotalHours            string
	AverageCheckIn        string
	AverageCheckOut       string
}

func departmentAttendanceReportLabelsForLanguage(language string) departmentAttendanceReportLabels {
	if isArabicLanguage(language) {
		return departmentAttendanceReportLabels{
			OverviewTitle:         "نظرة عامة على حضور القسم",
			DateRange:             "نطاق التاريخ",
			AttendanceRate:        "معدل الحضور",
			LateArrivalRate:       "معدل التأخر",
			MissingPunchRate:      "معدل البصمات المفقودة",
			TotalRecords:          "إجمالي السجلات",
			RecordsAnalyzed:       "السجلات المحللة",
			DaysCovered:           "الأيام المغطاة",
			TotalExceptions:       "إجمالي الاستثناءات",
			BestAttendanceDay:     "أفضل يوم حضور",
			WorstAttendanceDay:    "أضعف يوم حضور",
			TopExceptions:         "أعلى الاستثناءات",
			NoExceptions:          "لا توجد استثناءات",
			AttendanceTrendTitle:  "اتجاه الحضور",
			NoTrendData:           "لا توجد بيانات للاتجاه",
			TrendDate:             "التاريخ",
			TrendAttendanceRate:   "معدل الحضور",
			TrendAbsenceRate:      "معدل الغياب",
			TrendLateArrivalRate:  "معدل التأخر",
			TrendMissingPunchRate: "معدل البصمات المفقودة",
			TrendExceptions:       "الاستثناءات",
			TrendRecords:          "السجلات",
			EmployeeBreakdownTitle: "تفاصيل الموظفين",
			Employee:              "الموظف",
			WorkingDays:           "أيام العمل",
			Present:               "حضور",
			Absent:                "غياب",
			Late:                  "تأخير",
			EarlyDeparture:        "انصراف مبكر",
			MissingInOut:          "دخول/خروج مفقود",
			TotalHours:            "إجمالي الساعات",
			AverageCheckIn:        "متوسط الدخول",
			AverageCheckOut:       "متوسط الخروج",
		}
	}

	return departmentAttendanceReportLabels{
		OverviewTitle:         "Department attendance overview",
		DateRange:             "Date Range",
		AttendanceRate:        "Attendance Rate",
		LateArrivalRate:       "Late Arrival Rate",
		MissingPunchRate:      "Missing Punch Rate",
		TotalRecords:          "Total Records",
		RecordsAnalyzed:       "Records Analyzed",
		DaysCovered:           "Days Covered",
		TotalExceptions:       "Total Exceptions",
		BestAttendanceDay:     "Best Attendance Day",
		WorstAttendanceDay:    "Worst Attendance Day",
		TopExceptions:         "Top Exceptions",
		NoExceptions:          "No exceptions",
		AttendanceTrendTitle:  "Attendance Trend",
		NoTrendData:           "No trend data",
		TrendDate:             "Date",
		TrendAttendanceRate:   "Attendance Rate",
		TrendAbsenceRate:      "Absence Rate",
		TrendLateArrivalRate:  "Late Arrival Rate",
		TrendMissingPunchRate: "Missing Punch Rate",
		TrendExceptions:       "Exceptions",
		TrendRecords:          "Records",
		EmployeeBreakdownTitle: "Employee Breakdown",
		Employee:              "Employee",
		WorkingDays:           "Working Days",
		Present:               "Present",
		Absent:                "Absent",
		Late:                  "Late",
		EarlyDeparture:        "Early Departure",
		MissingInOut:          "Missing In/Out",
		TotalHours:            "Total Hours",
		AverageCheckIn:        "Average Check-In",
		AverageCheckOut:       "Average Check-Out",
	}
}

func departmentAttendanceReportSheetName(language string) string {
	if isArabicLanguage(language) {
		return "تقرير القسم"
	}
	return "Department Report"
}
