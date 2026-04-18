package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

const maxAbsenceRangeDays = 400

type departmentEmployeeForAttendance struct {
	ID            int64
	UID           string
	Name          string
	DepartmentUID *string
	HireDate      time.Time
	Status        domain.EmployeeStatus
}

type attendanceExceptionDetail struct {
	Type         domain.AttendanceExceptionType
	MinutesDelta *int
}

func buildExceptionTypes(item DailyAttendanceLogItem) []domain.AttendanceExceptionType {
	details := buildExceptionDetails(item)
	types := make([]domain.AttendanceExceptionType, 0, len(details))
	for _, detail := range details {
		types = append(types, detail.Type)
	}
	return types
}

func buildExceptionDetails(item DailyAttendanceLogItem) []attendanceExceptionDetail {
	exceptions := make([]attendanceExceptionDetail, 0, 4)

	if item.IsAbsent {
		return []attendanceExceptionDetail{{Type: domain.AttendanceExceptionTypeAbsence}}
	}
	if item.MissingCheckIn {
		exceptions = append(exceptions, attendanceExceptionDetail{Type: domain.AttendanceExceptionTypeMissedPunchIn})
	}
	if item.MissingCheckOut {
		exceptions = append(exceptions, attendanceExceptionDetail{Type: domain.AttendanceExceptionTypeMissedPunchOut})
	}
	if item.LateArrival {
		exceptions = append(exceptions, attendanceExceptionDetail{Type: domain.AttendanceExceptionTypeLateArrival, MinutesDelta: item.LateMinutes})
	}
	if item.EarlyDeparture {
		exceptions = append(exceptions, attendanceExceptionDetail{Type: domain.AttendanceExceptionTypeEarlyDeparture, MinutesDelta: item.EarlyMinutes})
	}

	return exceptions
}

