package usecases

import (
	"context"
	"math"
	"time"

	"github.com/banumusa/backend/core/ports"
)

func buildDailyAttendanceLogItems(ctx context.Context, db ports.DB, employeeRepo ports.EmployeeRepository, deptRepo ports.DepartmentRepository, shiftRepo ports.ShiftRepository, groups []*ports.DailyAttendanceGroup) ([]DailyAttendanceLogItem, error) {
	return buildDailyAttendanceLogItemsWithPermissions(ctx, db, employeeRepo, deptRepo, shiftRepo, nil, groups)
}

// buildDailyAttendanceLogItemsWithPermissions is the permission-aware variant. Pass permRepo == nil
// to preserve legacy behavior (no excuse adjustments).
func buildDailyAttendanceLogItemsWithPermissions(ctx context.Context, db ports.DB, employeeRepo ports.EmployeeRepository, deptRepo ports.DepartmentRepository, shiftRepo ports.ShiftRepository, permRepo ports.PermissionRequestRepository, groups []*ports.DailyAttendanceGroup) ([]DailyAttendanceLogItem, error) {
	records := make([]DailyAttendanceLogItem, 0, len(groups))
	for _, group := range groups {
		shift, err := resolveEffectiveShift(ctx, db, employeeRepo, deptRepo, shiftRepo, group.EmployeeUID, group.DepartmentUID)
		if err != nil {
			return nil, err
		}

		adjuster, err := newPermissionAdjusterForDate(ctx, db, permRepo, group.EmployeeUID, group.Date, shift)
		if err != nil {
			return nil, err
		}

		item := DailyAttendanceLogItem{
			Date:              group.Date,
			EmployeeUID:       group.EmployeeUID,
			EmployeeName:      group.EmployeeName,
			DepartmentUID:     group.DepartmentUID,
			CheckIn:           group.CheckIn,
			CheckInLogUID:     group.CheckInLogUID,
			CheckOut:          group.CheckOut,
			CheckOutLogUID:    group.CheckOutLogUID,
			CheckInDevice:     group.CheckInDevice,
			CheckInDeviceUID:  group.CheckInDeviceUID,
			CheckOutDevice:    group.CheckOutDevice,
			CheckOutDeviceUID: group.CheckOutDeviceUID,
			HasEditHistory:    group.HasEditHistory,
			GraceMinutes:      shift.GraceMinutes,
			PermissionUIDs:    adjuster.PermissionUIDs(),
		}

		item.MissingCheckIn = group.CheckIn == nil && group.CheckOut != nil

		workStart, err := buildTimeOnDate(group.Date, shift.StartTime)
		if err != nil {
			return nil, ErrInvalidWorkDayStart
		}
		workEnd, err := buildTimeOnDate(group.Date, shift.EndTime)
		if err != nil {
			return nil, ErrInvalidWorkDayEnd
		}

		item.IsAbsent = (group.CheckIn == nil && group.CheckOut == nil) && isLeaveAwareAbsence(group.Date, time.Now(), workEnd)

		if group.CheckIn != nil && group.CheckOut != nil && group.CheckOut.After(*group.CheckIn) {
			item.WorkedHours = group.CheckOut.Sub(*group.CheckIn).Hours()
		}
		if group.CheckIn != nil {
			lateThreshold := adjuster.LateThreshold(workStart).Add(time.Duration(shift.GraceMinutes) * time.Minute)
			if group.CheckIn.After(lateThreshold) {
				item.LateArrival = true
				delta := durationMinutesCeil(group.CheckIn.Sub(lateThreshold))
				item.LateMinutes = &delta
			}
		}
		if group.CheckOut != nil {
			earlyThreshold := adjuster.EarlyThreshold(workEnd).Add(-time.Duration(shift.GraceMinutes) * time.Minute)
			if group.CheckOut.Before(earlyThreshold) {
				item.EarlyDeparture = true
				delta := durationMinutesCeil(earlyThreshold.Sub(*group.CheckOut))
				item.EarlyMinutes = &delta
			}
		}

		// Missing checkout / synthetic early
		if group.CheckOut == nil && group.CheckIn != nil {
			suppress, syntheticEarlyAt := adjuster.MissingCheckOutHandling()
			if suppress {
				item.MissingCheckOut = false
				if syntheticEarlyAt != nil {
					item.EarlyDeparture = true
					defaultEarly := workEnd.Add(-time.Duration(shift.GraceMinutes) * time.Minute)
					if syntheticEarlyAt.Before(defaultEarly) {
						delta := durationMinutesCeil(defaultEarly.Sub(*syntheticEarlyAt))
						item.EarlyMinutes = &delta
					}
				}
			} else {
				item.MissingCheckOut = shouldFlagMissingPunchOut(group.Date, workEnd)
			}
		}

		item.Exceptions = buildExceptionTypes(item)
		records = append(records, item)
	}

	return records, nil
}

func durationMinutesCeil(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	return int(math.Ceil(d.Minutes()))
}

func shouldFlagMissingPunchOut(day time.Time, workEnd time.Time) bool {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	dayOnly := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)

	if dayOnly.Before(today) {
		return true
	}
	if dayOnly.After(today) {
		return false
	}
	return !now.Before(workEnd)
}

func hasLeaveOnAttendanceDay(ctx context.Context, db ports.DB, leaveRepo ports.LeaveRecordRepository, employeeID int64, day time.Time) (bool, error) {
	if leaveRepo == nil {
		return false, nil
	}

	return leaveRepo.HasLeaveOnDate(ctx, db, employeeID, day)
}

func isLeaveAwareAbsence(day time.Time, now time.Time, workEnd time.Time) bool {
	dayOnly := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, time.Local)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	if dayOnly.Before(today) {
		return true
	}
	if dayOnly.After(today) {
		return false
	}
	return !now.Before(workEnd)
}
