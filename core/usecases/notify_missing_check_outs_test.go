package usecases

import (
	"context"
	"testing"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type mockAttendanceRecordRepoForMissingCheckOuts struct {
	groups []*ports.DailyAttendanceGroup
	byDate map[string][]*domain.AttendanceRecord
}

func (m *mockAttendanceRecordRepoForMissingCheckOuts) Create(ctx context.Context, q ports.Querier, record *domain.AttendanceRecord) (bool, error) {
	return false, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.AttendanceRecord, error) {
	return nil, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) Update(ctx context.Context, q ports.Querier, record *domain.AttendanceRecord) error {
	return nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) ListByDate(ctx context.Context, q ports.Querier, date time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	if employeeUID == nil || m.byDate == nil {
		return nil, nil
	}
	return m.byDate[*employeeUID], nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) ListByDateRange(ctx context.Context, q ports.Querier, startDate, endDate time.Time, employeeUID *string) ([]*domain.AttendanceRecord, error) {
	return nil, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) ListByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	return nil, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) CountByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) ListByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.AttendanceRecordWithEmployee, error) {
	return nil, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) CountByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) ListDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	return nil, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) CountDailyByDepartmentUID(ctx context.Context, q ports.Querier, departmentUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) ListDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	return nil, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) CountDailyByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string, filter ports.DepartmentAttendanceLogsFilter) (int, error) {
	return 0, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) ListDaily(ctx context.Context, q ports.Querier, filter ports.DepartmentAttendanceLogsFilter, params ports.ListParams) ([]*ports.DailyAttendanceGroup, error) {
	return m.groups, nil
}
func (m *mockAttendanceRecordRepoForMissingCheckOuts) ResolveEmployeeUIDByDeviceUserID(ctx context.Context, q ports.Querier, deviceUserID string) (*string, error) {
	return nil, nil
}

type mockAttendanceReminderRepo struct {
	existing map[string]*domain.AttendanceReminder
	created  []*domain.AttendanceReminder
}

func (m *mockAttendanceReminderRepo) Create(ctx context.Context, q ports.Querier, reminder *domain.AttendanceReminder) error {
	m.created = append(m.created, reminder)
	return nil
}
func (m *mockAttendanceReminderRepo) GetByEmployeeDateAndType(ctx context.Context, q ports.Querier, employeeUID, attendanceDate string, reminderType domain.AttendanceReminderType) (*domain.AttendanceReminder, error) {
	if m.existing == nil {
		return nil, nil
	}
	return m.existing[employeeUID+"|"+attendanceDate+"|"+string(reminderType)], nil
}

type mockEmployeeRepoForMissingCheckOuts struct {
	employee  *domain.Employee
	employees []*domain.Employee
}

func (m *mockEmployeeRepoForMissingCheckOuts) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.Employee, error) {
	return nil, nil
}
func (m *mockEmployeeRepoForMissingCheckOuts) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Employee, error) {
	if m.employee != nil && m.employee.UID == uid {
		return m.employee, nil
	}
	return nil, nil
}
func (m *mockEmployeeRepoForMissingCheckOuts) Create(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}
func (m *mockEmployeeRepoForMissingCheckOuts) Update(ctx context.Context, q ports.Querier, employee *domain.Employee) error {
	return nil
}
func (m *mockEmployeeRepoForMissingCheckOuts) List(ctx context.Context, q ports.Querier, filter *ports.EmployeeListFilter) ([]*domain.Employee, error) {
	if m.employees != nil {
		return m.employees, nil
	}
	return nil, nil
}
func (m *mockEmployeeRepoForMissingCheckOuts) ExistingGovernmentIDs(ctx context.Context, q ports.Querier, governmentIDs []string) ([]string, error) {
	return nil, nil
}
func (m *mockEmployeeRepoForMissingCheckOuts) ExistingMobiles(ctx context.Context, q ports.Querier, mobiles []string) ([]string, error) {
	return nil, nil
}
func (m *mockEmployeeRepoForMissingCheckOuts) ExistingUniversityIDs(ctx context.Context, q ports.Querier, universityIDs []string) ([]string, error) {
	return nil, nil
}
func (m *mockEmployeeRepoForMissingCheckOuts) Count(ctx context.Context, q ports.Querier) (int, error) {
	return 0, nil
}

