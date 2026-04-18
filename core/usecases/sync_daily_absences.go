package usecases

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type SyncDailyAbsencesOutput struct {
	Date                 string
	SyncedAbsenceCount   int
	ClearedAbsenceCount  int
	SkippedLeaveCount    int
	SkippedPresentCount  int
	SkippedHireDateCount int
	SkippedNonWorking    bool
}

type SyncDailyAbsencesUseCase struct {
	db          ports.DB
	leaveRepo   ports.LeaveRecordRepository
	weekendRepo ports.WeekendConfigRepository
	holidayRepo ports.HolidayDefinitionRepository
	location    *time.Location
}

func NewSyncDailyAbsencesUseCase(
	db ports.DB,
	leaveRepo ports.LeaveRecordRepository,
	weekendRepo ports.WeekendConfigRepository,
	holidayRepo ports.HolidayDefinitionRepository,
	timezone string,
) *SyncDailyAbsencesUseCase {
	loc, err := time.LoadLocation(strings.TrimSpace(timezone))
	if err != nil {
		loc = time.FixedZone("Africa/Cairo", 2*60*60)
	}

	return &SyncDailyAbsencesUseCase{
		db:          db,
		leaveRepo:   leaveRepo,
		weekendRepo: weekendRepo,
		holidayRepo: holidayRepo,
		location:    loc,
	}
}

func (uc *SyncDailyAbsencesUseCase) Execute(ctx context.Context) (*SyncDailyAbsencesOutput, error) {
	now := time.Now().In(uc.location)
	targetDate := normalizeDateOnly(now.AddDate(0, 0, -1))
	output := &SyncDailyAbsencesOutput{Date: targetDate.Format("2006-01-02")}

	checker, err := newNonWorkingDateChecker(ctx, uc.db, uc.weekendRepo, uc.holidayRepo, &targetDate, &targetDate)
	if err != nil {
		return nil, err
	}

	nonWorking, err := checker.IsNonWorking(targetDate)
	if err != nil {
		return nil, err
	}
	if nonWorking {
		cleared, err := deleteAbsenceExceptionsForDate(ctx, uc.db, targetDate, nil)
		if err != nil {
			return nil, err
		}
		output.ClearedAbsenceCount = cleared
		output.SkippedNonWorking = true
		return output, nil
	}

	employees, err := listActiveEmployeesForAbsenceSync(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	presentEmployeeUIDs, err := listPresentEmployeeUIDsForDate(ctx, uc.db, targetDate)
	if err != nil {
		return nil, err
	}

	absenceItems := make([]DailyAttendanceLogItem, 0, len(employees))
	absentEmployeeUIDs := make([]string, 0, len(employees))

	for _, employee := range employees {
		if employee.HireDate.After(targetDate) {
			output.SkippedHireDateCount++
			continue
		}

		if _, present := presentEmployeeUIDs[employee.UID]; present {
			output.SkippedPresentCount++
			continue
		}

		hasLeave, err := hasLeaveOnDateForAbsenceSync(ctx, uc.db, uc.leaveRepo, employee.ID, targetDate)
		if err != nil {
			return nil, err
		}
		if hasLeave {
			output.SkippedLeaveCount++
			continue
		}

		item := DailyAttendanceLogItem{
			Date:         targetDate,
			EmployeeUID:  employee.UID,
			EmployeeName: employee.Name,
			IsAbsent:     true,
			Exceptions: []domain.AttendanceExceptionType{
				domain.AttendanceExceptionTypeAbsence,
			},
		}
		if employee.DepartmentUID != nil {
			item.DepartmentUID = employee.DepartmentUID
		}

		absenceItems = append(absenceItems, item)
		absentEmployeeUIDs = append(absentEmployeeUIDs, employee.UID)
	}

	if len(absenceItems) > 0 {
		if err := syncAttendanceExceptions(ctx, uc.db, absenceItems); err != nil {
			return nil, err
		}
	}

	cleared, err := deleteAbsenceExceptionsForDate(ctx, uc.db, targetDate, absentEmployeeUIDs)
	if err != nil {
		return nil, err
	}

	output.SyncedAbsenceCount = len(absenceItems)
	output.ClearedAbsenceCount = cleared
	return output, nil
}

func deleteAbsenceExceptionsForDate(ctx context.Context, db ports.DB, targetDate time.Time, keepEmployeeUIDs []string) (int, error) {
	dateStr := targetDate.Format("2006-01-02")

	if len(keepEmployeeUIDs) == 0 {
		result, err := db.ExecContext(ctx,
			`DELETE FROM attendance_exceptions WHERE attendance_date = ? AND exception_type = ?`,
			dateStr,
			domain.AttendanceExceptionTypeAbsence,
		)
		if err != nil {
			return 0, err
		}
		affected, err := result.RowsAffected()
		if err != nil {
			return 0, err
		}
		return int(affected), nil
	}

	placeholders := make([]string, 0, len(keepEmployeeUIDs))
	args := make([]any, 0, len(keepEmployeeUIDs)+2)
	args = append(args, dateStr, domain.AttendanceExceptionTypeAbsence)
	for _, uid := range keepEmployeeUIDs {
		placeholders = append(placeholders, "?")
		args = append(args, uid)
	}

	query := fmt.Sprintf(
		`DELETE FROM attendance_exceptions WHERE attendance_date = ? AND exception_type = ? AND employee_uid NOT IN (%s)`,
		strings.Join(placeholders, ","),
	)

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return int(affected), nil
}

func listPresentEmployeeUIDsForDate(ctx context.Context, db ports.DB, targetDate time.Time) (map[string]struct{}, error) {
	rows, err := db.QueryContext(
		ctx,
		`SELECT DISTINCT employee_uid FROM attendance_records WHERE substr(punched_at, 1, 10) = ?`,
		targetDate.Format("2006-01-02"),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]struct{})
	for rows.Next() {
		var uid string
		if err := rows.Scan(&uid); err != nil {
			return nil, err
		}
		result[uid] = struct{}{}
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func listActiveEmployeesForAbsenceSync(ctx context.Context, db ports.DB) ([]departmentEmployeeForAttendance, error) {
	rows, err := db.QueryContext(ctx, `
		SELECT id, uid, name, department_uid, hire_date, status
		FROM employees
		WHERE status = ?
		ORDER BY uid ASC
	`, domain.EmployeeStatusActive)
	if err != nil {
		return nil, err
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

func hasLeaveOnDateForAbsenceSync(ctx context.Context, db ports.DB, leaveRepo ports.LeaveRecordRepository, employeeID int64, day time.Time) (bool, error) {
	if leaveRepo == nil {
		return false, nil
	}

	dateStr := day.Format("2006-01-02")
	var count int
	err := db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM leave_records WHERE employee_id = ? AND date(?) BETWEEN date(start_date) AND date(end_date)`,
		employeeID,
		dateStr,
	).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
