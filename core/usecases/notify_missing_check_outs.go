package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type NotifyMissingCheckOutsOutput struct {
	NotifiedCount int
}

type NotifyMissingCheckOutsUseCase struct {
	db                     ports.DB
	attendanceRecordRepo   ports.AttendanceRecordRepository
	attendanceReminderRepo ports.AttendanceReminderRepository
	employeeRepo           ports.EmployeeRepository
	deptRepo               ports.DepartmentRepository
	shiftRepo              ports.ShiftRepository
	holidayRepo            ports.HolidayDefinitionRepository
	leaveRecordRepo        ports.LeaveRecordRepository
	userRepo               ports.UserRepository
	notificationService    ports.NotificationService
	now                    func() time.Time
}

func NewNotifyMissingCheckOutsUseCase(
	db ports.DB,
	attendanceRecordRepo ports.AttendanceRecordRepository,
	attendanceReminderRepo ports.AttendanceReminderRepository,
	employeeRepo ports.EmployeeRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	holidayRepo ports.HolidayDefinitionRepository,
	leaveRecordRepo ports.LeaveRecordRepository,
	userRepo ports.UserRepository,
	notificationService ports.NotificationService,
) *NotifyMissingCheckOutsUseCase {
	return &NotifyMissingCheckOutsUseCase{
		db:                     db,
		attendanceRecordRepo:   attendanceRecordRepo,
		attendanceReminderRepo: attendanceReminderRepo,
		employeeRepo:           employeeRepo,
		deptRepo:               deptRepo,
		shiftRepo:              shiftRepo,
		holidayRepo:            holidayRepo,
		leaveRecordRepo:        leaveRecordRepo,
		userRepo:               userRepo,
		notificationService:    notificationService,
		now:                    time.Now,
	}
}

func (uc *NotifyMissingCheckOutsUseCase) Execute(ctx context.Context) (*NotifyMissingCheckOutsOutput, error) {
	now := uc.now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)

	params := ports.ListParams{Page: 1, PageSize: 10000, SortBy: "date", SortOrder: ports.SortOrderAsc}
	groups, err := uc.attendanceRecordRepo.ListDaily(ctx, uc.db, ports.DepartmentAttendanceLogsFilter{
		StartDate: &today,
		EndDate:   &today,
	}, params)
	if err != nil {
		slog.Error("notify_missing_check_outs.Execute.list_daily_groups", "error", err)
		return nil, err
	}

	notifiedCount := 0
	missingCheckInCount, err := uc.notifyMissingCheckIns(ctx, today, now)
	if err != nil {
		return nil, err
	}
	notifiedCount += missingCheckInCount

	for _, group := range groups {
		if group == nil || group.CheckIn == nil || group.CheckOut != nil {
			continue
		}

		notified, err := uc.notifyEmployee(ctx, now, group)
		if err != nil {
			slog.Error("notify_missing_check_outs.Execute.notify_employee", "error", err, "employee_uid", group.EmployeeUID)
			continue
		}
		if notified {
			notifiedCount++
		}
	}

	return &NotifyMissingCheckOutsOutput{NotifiedCount: notifiedCount}, nil
}

func (uc *NotifyMissingCheckOutsUseCase) notifyMissingCheckIns(ctx context.Context, today, now time.Time) (int, error) {
	if uc.holidayRepo != nil {
		holidays, err := uc.holidayRepo.ListByDateRange(ctx, uc.db, today, today)
		if err != nil {
			slog.Error("notify_missing_check_outs.notifyMissingCheckIns.list_holidays", "error", err)
			return 0, err
		}
		if len(holidays) > 0 {
			return 0, nil
		}
	}

	activeStatus := domain.EmployeeStatusActive
	employees, err := uc.employeeRepo.List(ctx, uc.db, &ports.EmployeeListFilter{Status: &activeStatus})
	if err != nil {
		slog.Error("notify_missing_check_outs.notifyMissingCheckIns.list_employees", "error", err)
		return 0, err
	}

	attendanceDate := today.Format("2006-01-02")
	notifiedCount := 0
	for _, employee := range employees {
		if employee == nil {
			continue
		}
		notified, err := uc.notifyEmployeeMissingCheckIn(ctx, now, today, attendanceDate, employee)
		if err != nil {
			slog.Error("notify_missing_check_outs.notifyMissingCheckIns.notify_employee", "error", err, "employee_uid", employee.UID)
			continue
		}
		if notified {
			notifiedCount++
		}
	}

	return notifiedCount, nil
}