type mockShiftRepoForMissingCheckOuts struct {
	shift *domain.Shift
}

func (m *mockShiftRepoForMissingCheckOuts) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.Shift, error) {
	if m.shift != nil && m.shift.UID == uid {
		return m.shift, nil
	}
	return nil, nil
}
func (m *mockShiftRepoForMissingCheckOuts) List(ctx context.Context, q ports.Querier) ([]*domain.Shift, error) {
	return nil, nil
}
func (m *mockShiftRepoForMissingCheckOuts) Upsert(ctx context.Context, q ports.Querier, shift *domain.Shift) error {
	return nil
}

type mockUserRepoForMissingCheckOuts struct {
	user *domain.User
}

func (m *mockUserRepoForMissingCheckOuts) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.User, error) {
	return nil, nil
}

type mockHolidayInstanceRepoForAttendanceReminders struct {
	holidays []*domain.HolidayDefinition
}

func (m *mockHolidayInstanceRepoForAttendanceReminders) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.HolidayDefinition, error) {
	return nil, nil
}
func (m *mockHolidayInstanceRepoForAttendanceReminders) GetByCode(ctx context.Context, q ports.Querier, code string) (*domain.HolidayDefinition, error) {
	return nil, nil
}
func (m *mockHolidayInstanceRepoForAttendanceReminders) GetByDate(ctx context.Context, q ports.Querier, date time.Time) ([]*domain.HolidayDefinition, error) {
	return m.holidays, nil
}
func (m *mockHolidayInstanceRepoForAttendanceReminders) GetByDefinitionAndYear(ctx context.Context, q ports.Querier, definitionID int64, year int) (*domain.HolidayDefinition, error) {
	return nil, nil
}
func (m *mockHolidayInstanceRepoForAttendanceReminders) ListByYear(ctx context.Context, q ports.Querier, year int) ([]*domain.HolidayDefinition, error) {
	return m.holidays, nil
}
func (m *mockHolidayInstanceRepoForAttendanceReminders) ListByDateRange(ctx context.Context, q ports.Querier, start, end time.Time) ([]*domain.HolidayDefinition, error) {
	return m.holidays, nil
}
func (m *mockHolidayInstanceRepoForAttendanceReminders) List(ctx context.Context, q ports.Querier) ([]*domain.HolidayDefinition, error) {
	return m.holidays, nil
}
func (m *mockHolidayInstanceRepoForAttendanceReminders) Create(ctx context.Context, q ports.Querier, instance *domain.HolidayDefinition) error {
	return nil
}
func (m *mockHolidayInstanceRepoForAttendanceReminders) Update(ctx context.Context, q ports.Querier, instance *domain.HolidayDefinition) error {
	return nil
}

type mockLeaveRecordRepoForAttendanceReminders struct {
	onLeave map[int64]bool
}

