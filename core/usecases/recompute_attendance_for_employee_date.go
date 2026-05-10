package usecases

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var recomputeAttendanceForEmployeeDate = func(
	ctx context.Context,
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	permRepo ports.PermissionRequestRepository,
	employeeUID string,
	date time.Time,
	knownShift *domain.Shift,
) error {
	if employeeRepo == nil || shiftRepo == nil {
		return nil
	}

	employee, err := employeeRepo.GetByUID(ctx, db, employeeUID)
	if err != nil {
		slog.Error("recompute_attendance.get_employee", "error", err, "employee_uid", employeeUID)
		return err
	}
	if employee == nil {
		return nil
	}

	shift := knownShift
	if shift == nil {
		shift, err = resolveEffectiveShift(ctx, db, employeeRepo, deptRepo, shiftRepo, employeeUID, employee.DepartmentUID)
		if err != nil {
			slog.Error("recompute_attendance.resolve_shift", "error", err, "employee_uid", employeeUID)
			return err
		}
	}

	day := normalizeDateOnly(date)

	// Re-read raw punches for that day; if none, clear any stale exception rows for the day.
	checkIn, checkOut, err := loadCheckInOutForDay(ctx, db, employeeUID, day)
	if err != nil {
		slog.Error("recompute_attendance.load_punches", "error", err, "employee_uid", employeeUID)
		return err
	}

	adjuster, err := newPermissionAdjusterForDate(ctx, db, permRepo, employeeUID, day, shift)
	if err != nil {
		slog.Error("recompute_attendance.adjuster", "error", err, "employee_uid", employeeUID)
		return err
	}

	item := computeAttendanceItem(employee, day, shift, checkIn, checkOut, adjuster)

	return syncAttendanceExceptions(ctx, db, []DailyAttendanceLogItem{item})
}

// computeAttendanceItem mirrors the per-day classification done in buildDailyAttendanceLogItems
// for a single (employee, date) pair, applying the permissionAdjuster.
func computeAttendanceItem(employee *domain.Employee, day time.Time, shift *domain.Shift, checkIn, checkOut *time.Time, adjuster *permissionAdjuster) DailyAttendanceLogItem {
	item := DailyAttendanceLogItem{
		Date:         day,
		EmployeeUID:  employee.UID,
		EmployeeName: employee.Name,
		CheckIn:      checkIn,
		CheckOut:     checkOut,
		GraceMinutes: shift.GraceMinutes,
	}
	if employee.DepartmentUID != nil {
		v := *employee.DepartmentUID
		item.DepartmentUID = &v
	}

	item.MissingCheckIn = checkIn == nil && checkOut != nil

	workStart, errStart := buildTimeOnDate(day, shift.StartTime)
	workEnd, errEnd := buildTimeOnDate(day, shift.EndTime)
	if errStart != nil || errEnd != nil {
		item.IsAbsent = checkIn == nil && checkOut == nil
		item.Exceptions = buildExceptionTypes(item)
		return item
	}

	item.IsAbsent = (checkIn == nil && checkOut == nil) && isLeaveAwareAbsence(day, time.Now(), workEnd)

	if checkIn != nil && checkOut != nil && checkOut.After(*checkIn) {
		item.WorkedHours = checkOut.Sub(*checkIn).Hours()
	}

	if checkIn != nil {
		threshold := adjuster.LateThreshold(workStart).Add(time.Duration(shift.GraceMinutes) * time.Minute)
		if checkIn.After(threshold) {
			item.LateArrival = true
			delta := int(math.Ceil(checkIn.Sub(threshold).Minutes()))
			item.LateMinutes = &delta
		}
	}

	if checkOut != nil {
		threshold := adjuster.EarlyThreshold(workEnd).Add(-time.Duration(shift.GraceMinutes) * time.Minute)
		if checkOut.Before(threshold) {
			item.EarlyDeparture = true
			delta := int(math.Ceil(threshold.Sub(*checkOut).Minutes()))
			item.EarlyMinutes = &delta
		}
	}

	if checkOut == nil && checkIn != nil {
		suppress, syntheticEarlyAt := adjuster.MissingCheckOutHandling()
		if suppress {
			item.MissingCheckOut = false
			if syntheticEarlyAt != nil {
				item.EarlyDeparture = true
				delta := int(math.Ceil(workEnd.Add(-time.Duration(shift.GraceMinutes) * time.Minute).Sub(*syntheticEarlyAt).Minutes()))
				if delta < 0 {
					delta = 0
				}
				item.EarlyMinutes = &delta
			}
		} else {
			item.MissingCheckOut = shouldFlagMissingPunchOut(day, workEnd)
		}
	}

	item.Exceptions = buildExceptionTypes(item)
	return item
}

// loadCheckInOutForDay reads the earliest check-in and latest check-out punches for the day.
func loadCheckInOutForDay(ctx context.Context, q ports.Querier, employeeUID string, day time.Time) (*time.Time, *time.Time, error) {
	dateStr := day.Format("2006-01-02")
	rows, err := q.QueryContext(ctx, `
		SELECT punched_at, punch_type
		FROM attendance_records
		WHERE employee_uid = $1
		  AND TO_CHAR(punched_at, 'YYYY-MM-DD') = $2`, employeeUID, dateStr)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var checkIn, checkOut *time.Time
	for rows.Next() {
		var punchedAt domain.Time
		var punchType string
		if err := rows.Scan(&punchedAt, &punchType); err != nil {
			return nil, nil, err
		}
		t := punchedAt.Time
		switch domain.AttendancePunchType(punchType) {
		case domain.AttendancePunchTypeCheckIn:
			if checkIn == nil || t.Before(*checkIn) {
				v := t
				checkIn = &v
			}
		case domain.AttendancePunchTypeCheckOut:
			if checkOut == nil || t.After(*checkOut) {
				v := t
				checkOut = &v
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return checkIn, checkOut, nil
}