func (uc *NotifyMissingCheckOutsUseCase) notifyEmployeeMissingCheckIn(ctx context.Context, now, today time.Time, attendanceDate string, employee *domain.Employee) (bool, error) {
	existing, err := uc.attendanceReminderRepo.GetByEmployeeDateAndType(ctx, uc.db, employee.UID, attendanceDate, domain.AttendanceReminderTypeMissingCheckIn)
	if err != nil {
		return false, err
	}
	if existing != nil {
		return false, nil
	}

	onLeave, err := hasLeaveOnAttendanceDay(ctx, uc.db, uc.leaveRecordRepo, employee.ID, today)
	if err != nil {
		return false, err
	}
	if onLeave {
		return false, nil
	}

	records, err := uc.attendanceRecordRepo.ListByDate(ctx, uc.db, today, &employee.UID)
	if err != nil {
		return false, err
	}
	for _, record := range records {
		if record == nil {
			continue
		}
		if record.PunchType == domain.AttendancePunchTypeCheckIn || record.PunchType == domain.AttendancePunchTypeUnknown {
			return false, nil
		}
	}

	shift, err := resolveEffectiveShift(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, employee.UID, employee.DepartmentUID)
	if err != nil {
		return false, err
	}

	workStart, err := buildTimeOnDate(today, shift.StartTime)
	if err != nil {
		return false, ErrInvalidWorkDayStart
	}

	threshold := workStart.Add(time.Duration(shift.GraceMinutes) * time.Minute).Add(time.Minute)
	if now.Before(threshold) {
		return false, nil
	}

	user, err := uc.userRepo.GetByEmployeeUID(ctx, uc.db, employee.UID)
	if err != nil || user == nil {
		return false, err
	}

	title := "notification.attendance_reminder.title"
	body := "notification.attendance_reminder.body"
	data := ports.NotificationData{
		"type":           "missing_check_in_reminder",
		"employeeUid":    employee.UID,
		"attendanceDate": attendanceDate,
	}

	sentCount, err := uc.notificationService.SendToUser(user.UID, title, body, nil, data)
	if err != nil {
		return false, err
	}
	if sentCount == 0 {
		return false, nil
	}

	reminder := domain.NewAttendanceReminder(employee.UID, attendanceDate, domain.AttendanceReminderTypeMissingCheckIn, now)
	if err := uc.attendanceReminderRepo.Create(ctx, uc.db, reminder); err != nil {
		return false, err
	}

	return true, nil
}

func (uc *NotifyMissingCheckOutsUseCase) notifyEmployee(ctx context.Context, now time.Time, group *ports.DailyAttendanceGroup) (bool, error) {
	shift, err := resolveEffectiveShift(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, group.EmployeeUID, group.DepartmentUID)
	if err != nil {
		return false, err
	}

	workEnd, err := buildTimeOnDate(group.Date, shift.EndTime)
	if err != nil {
		return false, ErrInvalidWorkDayEnd
	}

	threshold := workEnd.Add(time.Minute)
	if now.Before(threshold) {
		return false, nil
	}

	attendanceDate := group.Date.Format("2006-01-02")
	existing, err := uc.attendanceReminderRepo.GetByEmployeeDateAndType(ctx, uc.db, group.EmployeeUID, attendanceDate, domain.AttendanceReminderTypeMissingCheckOut)
	if err != nil {
		return false, err
	}
	if existing != nil {
		return false, nil
	}

	user, err := uc.userRepo.GetByEmployeeUID(ctx, uc.db, group.EmployeeUID)
	if err != nil || user == nil {
		return false, err
	}

	title := "notification.attendance_reminder.title"
	body := "notification.attendance_reminder.body"
	data := ports.NotificationData{
		"type":           "missing_check_out_reminder",
		"employeeUid":    group.EmployeeUID,
		"attendanceDate": attendanceDate,
	}

	sentCount, err := uc.notificationService.SendToUser(user.UID, title, body, nil, data)
	if err != nil {
		return false, err
	}
	if sentCount == 0 {
		return false, nil
	}

	reminder := domain.NewAttendanceReminder(group.EmployeeUID, attendanceDate, domain.AttendanceReminderTypeMissingCheckOut, now)
	if err := uc.attendanceReminderRepo.Create(ctx, uc.db, reminder); err != nil {
		return false, err
	}

	return true, nil
}