func (m *mockLeaveRecordRepoForAttendanceReminders) GetByID(ctx context.Context, q ports.Querier, id int64) (*domain.LeaveRecord, error) {
	return nil, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.LeaveRecord, error) {
	return nil, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) Create(ctx context.Context, q ports.Querier, record *domain.LeaveRecord) error {
	return nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) ListByEmployee(ctx context.Context, q ports.Querier, employeeID int64) ([]*domain.LeaveRecord, error) {
	return nil, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) ListByEmployeePaginated(ctx context.Context, q ports.Querier, employeeID int64, limit, offset int) ([]*domain.LeaveRecord, error) {
	return nil, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) CountByEmployee(ctx context.Context, q ports.Querier, employeeID int64) (int, error) {
	return 0, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) ListByEmployeeAndDateRange(ctx context.Context, q ports.Querier, employeeID int64, start, end time.Time) ([]*domain.LeaveRecord, error) {
	return nil, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) ListByEmployeeAndType(ctx context.Context, q ports.Querier, employeeID, leaveTypeID int64) ([]*domain.LeaveRecord, error) {
	return nil, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) ListAllPaginated(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter, limit, offset int) ([]*ports.LeaveRecordWithEmployee, error) {
	return nil, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) CountAll(ctx context.Context, q ports.Querier, filter ports.ListAllLeaveRecordsFilter) (int, error) {
	return 0, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) CountOnLeaveToday(ctx context.Context, q ports.Querier, date time.Time) (int, error) {
	return 0, nil
}
func (m *mockLeaveRecordRepoForAttendanceReminders) HasLeaveOnDate(ctx context.Context, q ports.Querier, employeeID int64, date time.Time) (bool, error) {
	if m.onLeave == nil {
		return false, nil
	}
	return m.onLeave[employeeID], nil
}
func (m *mockUserRepoForMissingCheckOuts) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMissingCheckOuts) GetByPhone(ctx context.Context, q ports.Querier, phone string) (*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMissingCheckOuts) Create(ctx context.Context, q ports.Querier, user *domain.User) error {
	return nil
}
func (m *mockUserRepoForMissingCheckOuts) Update(ctx context.Context, q ports.Querier, user *domain.User) error {
	return nil
}
func (m *mockUserRepoForMissingCheckOuts) List(ctx context.Context, q ports.Querier, limit, offset int) ([]*domain.User, error) {
	return nil, nil
}
func (m *mockUserRepoForMissingCheckOuts) Count(ctx context.Context, q ports.Querier) (int, error) {
	return 0, nil
}
func (m *mockUserRepoForMissingCheckOuts) ExistingPhones(ctx context.Context, q ports.Querier, phones []string) ([]string, error) {
	return nil, nil
}
func (m *mockUserRepoForMissingCheckOuts) GetByEmployeeUID(ctx context.Context, q ports.Querier, employeeUID string) (*domain.User, error) {
	if m.user != nil && m.user.EmployeeUID != nil && *m.user.EmployeeUID == employeeUID {
		return m.user, nil
	}
	return nil, nil
}

type mockNotificationService struct {
	sentCount int
	calls     int
	lastUser  string
	lastData  ports.NotificationData
}

func (m *mockNotificationService) SendToUser(userUID, title, body string, data ports.NotificationData) (int, error) {
	m.calls++
	m.lastUser = userUID
	m.lastData = data
	return m.sentCount, nil
}
func (m *mockNotificationService) SendToUsers(userUIDs []string, title, body string, data ports.NotificationData) (int, error) {
	return 0, nil
}

func TestNotifyMissingCheckOutsUseCaseExecute(t *testing.T) {
	employeeUID := "emp_1"
	shiftUID := "shift_1"
	day := time.Date(2026, 4, 18, 0, 0, 0, 0, time.Local)
	checkIn := time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC)

	recordRepo := &mockAttendanceRecordRepoForMissingCheckOuts{
		groups: []*ports.DailyAttendanceGroup{{
			Date:        day,
			EmployeeUID: employeeUID,
			CheckIn:     &checkIn,
		}},
	}
	reminderRepo := &mockAttendanceReminderRepo{}
	employeeRepo := &mockEmployeeRepoForMissingCheckOuts{
		employee: &domain.Employee{ID: 1, UID: employeeUID, ShiftUID: &shiftUID, Status: domain.EmployeeStatusActive},
	}
	shiftRepo := &mockShiftRepoForMissingCheckOuts{
		shift: &domain.Shift{UID: shiftUID, StartTime: "09:00", EndTime: "17:00", GraceMinutes: 15},
	}
	userRepo := &mockUserRepoForMissingCheckOuts{
		user: &domain.User{UID: "usr_1", EmployeeUID: &employeeUID},
	}
	notificationService := &mockNotificationService{sentCount: 1}

	uc := NewNotifyMissingCheckOutsUseCase(nil, recordRepo, reminderRepo, employeeRepo, nil, shiftRepo, &mockHolidayInstanceRepoForAttendanceReminders{}, &mockLeaveRecordRepoForAttendanceReminders{}, userRepo, notificationService)
	uc.now = func() time.Time {
		return time.Date(2026, 4, 18, 17, 1, 0, 0, time.Local)
	}

	output, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.NotifiedCount != 1 {
		t.Fatalf("expected 1 notification, got %d", output.NotifiedCount)
	}
	if notificationService.calls != 1 {
		t.Fatalf("expected notification service to be called once, got %d", notificationService.calls)
	}
	if len(reminderRepo.created) != 1 {
		t.Fatalf("expected one reminder record, got %d", len(reminderRepo.created))
	}
}

