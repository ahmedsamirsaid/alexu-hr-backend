package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type DailyAttendanceLogItem struct {
	Date            time.Time
	EmployeeUID     string
	EmployeeName    string
	DepartmentUID   *string
	CheckIn         *time.Time
	CheckOut        *time.Time
	CheckInDevice   *string
	CheckOutDevice  *string
	WorkedHours     float64
	LateArrival     bool
	EarlyDeparture  bool
	MissingCheckIn  bool
	MissingCheckOut bool
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
	db            ports.DB
	deptRepo      ports.DepartmentRepository
	recordRepo    ports.AttendanceRecordRepository
	workHoursRepo ports.WorkHoursConfigRepository
}

func NewListDailyDepartmentAttendanceLogsUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	recordRepo ports.AttendanceRecordRepository,
	workHoursRepo ports.WorkHoursConfigRepository,
) *ListDailyDepartmentAttendanceLogsUseCase {
	return &ListDailyDepartmentAttendanceLogsUseCase{
		db:            db,
		deptRepo:      deptRepo,
		recordRepo:    recordRepo,
		workHoursRepo: workHoursRepo,
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

	total, err := uc.recordRepo.CountDailyByDepartmentUID(ctx, uc.db, input.DepartmentUID, filter)
	if err != nil {
		return nil, err
	}

	groups, err := uc.recordRepo.ListDailyByDepartmentUID(ctx, uc.db, input.DepartmentUID, filter, params)
	if err != nil {
		return nil, err
	}

	records, err := buildDailyAttendanceLogItems(ctx, uc.db, uc.workHoursRepo, groups)
	if err != nil {
		return nil, err
	}

	return &ListDailyDepartmentAttendanceLogsOutput{
		DepartmentUID: input.DepartmentUID,
		Total:         total,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    ports.TotalPages(total, params.PageSize),
		Records:       records,
	}, nil
}
