package usecases

import (
	"context"
	"sort"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type WorkingHoursByDay struct {
	Date        time.Time
	WorkedHours float64
}

type AttendanceDayBreakdown struct {
	Date                  time.Time
	CheckIn               *time.Time
	CheckOut              *time.Time
	WorkedHours           float64
	IsLate                bool
	LateMinutes           int
	IsEarlyDeparture      bool
	EarlyDepartureMinutes int
	IsAbsent              bool
	IsOnLeave             bool
	LeaveTypeName         *string
}

type GetMonthlyAttendanceStatsInput struct {
	EmployeeUID string
	StartDate   *time.Time
	EndDate     *time.Time
	PeriodLabel *string
}

type GetMonthlyAttendanceStatsOutput struct {
	EmployeeName            string
	Month                   string
	TotalWorkedHours        float64
	AverageCheckInTime      *time.Time
	AverageCheckOutTime     *time.Time
	MissingCheckInCount     int
	MissingCheckOutCount    int
	WorkingHoursByDay       []WorkingHoursByDay
	LateDaysCount           int
	EarlyDepartureDaysCount int
	AbsentDaysCount         int
	DaysBreakdown           []AttendanceDayBreakdown
}

type MonthlyAttendanceStatsDependencies struct {
	EmployeeRepo ports.EmployeeRepository
	DeptRepo     ports.DepartmentRepository
	ShiftRepo    ports.ShiftRepository
	WeekendRepo  ports.WeekendConfigRepository
	HolidayRepo  ports.HolidayDefinitionRepository
}

type GetMonthlyAttendanceStatsUseCase struct {
	db           ports.DB
	recordRepo   ports.AttendanceRecordRepository
	employeeRepo ports.EmployeeRepository
	deptRepo     ports.DepartmentRepository
	shiftRepo    ports.ShiftRepository
	weekendRepo  ports.WeekendConfigRepository
	holidayRepo  ports.HolidayDefinitionRepository
	now          func() time.Time
}

func NewGetMonthlyAttendanceStatsUseCase(
	db ports.DB,
	recordRepo ports.AttendanceRecordRepository,
	deps ...MonthlyAttendanceStatsDependencies,
) *GetMonthlyAttendanceStatsUseCase {
	uc := &GetMonthlyAttendanceStatsUseCase{
		db:         db,
		recordRepo: recordRepo,
		now:        time.Now,
	}
	if len(deps) > 0 {
		uc.employeeRepo = deps[0].EmployeeRepo
		uc.deptRepo = deps[0].DeptRepo
		uc.shiftRepo = deps[0].ShiftRepo
		uc.weekendRepo = deps[0].WeekendRepo
		uc.holidayRepo = deps[0].HolidayRepo
	}
	return uc
}

func (uc *GetMonthlyAttendanceStatsUseCase) Execute(ctx context.Context, input GetMonthlyAttendanceStatsInput) (*GetMonthlyAttendanceStatsOutput, error) {
	current := uc.now()
	monthStart := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
	monthEnd := time.Date(current.Year(), current.Month(), current.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), current.Location())

	rangeStart := monthStart
	rangeEnd := monthEnd

	if input.StartDate != nil || input.EndDate != nil {
		if input.StartDate != nil {
			rangeStart = time.Date(input.StartDate.Year(), input.StartDate.Month(), input.StartDate.Day(), 0, 0, 0, 0, input.StartDate.Location())
		}
		if input.EndDate != nil {
			rangeEnd = time.Date(input.EndDate.Year(), input.EndDate.Month(), input.EndDate.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), input.EndDate.Location())
		}
		if input.StartDate == nil {
			rangeStart = time.Date(rangeEnd.Year(), rangeEnd.Month(), rangeEnd.Day(), 0, 0, 0, 0, rangeEnd.Location())
		}
		if input.EndDate == nil {
			rangeEnd = time.Date(rangeStart.Year(), rangeStart.Month(), rangeStart.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), rangeStart.Location())
		}
		if rangeEnd.Before(rangeStart) {
			rangeStart, rangeEnd = rangeEnd, rangeStart
			rangeStart = time.Date(rangeStart.Year(), rangeStart.Month(), rangeStart.Day(), 0, 0, 0, 0, rangeStart.Location())
			rangeEnd = time.Date(rangeEnd.Year(), rangeEnd.Month(), rangeEnd.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), rangeEnd.Location())
		}
	}

	employeeUID := input.EmployeeUID
	var employeeName string
	var employeeDepartmentUID *string
	var employeeHireDate *time.Time
	var employeeID int64
	if uc.employeeRepo != nil {
		employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, employeeUID)
		if err != nil {
			return nil, err
		}
		if employee == nil {
			return nil, ErrEmployeeNotFound
		}
		employeeName = employee.Name
		employeeDepartmentUID = employee.DepartmentUID
		employeeID = employee.ID
		hireDate := normalizeDateOnly(employee.HireDate)
		employeeHireDate = &hireDate
	}

	records, err := uc.recordRepo.ListByDateRange(ctx, uc.db, rangeStart, rangeEnd, &employeeUID)
	if err != nil {
		return nil, err
	}

	byEmployeeDay := make(map[string]*monthlyAttendanceDaySummary)
	for _, rec := range records {
		day := time.Date(rec.PunchedAt.Year(), rec.PunchedAt.Month(), rec.PunchedAt.Day(), 0, 0, 0, 0, rec.PunchedAt.Location())
		key := rec.EmployeeUID + "|" + day.Format("2006-01-02")

		item, ok := byEmployeeDay[key]
		if !ok {
			item = &monthlyAttendanceDaySummary{
				date:       day,
				employeeID: rec.EmployeeUID,
			}
			byEmployeeDay[key] = item
		}

		switch rec.PunchType {
		case domain.AttendancePunchTypeCheckIn:
			if item.checkIn == nil || rec.PunchedAt.Before(*item.checkIn) {
				t := rec.PunchedAt
				item.checkIn = &t
			}
		case domain.AttendancePunchTypeCheckOut:
			if item.checkOut == nil || rec.PunchedAt.After(*item.checkOut) {
				t := rec.PunchedAt
				item.checkOut = &t
			}
		case domain.AttendancePunchTypeUnknown:
			if item.checkIn == nil || rec.PunchedAt.Before(*item.checkIn) {
				t := rec.PunchedAt
				item.checkIn = &t
			}
			if item.checkOut == nil || rec.PunchedAt.After(*item.checkOut) {
				t := rec.PunchedAt
				item.checkOut = &t
			}
		}
	}

	totalWorkedHours := 0.0
	missingCheckInCount := 0
	missingCheckOutCount := 0
	totalCheckInSeconds := 0
	totalCheckOutSeconds := 0
	checkInCount := 0
	checkOutCount := 0
	workedHoursByDay := make(map[string]float64)

	for _, item := range byEmployeeDay {
		if item.checkIn == nil {
			missingCheckInCount++
		}
		if item.checkIn != nil {
			totalCheckInSeconds += (item.checkIn.Hour() * 3600) + (item.checkIn.Minute() * 60) + item.checkIn.Second()
			checkInCount++
		}
		if item.checkOut == nil {
			missingCheckOutCount++
		}
		if item.checkOut != nil {
			totalCheckOutSeconds += (item.checkOut.Hour() * 3600) + (item.checkOut.Minute() * 60) + item.checkOut.Second()
			checkOutCount++
		}
		if item.checkIn != nil && item.checkOut != nil && item.checkOut.After(*item.checkIn) {
			worked := item.checkOut.Sub(*item.checkIn).Hours()
			totalWorkedHours += worked
			workedHoursByDay[item.date.Format("2006-01-02")] += worked
		}
	}

	dailyItemsByDate, err := uc.buildMonthlyDailyItems(ctx, byEmployeeDay, employeeName, employeeDepartmentUID)
	if err != nil {
		return nil, err
	}

	var averageCheckInTime *time.Time
	if checkInCount > 0 {
		avgSeconds := totalCheckInSeconds / checkInCount
		t := time.Date(rangeStart.Year(), rangeStart.Month(), rangeStart.Day(), avgSeconds/3600, (avgSeconds%3600)/60, avgSeconds%60, 0, rangeStart.Location())
		averageCheckInTime = &t
	}

	var averageCheckOutTime *time.Time
	if checkOutCount > 0 {
		avgSeconds := totalCheckOutSeconds / checkOutCount
		t := time.Date(rangeStart.Year(), rangeStart.Month(), rangeStart.Day(), avgSeconds/3600, (avgSeconds%3600)/60, avgSeconds%60, 0, rangeStart.Location())
		averageCheckOutTime = &t
	}

	daysInRange := int(rangeEnd.Sub(time.Date(rangeStart.Year(), rangeStart.Month(), rangeStart.Day(), 0, 0, 0, 0, rangeStart.Location())).Hours()/24) + 1
	if daysInRange < 0 {
		daysInRange = 0
	}

	daily := make([]WorkingHoursByDay, 0, daysInRange)
	for d := time.Date(rangeStart.Year(), rangeStart.Month(), rangeStart.Day(), 0, 0, 0, 0, rangeStart.Location()); !d.After(rangeEnd); d = d.AddDate(0, 0, 1) {
		day := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
		daily = append(daily, WorkingHoursByDay{
			Date:        day,
			WorkedHours: workedHoursByDay[day.Format("2006-01-02")],
		})
	}

	sort.Slice(daily, func(i, j int) bool {
		return daily[i].Date.Before(daily[j].Date)
	})

	rangeStartDay := normalizeDateOnly(rangeStart)
	rangeEndDay := normalizeDateOnly(rangeEnd)
	daysBreakdown, lateDaysCount, earlyDepartureDaysCount, absentDaysCount, err := uc.buildWorkingDaysBreakdown(
		ctx,
		employeeUID,
		employeeName,
		employeeDepartmentUID,
		employeeID,
		employeeHireDate,
		rangeStartDay,
		rangeEndDay,
		dailyItemsByDate,
	)
	if err != nil {
		return nil, err
	}

	period := rangeStart.Format("2006-01")
	if input.PeriodLabel != nil && *input.PeriodLabel != "" {
		period = *input.PeriodLabel
	}

	return &GetMonthlyAttendanceStatsOutput{
		EmployeeName:            employeeName,
		Month:                   period,
		TotalWorkedHours:        totalWorkedHours,
		AverageCheckInTime:      averageCheckInTime,
		AverageCheckOutTime:     averageCheckOutTime,
		MissingCheckInCount:     missingCheckInCount,
		MissingCheckOutCount:    missingCheckOutCount,
		WorkingHoursByDay:       daily,
		LateDaysCount:           lateDaysCount,
		EarlyDepartureDaysCount: earlyDepartureDaysCount,
		AbsentDaysCount:         absentDaysCount,
		DaysBreakdown:           daysBreakdown,
	}, nil
}