func TestNotifyMissingCheckOutsUseCaseSkipsExistingReminder(t *testing.T) {
	employeeUID := "emp_1"
	shiftUID := "shift_1"
	attendanceDate := "2026-04-18"
	day := time.Date(2026, 4, 18, 0, 0, 0, 0, time.Local)
	checkIn := time.Date(2026, 4, 18, 9, 0, 0, 0, time.UTC)

	recordRepo := &mockAttendanceRecordRepoForMissingCheckOuts{
		groups: []*ports.DailyAttendanceGroup{{
			Date:        day,
			EmployeeUID: employeeUID,
			CheckIn:     &checkIn,
		}},
	}
	reminderRepo := &mockAttendanceReminderRepo{
		existing: map[string]*domain.AttendanceReminder{
			employeeUID + "|" + attendanceDate + "|" + string(domain.AttendanceReminderTypeMissingCheckOut): {
				UID:            "arn_1",
				EmployeeUID:    employeeUID,
				AttendanceDate: attendanceDate,
				ReminderType:   domain.AttendanceReminderTypeMissingCheckOut,
			},
		},
	}
	employeeRepo := &mockEmployeeRepoForMissingCheckOuts{
		employee: &domain.Employee{ID: 1, UID: employeeUID, ShiftUID: &shiftUID, Status: domain.EmployeeStatusActive},
	}
	shiftRepo := &mockShiftRepoForMissingCheckOuts{
		shift: &domain.Shift{UID: shiftUID, StartTime: "09:00", EndTime: "17:00", GraceMinutes: 15},
	}
	userRepo := &mockUserRepoForMissingCheckOuts{
		user: &domain.User{UID: "usr_1", EmployeeUID: &employeeUID},
	}
	notificationService := &mockNotificationService{sentCount: 1}

	uc := NewNotifyMissingCheckOutsUseCase(nil, recordRepo, reminderRepo, employeeRepo, nil, shiftRepo, &mockHolidayInstanceRepoForAttendanceReminders{}, &mockLeaveRecordRepoForAttendanceReminders{}, userRepo, notificationService)
	uc.now = func() time.Time {
		return time.Date(2026, 4, 18, 17, 2, 0, 0, time.Local)
	}

	output, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.NotifiedCount != 0 {
		t.Fatalf("expected 0 notifications, got %d", output.NotifiedCount)
	}
	if notificationService.calls != 0 {
		t.Fatalf("expected notification service not to be called, got %d", notificationService.calls)
	}
}