func syncAttendanceExceptions(ctx context.Context, db ports.DB, items []DailyAttendanceLogItem) error {
	const upsertQuery = `
		INSERT INTO attendance_exceptions (
			uid, employee_uid, attendance_date, exception_type, check_in, check_out, grace_minutes, minutes_delta, created_at, updated_at
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(employee_uid, attendance_date, exception_type) DO UPDATE SET
			check_in = excluded.check_in,
			check_out = excluded.check_out,
			grace_minutes = excluded.grace_minutes,
			minutes_delta = excluded.minutes_delta,
			updated_at = excluded.updated_at`

	for _, item := range items {
		dateStr := item.Date.Format("2006-01-02")
		details := buildExceptionDetails(item)

		if err := deleteStaleAttendanceExceptions(ctx, db, item.EmployeeUID, dateStr, details); err != nil {
			return err
		}

		checkIn := nullableRFC3339(item.CheckIn)
		checkOut := nullableRFC3339(item.CheckOut)
		now := time.Now().Format(time.RFC3339Nano)

		for _, detail := range details {
			var grace any
			if item.GraceMinutes > 0 {
				grace = item.GraceMinutes
			}

			var minutesDelta any
			if detail.MinutesDelta != nil {
				minutesDelta = *detail.MinutesDelta
			}

			_, err := db.ExecContext(ctx, upsertQuery,
				domain.GenerateUID("aex"),
				item.EmployeeUID,
				dateStr,
				detail.Type,
				checkIn,
				checkOut,
				grace,
				minutesDelta,
				now,
				now,
			)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func deleteStaleAttendanceExceptions(ctx context.Context, db ports.DB, employeeUID, dateStr string, details []attendanceExceptionDetail) error {
	if len(details) == 0 {
		_, err := db.ExecContext(ctx, `DELETE FROM attendance_exceptions WHERE employee_uid = ? AND attendance_date = ?`, employeeUID, dateStr)
		return err
	}

	placeholders := make([]string, 0, len(details))
	args := make([]any, 0, len(details)+2)
	args = append(args, employeeUID, dateStr)
	for _, detail := range details {
		placeholders = append(placeholders, "?")
		args = append(args, detail.Type)
	}

	query := fmt.Sprintf(
		`DELETE FROM attendance_exceptions WHERE employee_uid = ? AND attendance_date = ? AND exception_type NOT IN (%s)`,
		strings.Join(placeholders, ","),
	)
	_, err := db.ExecContext(ctx, query, args...)
	return err
}

func nullableRFC3339(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format(time.RFC3339Nano)
}

func expandDateRange(startDate, endDate *time.Time) (time.Time, time.Time, bool) {
	if startDate == nil && endDate == nil {
		return time.Time{}, time.Time{}, false
	}

	var start time.Time
	var end time.Time
	if startDate != nil {
		start = normalizeDateOnly(*startDate)
	}
	if endDate != nil {
		end = normalizeDateOnly(*endDate)
	}

	if startDate == nil {
		start = end
	}
	if endDate == nil {
		end = start
	}
	if end.Before(start) {
		start, end = end, start
	}

	if countDaysInclusive(start, end) > maxAbsenceRangeDays {
		return time.Time{}, time.Time{}, false
	}

	return start, end, true
}

func countDaysInclusive(start, end time.Time) int {
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

func normalizeDateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.Local)
}

type nonWorkingDateChecker struct {
	holidayRepo ports.HolidayDefinitionRepository
	weekendSet  map[int]struct{}
	holidaySet  map[string]struct{}
	holidayMemo map[string]bool
	db          ports.DB
	ctx         context.Context
}

func newNonWorkingDateChecker(ctx context.Context, db ports.DB, weekendRepo ports.WeekendConfigRepository, holidayRepo ports.HolidayDefinitionRepository, start, end *time.Time) (*nonWorkingDateChecker, error) {
	checker := &nonWorkingDateChecker{
		holidayRepo: holidayRepo,
		weekendSet:  make(map[int]struct{}),
		holidaySet:  make(map[string]struct{}),
		holidayMemo: make(map[string]bool),
		db:          db,
		ctx:         ctx,
	}

	if weekendRepo == nil {
		return checker, nil
	}

	weekendDays, err := weekendRepo.GetWeekendDays(ctx, db)
	if err != nil {
		return nil, err
	}

	for _, day := range weekendDays {
		checker.weekendSet[day] = struct{}{}
	}

	if holidayRepo != nil && start != nil && end != nil {
		holidays, err := holidayRepo.ListByDateRange(ctx, db, normalizeDateOnly(*start), normalizeDateOnly(*end).Add(24*time.Hour-time.Nanosecond))
		if err != nil {
			return nil, err
		}

		for _, holiday := range holidays {
			checker.holidaySet[normalizeDateOnly(holiday.Date).Format("2006-01-02")] = struct{}{}
		}
	}

	return checker, nil
}

func (c *nonWorkingDateChecker) IsNonWorking(dateValue time.Time) (bool, error) {
	dateOnly := normalizeDateOnly(dateValue)
	if _, exists := c.weekendSet[int(dateOnly.Weekday())]; exists {
		return true, nil
	}

	dateKey := dateOnly.Format("2006-01-02")
	if _, exists := c.holidaySet[dateKey]; exists {
		return true, nil
	}

	if memoValue, exists := c.holidayMemo[dateKey]; exists {
		return memoValue, nil
	}
	if c.holidayRepo == nil {
		c.holidayMemo[dateKey] = false
		return false, nil
	}

	holidays, err := c.holidayRepo.ListByDateRange(c.ctx, c.db, dateOnly, dateOnly.Add(24*time.Hour-time.Nanosecond))
	if err != nil {
		return false, err
	}
	hasHoliday := len(holidays) > 0
	c.holidayMemo[dateKey] = hasHoliday
	if hasHoliday {
		c.holidaySet[dateKey] = struct{}{}
	}

	return hasHoliday, nil
}

func listDepartmentEmployeesForAttendance(ctx context.Context, db ports.DB, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) ([]departmentEmployeeForAttendance, error) {
	query := `
		SELECT id, uid, name, department_uid, hire_date, status
		FROM employees
		WHERE department_uid = ? AND status = ?`
	args := []any{departmentUID, domain.EmployeeStatusActive}

	if filter.EmployeeUID != nil {
		query += ` AND uid = ?`
		args = append(args, *filter.EmployeeUID)
	}

	if filter.EmployeeName != nil {
		mode := "contains"
		if filter.EmployeeNameMode != nil && *filter.EmployeeNameMode != "" {
			mode = *filter.EmployeeNameMode
		}

		switch mode {
		case "equals":
			query += ` AND name = ?`
			args = append(args, *filter.EmployeeName)
		case "contains":
			query += ` AND name LIKE ?`
			args = append(args, "%"+*filter.EmployeeName+"%")
		}
	}

	query += ` ORDER BY name ASC, uid ASC`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return nil, nil
	}
	defer rows.Close()

	employees := make([]departmentEmployeeForAttendance, 0)
	for rows.Next() {
		var employee departmentEmployeeForAttendance
		var deptUID sql.NullString
		var hireDate domain.Time

		if err := rows.Scan(&employee.ID, &employee.UID, &employee.Name, &deptUID, &hireDate, &employee.Status); err != nil {
			return nil, err
		}

		employee.HireDate = normalizeDateOnly(hireDate.Time)
		if deptUID.Valid {
			employee.DepartmentUID = &deptUID.String
		}

		employees = append(employees, employee)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return employees, nil
}

func appendDepartmentAbsenceItems(ctx context.Context, db ports.DB, leaveRepo ports.LeaveRecordRepository, weekendRepo ports.WeekendConfigRepository, holidayRepo ports.HolidayDefinitionRepository, items []DailyAttendanceLogItem, employees []departmentEmployeeForAttendance, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, start, end time.Time) ([]DailyAttendanceLogItem, error) {
	_ = leaveRepo
	_ = employees

	checker, err := newNonWorkingDateChecker(ctx, db, weekendRepo, holidayRepo, &start, &end)
	if err != nil {
		return nil, err
	}

	if filter.DeviceUID != nil || filter.PunchType != nil {
		return items, nil
	}

	query := `
		SELECT ae.attendance_date, e.uid, e.name, e.department_uid
		FROM attendance_exceptions ae
		INNER JOIN employees e ON e.uid = ae.employee_uid
		WHERE ae.exception_type = 'absence'
			AND e.department_uid = ?
			AND ae.attendance_date BETWEEN ? AND ?`
	args := []any{departmentUID, start.Format("2006-01-02"), end.Format("2006-01-02")}

	if filter.EmployeeUID != nil {
		query += ` AND e.uid = ?`
		args = append(args, *filter.EmployeeUID)
	}
	if filter.EmployeeName != nil {
		mode := "contains"
		if filter.EmployeeNameMode != nil && *filter.EmployeeNameMode != "" {
			mode = *filter.EmployeeNameMode
		}

		switch mode {
		case "equals":
			query += ` AND e.name = ?`
			args = append(args, *filter.EmployeeName)
		case "contains":
			query += ` AND e.name LIKE ?`
			args = append(args, "%"+*filter.EmployeeName+"%")
		}
	}

	query += ` ORDER BY ae.attendance_date ASC, e.name ASC, e.uid ASC`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return items, nil
	}
	defer rows.Close()

	itemByKey := make(map[string]struct{}, len(items))
	for _, item := range items {
		itemByKey[dailyAttendanceKey(item.EmployeeUID, item.Date)] = struct{}{}
	}

	for rows.Next() {
		var dateValue string
		var employeeUID string
		var employeeName string
		var departmentUIDValue sql.NullString

		if err := rows.Scan(&dateValue, &employeeUID, &employeeName, &departmentUIDValue); err != nil {
			return nil, err
		}

		dateParsed, err := time.Parse("2006-01-02", dateValue)
		if err != nil {
			return nil, err
		}

		nonWorking, err := checker.IsNonWorking(dateParsed)
		if err != nil {
			return nil, err
		}
		if nonWorking {
			continue
		}

		key := dailyAttendanceKey(employeeUID, dateParsed)
		if _, exists := itemByKey[key]; exists {
			continue
		}

		absentItem := DailyAttendanceLogItem{
			Date:         dateParsed,
			EmployeeUID:  employeeUID,
			EmployeeName: employeeName,
			IsAbsent:     true,
			Exceptions: []domain.AttendanceExceptionType{
				domain.AttendanceExceptionTypeAbsence,
			},
		}
		if departmentUIDValue.Valid {
			absentItem.DepartmentUID = &departmentUIDValue.String
		}

		itemByKey[key] = struct{}{}
		items = append(items, absentItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func appendDepartmentAbsenceItemsWithoutRange(ctx context.Context, db ports.DB, leaveRepo ports.LeaveRecordRepository, weekendRepo ports.WeekendConfigRepository, holidayRepo ports.HolidayDefinitionRepository, items []DailyAttendanceLogItem, employees []departmentEmployeeForAttendance, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) ([]DailyAttendanceLogItem, error) {
	_ = leaveRepo

	checker, err := newNonWorkingDateChecker(ctx, db, weekendRepo, holidayRepo, nil, nil)
	if err != nil {
		return nil, err
	}

	if filter.DeviceUID != nil || filter.PunchType != nil {
		return items, nil
	}

	query := `
		SELECT ae.attendance_date, e.uid, e.name, e.department_uid
		FROM attendance_exceptions ae
		INNER JOIN employees e ON e.uid = ae.employee_uid
		WHERE ae.exception_type = 'absence'
			AND e.department_uid = ?`
	args := []any{departmentUID}

	if filter.EmployeeUID != nil {
		query += ` AND e.uid = ?`
		args = append(args, *filter.EmployeeUID)
	}
	if filter.EmployeeName != nil {
		mode := "contains"
		if filter.EmployeeNameMode != nil && *filter.EmployeeNameMode != "" {
			mode = *filter.EmployeeNameMode
		}

		switch mode {
		case "equals":
			query += ` AND e.name = ?`
			args = append(args, *filter.EmployeeName)
		case "contains":
			query += ` AND e.name LIKE ?`
			args = append(args, "%"+*filter.EmployeeName+"%")
		}
	}

	query += ` ORDER BY ae.attendance_date ASC, e.name ASC, e.uid ASC`

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return items, nil
	}
	defer rows.Close()

	itemByKey := make(map[string]struct{}, len(items))
	for _, item := range items {
		itemByKey[dailyAttendanceKey(item.EmployeeUID, item.Date)] = struct{}{}
	}

	visibleDates := visibleAttendanceDates(items)
	if len(visibleDates) > 0 {
		if len(employees) == 0 {
			employees, err = listDepartmentEmployeesForAttendance(ctx, db, departmentUID, filter)
			if err != nil {
				return nil, err
			}
		}

		for _, dateValue := range visibleDates {
			nonWorking, err := checker.IsNonWorking(dateValue)
			if err != nil {
				return nil, err
			}
			if nonWorking {
				continue
			}

			for _, employee := range employees {
				if employee.HireDate.After(dateValue) {
					continue
				}

				hasLeave, err := hasLeaveOnAttendanceDay(ctx, db, leaveRepo, employee.ID, dateValue)
				if err != nil {
					return nil, err
				}
				if hasLeave {
					continue
				}

				key := dailyAttendanceKey(employee.UID, dateValue)
				if _, exists := itemByKey[key]; exists {
					continue
				}

				itemByKey[key] = struct{}{}
				items = append(items, DailyAttendanceLogItem{
					Date:          dateValue,
					EmployeeUID:   employee.UID,
					EmployeeName:  employee.Name,
					DepartmentUID: employee.DepartmentUID,
					IsAbsent:      true,
					Exceptions: []domain.AttendanceExceptionType{
						domain.AttendanceExceptionTypeAbsence,
					},
				})
			}
		}
	}

	for rows.Next() {
		var dateValue string
		var employeeUID string
		var employeeName string
		var departmentUIDValue sql.NullString

		if err := rows.Scan(&dateValue, &employeeUID, &employeeName, &departmentUIDValue); err != nil {
			return nil, err
		}

		dateParsed, err := time.Parse("2006-01-02", dateValue)
		if err != nil {
			return nil, err
		}

		nonWorking, err := checker.IsNonWorking(dateParsed)
		if err != nil {
			return nil, err
		}
		if nonWorking {
			continue
		}

		key := dailyAttendanceKey(employeeUID, dateParsed)
		if _, exists := itemByKey[key]; exists {
			continue
		}

		absentItem := DailyAttendanceLogItem{
			Date:         dateParsed,
			EmployeeUID:  employeeUID,
			EmployeeName: employeeName,
			IsAbsent:     true,
			Exceptions: []domain.AttendanceExceptionType{
				domain.AttendanceExceptionTypeAbsence,
			},
		}
		if departmentUIDValue.Valid {
			absentItem.DepartmentUID = &departmentUIDValue.String
		}

		itemByKey[key] = struct{}{}
		items = append(items, absentItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func appendEmployeeAbsenceItems(ctx context.Context, db ports.DB, leaveRepo ports.LeaveRecordRepository, weekendRepo ports.WeekendConfigRepository, holidayRepo ports.HolidayDefinitionRepository, items []DailyAttendanceLogItem, employee *domain.Employee, start, end time.Time) ([]DailyAttendanceLogItem, error) {
	_ = leaveRepo

	checker, err := newNonWorkingDateChecker(ctx, db, weekendRepo, holidayRepo, &start, &end)
	if err != nil {
		return nil, err
	}

	if employee == nil || employee.Status != domain.EmployeeStatusActive {
		return items, nil
	}

	query := `
		SELECT ae.attendance_date, e.uid, e.name, e.department_uid
		FROM attendance_exceptions ae
		INNER JOIN employees e ON e.uid = ae.employee_uid
		WHERE ae.exception_type = 'absence'
			AND e.uid = ?
			AND ae.attendance_date BETWEEN ? AND ?
		ORDER BY ae.attendance_date ASC`
	args := []any{employee.UID, start.Format("2006-01-02"), end.Format("2006-01-02")}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itemByKey := make(map[string]struct{}, len(items))
	for _, item := range items {
		itemByKey[dailyAttendanceKey(item.EmployeeUID, item.Date)] = struct{}{}
	}

	for rows.Next() {
		var dateValue string
		var employeeUID string
		var employeeName string
		var departmentUID sql.NullString

		if err := rows.Scan(&dateValue, &employeeUID, &employeeName, &departmentUID); err != nil {
			return nil, err
		}

		dateParsed, err := time.Parse("2006-01-02", dateValue)
		if err != nil {
			return nil, err
		}

		nonWorking, err := checker.IsNonWorking(dateParsed)
		if err != nil {
			return nil, err
		}
		if nonWorking {
			continue
		}

		key := dailyAttendanceKey(employeeUID, dateParsed)
		if _, exists := itemByKey[key]; exists {
			continue
		}

		absentItem := DailyAttendanceLogItem{
			Date:          dateParsed,
			EmployeeUID:   employeeUID,
			EmployeeName:  employeeName,
			DepartmentUID: nil,
			IsAbsent:      true,
			Exceptions: []domain.AttendanceExceptionType{
				domain.AttendanceExceptionTypeAbsence,
			},
		}
		if departmentUID.Valid {
			absentItem.DepartmentUID = &departmentUID.String
		}

		itemByKey[key] = struct{}{}
		items = append(items, absentItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func appendEmployeeAbsenceItemsWithoutRange(ctx context.Context, db ports.DB, leaveRepo ports.LeaveRecordRepository, weekendRepo ports.WeekendConfigRepository, holidayRepo ports.HolidayDefinitionRepository, items []DailyAttendanceLogItem, employee *domain.Employee) ([]DailyAttendanceLogItem, error) {
	_ = leaveRepo

	checker, err := newNonWorkingDateChecker(ctx, db, weekendRepo, holidayRepo, nil, nil)
	if err != nil {
		return nil, err
	}

	if employee == nil || employee.Status != domain.EmployeeStatusActive {
		return items, nil
	}

	query := `
		SELECT ae.attendance_date, e.uid, e.name, e.department_uid
		FROM attendance_exceptions ae
		INNER JOIN employees e ON e.uid = ae.employee_uid
		WHERE ae.exception_type = 'absence'
			AND e.uid = ?
		ORDER BY ae.attendance_date ASC`

	rows, err := db.QueryContext(ctx, query, employee.UID)
	if err != nil {
		return nil, err
	}
	if rows == nil {
		return items, nil
	}
	if rows == nil {
		return items, nil
	}
	defer rows.Close()

	itemByKey := make(map[string]struct{}, len(items))
	for _, item := range items {
		itemByKey[dailyAttendanceKey(item.EmployeeUID, item.Date)] = struct{}{}
	}

	visibleDates := visibleAttendanceDates(items)
	for _, dateValue := range visibleDates {
		nonWorking, err := checker.IsNonWorking(dateValue)
		if err != nil {
			return nil, err
		}
		if nonWorking {
			continue
		}

		hasLeave, err := hasLeaveOnAttendanceDay(ctx, db, leaveRepo, employee.ID, dateValue)
		if err != nil {
			return nil, err
		}
		if hasLeave || employee.HireDate.After(dateValue) {
			continue
		}

		key := dailyAttendanceKey(employee.UID, dateValue)
		if _, exists := itemByKey[key]; exists {
			continue
		}

		itemByKey[key] = struct{}{}
		items = append(items, DailyAttendanceLogItem{
			Date:         dateValue,
			EmployeeUID:  employee.UID,
			EmployeeName: employee.Name,
			IsAbsent:     true,
			Exceptions: []domain.AttendanceExceptionType{
				domain.AttendanceExceptionTypeAbsence,
			},
		})
	}

	for rows.Next() {
		var dateValue string
		var employeeUID string
		var employeeName string
		var departmentUID sql.NullString

		if err := rows.Scan(&dateValue, &employeeUID, &employeeName, &departmentUID); err != nil {
			return nil, err
		}

		dateParsed, err := time.Parse("2006-01-02", dateValue)
		if err != nil {
			return nil, err
		}

		nonWorking, err := checker.IsNonWorking(dateParsed)
		if err != nil {
			return nil, err
		}
		if nonWorking {
			continue
		}

		key := dailyAttendanceKey(employeeUID, dateParsed)
		if _, exists := itemByKey[key]; exists {
			continue
		}

		absentItem := DailyAttendanceLogItem{
			Date:         dateParsed,
			EmployeeUID:  employeeUID,
			EmployeeName: employeeName,
			IsAbsent:     true,
			Exceptions: []domain.AttendanceExceptionType{
				domain.AttendanceExceptionTypeAbsence,
			},
		}
		if departmentUID.Valid {
			absentItem.DepartmentUID = &departmentUID.String
		}

		itemByKey[key] = struct{}{}
		items = append(items, absentItem)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func visibleAttendanceDates(items []DailyAttendanceLogItem) []time.Time {
	seen := make(map[string]time.Time, len(items))
	for _, item := range items {
		dateOnly := normalizeDateOnly(item.Date)
		seen[dateOnly.Format("2006-01-02")] = dateOnly
	}

	dates := make([]time.Time, 0, len(seen))
	for _, value := range seen {
		dates = append(dates, value)
	}

	sort.Slice(dates, func(i, j int) bool {
		return dates[i].Before(dates[j])
	})

	return dates
}

func dailyAttendanceKey(employeeUID string, date time.Time) string {
	return employeeUID + "|" + date.Format("2006-01-02")
}

func sortDailyAttendanceItems(items []DailyAttendanceLogItem, params ports.ListParams) {
	sort.SliceStable(items, func(i, j int) bool {
		left := items[i]
		right := items[j]

		if params.SortBy == "employeeName" {
			cmp := strings.Compare(left.EmployeeName, right.EmployeeName)
			if cmp != 0 {
				if params.SortOrder == ports.SortOrderDesc {
					return cmp > 0
				}
				return cmp < 0
			}

			if !left.Date.Equal(right.Date) {
				if params.SortOrder == ports.SortOrderDesc {
					return left.Date.After(right.Date)
				}
				return left.Date.Before(right.Date)
			}

			return left.EmployeeUID < right.EmployeeUID
		}

		if !left.Date.Equal(right.Date) {
			if params.SortOrder == ports.SortOrderDesc {
				return left.Date.After(right.Date)
			}
			return left.Date.Before(right.Date)
		}

		cmp := strings.Compare(left.EmployeeName, right.EmployeeName)
		if cmp != 0 {
			return cmp < 0
		}

		return left.EmployeeUID < right.EmployeeUID
	})
}

func paginateDailyAttendanceItems(items []DailyAttendanceLogItem, params ports.ListParams) []DailyAttendanceLogItem {
	if len(items) == 0 {
		return []DailyAttendanceLogItem{}
	}

	start := params.Offset()
	if start >= len(items) {
		return []DailyAttendanceLogItem{}
	}

	end := start + params.PageSize
	if end > len(items) {
		end = len(items)
	}

	return slices.Clone(items[start:end])
}

func shouldRecordAbsenceForDay(day time.Time, now time.Time) bool {
	dayOnly := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, now.Location())
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	return dayOnly.Before(today)
}