type monthlyAttendanceDaySummary struct {
	date       time.Time
	checkIn    *time.Time
	checkOut   *time.Time
	employeeID string
}

func (uc *GetMonthlyAttendanceStatsUseCase) buildMonthlyDailyItems(ctx context.Context, summaries map[string]*monthlyAttendanceDaySummary, employeeName string, departmentUID *string) (map[string]DailyAttendanceLogItem, error) {
	itemsByDate := make(map[string]DailyAttendanceLogItem, len(summaries))
	if len(summaries) == 0 {
		return itemsByDate, nil
	}

	groups := make([]*ports.DailyAttendanceGroup, 0, len(summaries))
	for _, summary := range summaries {
		groups = append(groups, &ports.DailyAttendanceGroup{
			Date:          summary.date,
			EmployeeUID:   summary.employeeID,
			EmployeeName:  employeeName,
			DepartmentUID: departmentUID,
			CheckIn:       summary.checkIn,
			CheckOut:      summary.checkOut,
		})
	}
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Date.Before(groups[j].Date)
	})

	if uc.shiftRepo == nil {
		for _, group := range groups {
			item := DailyAttendanceLogItem{
				Date:            group.Date,
				EmployeeUID:     group.EmployeeUID,
				EmployeeName:    group.EmployeeName,
				DepartmentUID:   group.DepartmentUID,
				CheckIn:         group.CheckIn,
				CheckOut:        group.CheckOut,
				MissingCheckIn:  group.CheckIn == nil && group.CheckOut != nil,
				MissingCheckOut: group.CheckOut == nil && group.CheckIn != nil,
			}
			if group.CheckIn != nil && group.CheckOut != nil && group.CheckOut.After(*group.CheckIn) {
				item.WorkedHours = group.CheckOut.Sub(*group.CheckIn).Hours()
			}
			itemsByDate[group.Date.Format("2006-01-02")] = item
		}
		return itemsByDate, nil
	}

	items, err := buildDailyAttendanceLogItems(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, groups)
	if err != nil {
		return nil, err
	}
	for _, item := range items {
		itemsByDate[item.Date.Format("2006-01-02")] = item
	}

	return itemsByDate, nil
}

