package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type EmployeeAttendanceLogItem struct {
	UID           string
	EmployeeUID   string
	EmployeeName  string
	DepartmentUID *string
	DeviceUID     string
	DeviceName    string
	DeviceUserID  string
	PunchedAt     time.Time
	PunchType     domain.AttendancePunchType
	RawPayload    *string
}

type ListEmployeeAttendanceLogsOutput struct {
	EmployeeUID   string
	DepartmentUID *string
	Total         int
	Page          int
	PageSize      int
	TotalPages    int
	Logs          []EmployeeAttendanceLogItem
}

type ListEmployeeAttendanceLogsInput struct {
	EmployeeUID      string
	EmployeeName     *string
	EmployeeNameMode *string
	DeviceUID        *string
	PunchType        *domain.AttendancePunchType
	StartDate        *time.Time
	EndDate          *time.Time
	ListParams       ports.ListParams
}

type ListEmployeeAttendanceLogsUseCase struct {
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	recordRepo   ports.AttendanceRecordRepository
}

func NewListEmployeeAttendanceLogsUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	recordRepo ports.AttendanceRecordRepository,
) *ListEmployeeAttendanceLogsUseCase {
	return &ListEmployeeAttendanceLogsUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		recordRepo:   recordRepo,
	}
}

func (uc *ListEmployeeAttendanceLogsUseCase) Execute(ctx context.Context, input ListEmployeeAttendanceLogsInput) (*ListEmployeeAttendanceLogsOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	params := input.ListParams.Normalize(20, 100, "punchedAt", ports.SortOrderDesc)
	filter := ports.DepartmentAttendanceLogsFilter{
		EmployeeName:     input.EmployeeName,
		EmployeeNameMode: input.EmployeeNameMode,
		DeviceUID:        input.DeviceUID,
		PunchType:        input.PunchType,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
	}

	total, err := uc.recordRepo.CountByEmployeeUID(ctx, uc.db, input.EmployeeUID, filter)
	if err != nil {
		return nil, err
	}

	records, err := uc.recordRepo.ListByEmployeeUID(ctx, uc.db, input.EmployeeUID, filter, params)
	if err != nil {
		return nil, err
	}

	logs := make([]EmployeeAttendanceLogItem, 0, len(records))
	for _, item := range records {
		logs = append(logs, EmployeeAttendanceLogItem{
			UID:           item.Record.UID,
			EmployeeUID:   item.Record.EmployeeUID,
			EmployeeName:  item.EmployeeName,
			DepartmentUID: item.DepartmentUID,
			DeviceUID:     item.Record.DeviceUID,
			DeviceName:    item.DeviceName,
			DeviceUserID:  item.Record.DeviceUserID,
			PunchedAt:     item.Record.PunchedAt,
			PunchType:     item.Record.PunchType,
			RawPayload:    item.Record.RawPayload,
		})
	}

	return &ListEmployeeAttendanceLogsOutput{
		EmployeeUID:   input.EmployeeUID,
		DepartmentUID: employee.DepartmentUID,
		Total:         total,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    ports.TotalPages(total, params.PageSize),
		Logs:          logs,
	}, nil
}