func TestNotifyMissingCheckOutsUseCaseSendsMissingCheckInReminder(t *testing.T) {
	employeeUID := "emp_1"
	shiftUID := "shift_1"
	day := time.Date(2026, 4, 18, 0, 0, 0, 0, time.Local)

	recordRepo := &mockAttendanceRecordRepoForMissingCheckOuts{}
	reminderRepo := &mockAttendanceReminderRepo{}
	employeeRepo := &mockEmployeeRepoForMissingCheckOuts{
		employees: []*domain.Employee{{
			ID: 1, UID: employeeUID, ShiftUID: &shiftUID, Status: domain.EmployeeStatusActive,
		}},
	}
	shiftRepo := &mockShiftRepoForMissingCheckOuts{
		shift: &domain.Shift{UID: shiftUID, StartTime: "09:00", EndTime: "17:00", GraceMinutes: 15},
	}
	userRepo := &mockUserRepoForMissingCheckOuts{
		user: &domain.User{UID: "usr_1", EmployeeUID: &employeeUID},
	}
	notificationService := &mockNotificationService{sentCount: 1}

	uc := NewNotifyMissingCheckOutsUseCase(nil, recordRepo, reminderRepo, employeeRepo, nil, shiftRepo, &mockHolidayInstanceRepoForAttendanceReminders{}, &mockLeaveRecordRepoForAttendanceReminders{}, userRepo, notificationService)
	uc.now = func() time.Time {
		return time.Date(day.Year(), day.Month(), day.Day(), 9, 16, 0, 0, time.Local)
	}

	output, err := uc.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.NotifiedCount != 1 {
		t.Fatalf("expected 1 notification, got %d", output.NotifiedCount)
	}
	if notificationService.lastData["type"] != "missing_check_in_reminder" {
		t.Fatalf("expected missing_check_in_reminder, got %q", notificationService.lastData["type"])
	}
	if len(reminderRepo.created) != 1 || reminderRepo.created[0].ReminderType != domain.AttendanceReminderTypeMissingCheckIn {
		t.Fatalf("expected missing check-in reminder record to be created")
	}
}

func TestNotifyMissingCheckOutsUseCaseSkipsMissingCheckInOnHolidayOrLeave(t *testing.T) {
	employeeUID := "emp_1"
	shiftUID := "shift_1"
	day := time.Date(2026, 4, 18, 0, 0, 0, 0, time.Local)

	recordRepo := &mockAttendanceRecordRepoForMissingCheckOuts{}
	reminderRepo := &mockAttendanceReminderRepo{}
	employeeRepo := &mockEmployeeRepoForMissingCheckOuts{
		employees: []*domain.Employee{{
			ID: 1, UID: employeeUID, ShiftUID: &shiftUID, Status: domain.EmployeeStatusActive,
		}},
	}
	shiftRepo := &mockShiftRepoForMissingCheckOuts{
		shift: &domain.Shift{UID: shiftUID, StartTime: "09:00", EndTime: "17:00", GraceMinutes: 15},
	}
	userRepo := &mockUserRepoForMissingCheckOuts{
		user: &domain.User{UID: "usr_1", EmployeeUID: &employeeUID},
	}
	notificationService := &mockNotificationService{sentCount: 1}

	holidayUC := NewNotifyMissingCheckOutsUseCase(nil, recordRepo, reminderRepo, employeeRepo, nil, shiftRepo, &mockHolidayInstanceRepoForAttendanceReminders{
		holidays: []*domain.HolidayDefinition{{UID: "hol_1"}},
	}, &mockLeaveRecordRepoForAttendanceReminders{}, userRepo, notificationService)
	holidayUC.now = func() time.Time {
		return time.Date(day.Year(), day.Month(), day.Day(), 9, 16, 0, 0, time.Local)
	}

	output, err := holidayUC.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.NotifiedCount != 0 || notificationService.calls != 0 {
		t.Fatalf("expected holiday run to skip notifications")
	}

	leaveUC := NewNotifyMissingCheckOutsUseCase(nil, recordRepo, reminderRepo, employeeRepo, nil, shiftRepo, &mockHolidayInstanceRepoForAttendanceReminders{}, &mockLeaveRecordRepoForAttendanceReminders{
		onLeave: map[int64]bool{1: true},
	}, userRepo, notificationService)
	leaveUC.now = func() time.Time {
		return time.Date(day.Year(), day.Month(), day.Day(), 9, 16, 0, 0, time.Local)
	}
	notificationService.calls = 0

	output, err = leaveUC.Execute(context.Background())
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if output.NotifiedCount != 0 || notificationService.calls != 0 {
		t.Fatalf("expected leave run to skip notifications")
	}
}