func (uc *GetMonthlyAttendanceStatsUseCase) buildWorkingDaysBreakdown(
	ctx context.Context,
	employeeUID string,
	employeeName string,
	departmentUID *string,
	employeeID int64,
	employeeHireDate *time.Time,
	rangeStart time.Time,
	rangeEnd time.Time,
	dailyItemsByDate map[string]DailyAttendanceLogItem,
) ([]AttendanceDayBreakdown, int, int, int, error) {
	if rangeEnd.Before(rangeStart) {
		return []AttendanceDayBreakdown{}, 0, 0, 0, nil
	}

	checker, err := newNonWorkingDateChecker(ctx, uc.db, uc.weekendRepo, uc.holidayRepo, &rangeStart, &rangeEnd)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	leaveTypeNamesByDate := map[string]string{}
	if employeeID > 0 {
		leaveTypeNamesByDate, err = uc.listLeaveTypeNamesByDate(ctx, employeeID, rangeStart, rangeEnd)
		if err != nil {
			return nil, 0, 0, 0, err
		}
	}

	absenceShift := domain.NewDefaultShift()
	if uc.shiftRepo != nil {
		shift, err := resolveEffectiveShift(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, employeeUID, departmentUID)
		if err != nil {
			return nil, 0, 0, 0, err
		}
		if shift != nil {
			absenceShift = shift
		}
	}

	breakdown := make([]AttendanceDayBreakdown, 0, countDaysInclusive(rangeStart, rangeEnd))
	lateDaysCount := 0
	earlyDepartureDaysCount := 0
	absentDaysCount := 0
	now := uc.now()

	for day := rangeStart; !day.After(rangeEnd); day = day.AddDate(0, 0, 1) {
		if employeeHireDate != nil && employeeHireDate.After(day) {
			continue
		}

		nonWorking, err := checker.IsNonWorking(day)
		if err != nil {
			return nil, 0, 0, 0, err
		}
		if nonWorking {
			continue
		}

		dateKey := day.Format("2006-01-02")
		item, ok := dailyItemsByDate[dateKey]
		if !ok {
			item = DailyAttendanceLogItem{
				Date:          day,
				EmployeeUID:   employeeUID,
				EmployeeName:  employeeName,
				DepartmentUID: departmentUID,
			}
		}

		var leaveTypeName *string
		if value, ok := leaveTypeNamesByDate[dateKey]; ok {
			v := value
			leaveTypeName = &v
		}

		workEnd, err := buildTimeOnDate(day, absenceShift.EndTime)
		if err != nil {
			return nil, 0, 0, 0, ErrInvalidWorkDayEnd
		}
		isAbsent := employeeHireDate != nil &&
			item.CheckIn == nil &&
			item.CheckOut == nil &&
			leaveTypeName == nil &&
			isLeaveAwareAbsence(day, now, workEnd)

		if item.LateArrival {
			lateDaysCount++
		}
		if item.EarlyDeparture {
			earlyDepartureDaysCount++
		}
		if isAbsent {
			absentDaysCount++
		}

		breakdown = append(breakdown, AttendanceDayBreakdown{
			Date:                  day,
			CheckIn:               item.CheckIn,
			CheckOut:              item.CheckOut,
			WorkedHours:           item.WorkedHours,
			IsLate:                item.LateArrival,
			LateMinutes:           intPointerValue(item.LateMinutes),
			IsEarlyDeparture:      item.EarlyDeparture,
			EarlyDepartureMinutes: intPointerValue(item.EarlyMinutes),
			IsAbsent:              isAbsent,
			IsOnLeave:             leaveTypeName != nil,
			LeaveTypeName:         leaveTypeName,
		})
	}

	return breakdown, lateDaysCount, earlyDepartureDaysCount, absentDaysCount, nil
}

