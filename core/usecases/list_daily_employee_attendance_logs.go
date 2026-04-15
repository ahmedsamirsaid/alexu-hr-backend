package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListDailyEmployeeAttendanceLogsOutput struct {
	EmployeeUID string
	Total       int
	Page        int
	PageSize    int
	TotalPages  int
	Records     []DailyAttendanceLogItem
}

type ListDailyEmployeeAttendanceLogsInput struct {
	EmployeeUID      string
	EmployeeName     *string
	EmployeeNameMode *string
	DeviceUID        *string
	PunchType        *domain.AttendancePunchType
	StartDate        *time.Time
	EndDate          *time.Time
	ListParams       ports.ListParams
}

type ListDailyEmployeeAttendanceLogsUseCase struct {
	db            ports.DB
	employeeRepo  ports.EmployeeRepository
	recordRepo    ports.AttendanceRecordRepository
	workHoursRepo ports.WorkHoursConfigRepository
}

func NewListDailyEmployeeAttendanceLogsUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	recordRepo ports.AttendanceRecordRepository,
	workHoursRepo ports.WorkHoursConfigRepository,
) *ListDailyEmployeeAttendanceLogsUseCase {
	return &ListDailyEmployeeAttendanceLogsUseCase{
		db:            db,
		employeeRepo:  employeeRepo,
		recordRepo:    recordRepo,
		workHoursRepo: workHoursRepo,
	}
}

func (uc *ListDailyEmployeeAttendanceLogsUseCase) Execute(ctx context.Context, input ListDailyEmployeeAttendanceLogsInput) (*ListDailyEmployeeAttendanceLogsOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	params := input.ListParams.Normalize(20, 100, "date", ports.SortOrderDesc)
	filter := ports.DepartmentAttendanceLogsFilter{
		EmployeeName:     input.EmployeeName,
		EmployeeNameMode: input.EmployeeNameMode,
		DeviceUID:        input.DeviceUID,
		PunchType:        input.PunchType,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
	}

	total, err := uc.recordRepo.CountDailyByEmployeeUID(ctx, uc.db, input.EmployeeUID, filter)
	if err != nil {
		return nil, err
	}

	groups, err := uc.recordRepo.ListDailyByEmployeeUID(ctx, uc.db, input.EmployeeUID, filter, params)
	if err != nil {
		return nil, err
	}

	records, err := buildDailyAttendanceLogItems(ctx, uc.db, uc.workHoursRepo, groups)
	if err != nil {
		return nil, err
	}

	return &ListDailyEmployeeAttendanceLogsOutput{
		EmployeeUID: input.EmployeeUID,
		Total:       total,
		Page:        params.Page,
		PageSize:    params.PageSize,
		TotalPages:  ports.TotalPages(total, params.PageSize),
		Records:     records,
	}, nil
}
