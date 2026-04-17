package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type DailyAttendanceLogItem struct {
	Date              time.Time
	EmployeeUID       string
	EmployeeName      string
	DepartmentUID     *string
	CheckIn           *time.Time
	CheckInLogUID     *string
	CheckOut          *time.Time
	CheckOutLogUID    *string
	CheckInDevice     *string
	CheckInDeviceUID  *string
	CheckOutDevice    *string
	CheckOutDeviceUID *string
	WorkedHours       float64
	LateArrival       bool
	EarlyDeparture    bool
	MissingCheckIn    bool
	MissingCheckOut   bool
	IsAbsent          bool
	GraceMinutes      int
	LateMinutes       *int
	EarlyMinutes      *int
	Exceptions        []domain.AttendanceExceptionType
}

type ListDailyDepartmentAttendanceLogsOutput struct {
	DepartmentUID string
	Total         int
	Page          int
	PageSize      int
	TotalPages    int
	Records       []DailyAttendanceLogItem
}

type ListDailyDepartmentAttendanceLogsInput struct {
	DepartmentUID    string
	EmployeeUID      *string
	EmployeeName     *string
	EmployeeNameMode *string
	DeviceUID        *string
	PunchType        *domain.AttendancePunchType
	StartDate        *time.Time
	EndDate          *time.Time
	ListParams       ports.ListParams
}

type ListDailyDepartmentAttendanceLogsUseCase struct {
	db           ports.DB
	deptRepo     ports.DepartmentRepository
	recordRepo   ports.AttendanceRecordRepository
	employeeRepo ports.EmployeeRepository
	shiftRepo    ports.ShiftRepository
	leaveRepo    ports.LeaveRecordRepository
	weekendRepo  ports.WeekendConfigRepository
	holidayRepo  ports.HolidayDefinitionRepository
}

func NewListDailyDepartmentAttendanceLogsUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	recordRepo ports.AttendanceRecordRepository,
	employeeRepo ports.EmployeeRepository,
	shiftRepo ports.ShiftRepository,
	leaveRepo ports.LeaveRecordRepository,
	weekendRepo ports.WeekendConfigRepository,
	holidayRepo ports.HolidayDefinitionRepository,
) *ListDailyDepartmentAttendanceLogsUseCase {
	return &ListDailyDepartmentAttendanceLogsUseCase{
		db:           db,
		deptRepo:     deptRepo,
		recordRepo:   recordRepo,
		employeeRepo: employeeRepo,
		shiftRepo:    shiftRepo,
		leaveRepo:    leaveRepo,
		weekendRepo:  weekendRepo,
		holidayRepo:  holidayRepo,
	}
}

func (uc *ListDailyDepartmentAttendanceLogsUseCase) Execute(ctx context.Context, input ListDailyDepartmentAttendanceLogsInput) (*ListDailyDepartmentAttendanceLogsOutput, error) {
	department, err := uc.deptRepo.GetByUID(ctx, uc.db, input.DepartmentUID)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, ErrDepartmentNotFound
	}

	params := input.ListParams.Normalize(20, 100, "date", ports.SortOrderDesc)
	filter := ports.DepartmentAttendanceLogsFilter{
		EmployeeUID:      input.EmployeeUID,
		EmployeeName:     input.EmployeeName,
		EmployeeNameMode: input.EmployeeNameMode,
		DeviceUID:        input.DeviceUID,
		PunchType:        input.PunchType,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
	}

	rangeStart, rangeEnd, hasRange := expandDateRange(filter.StartDate, filter.EndDate)
	if !hasRange {
		rangeEnd = time.Now()
		rangeStart = normalizeDateOnly(rangeEnd.AddDate(0, 0, -6))
		rangeEnd = normalizeDateOnly(rangeEnd).Add(24*time.Hour - time.Nanosecond)
	}

	filter.StartDate = &rangeStart
	filter.EndDate = &rangeEnd

	allParams := ports.ListParams{
		Page:      1,
		PageSize:  100000,
		SortBy:    "date",
		SortOrder: ports.SortOrderAsc,
	}

	groups, err := uc.recordRepo.ListDailyByDepartmentUID(ctx, uc.db, input.DepartmentUID, filter, allParams)
	if err != nil {
		return nil, err
	}

	records, err := buildDailyAttendanceLogItems(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, groups)
	if err != nil {
		return nil, err
	}

	employees, err := listDepartmentEmployeesForAttendance(ctx, uc.db, input.DepartmentUID, filter)
	if err != nil {
		return nil, err
	}

	records, err = appendDepartmentAbsenceItems(ctx, uc.db, uc.leaveRepo, uc.weekendRepo, uc.holidayRepo, records, employees, input.DepartmentUID, filter, rangeStart, rangeEnd)
	if err != nil {
		return nil, err
	}
	sortDailyAttendanceItems(records, params)

	total := len(records)
	pagedRecords := paginateDailyAttendanceItems(records, params)

	return &ListDailyDepartmentAttendanceLogsOutput{
		DepartmentUID: input.DepartmentUID,
		Total:         total,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    ports.TotalPages(total, params.PageSize),
		Records:       pagedRecords,
	}, nil
}