func (uc *GetMonthlyAttendanceStatsUseCase) listLeaveTypeNamesByDate(ctx context.Context, employeeID int64, start, end time.Time) (map[string]string, error) {
	result := make(map[string]string)
	if employeeID == 0 {
		return result, nil
	}

	rows, err := uc.db.QueryContext(ctx, `
		SELECT TO_CHAR(lr.start_date, 'YYYY-MM-DD'), TO_CHAR(lr.end_date, 'YYYY-MM-DD'), lt.name_en
		FROM leave_records lr
		INNER JOIN leave_types lt ON lt.id = lr.leave_type_id
		WHERE lr.employee_id = $1
			AND lr.start_date <= $2
			AND lr.end_date >= $3
		ORDER BY lr.start_date ASC, lr.uid ASC
	`, employeeID, end.Format("2006-01-02"), start.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var leaveStartValue string
		var leaveEndValue string
		var leaveTypeName string

		if err := rows.Scan(&leaveStartValue, &leaveEndValue, &leaveTypeName); err != nil {
			return nil, err
		}

		leaveStart, err := parseDateOrTimestamp(leaveStartValue)
		if err != nil {
			return nil, err
		}
		leaveEnd, err := parseDateOrTimestamp(leaveEndValue)
		if err != nil {
			return nil, err
		}

		if leaveStart.Before(start) {
			leaveStart = start
		}
		if leaveEnd.After(end) {
			leaveEnd = end
		}

		for day := leaveStart; !day.After(leaveEnd); day = day.AddDate(0, 0, 1) {
			dateKey := day.Format("2006-01-02")
			if _, exists := result[dateKey]; !exists {
				result[dateKey] = leaveTypeName
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func intPointerValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
