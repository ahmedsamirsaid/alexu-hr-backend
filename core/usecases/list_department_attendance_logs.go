package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type DepartmentAttendanceLogItem struct {
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

type ListDepartmentAttendanceLogsOutput struct {
	DepartmentUID string
	Total         int
	Page          int
	PageSize      int
	TotalPages    int
	Logs          []DepartmentAttendanceLogItem
}

type ListDepartmentAttendanceLogsInput struct {
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

type ListDepartmentAttendanceLogsUseCase struct {
	db         ports.DB
	deptRepo   ports.DepartmentRepository
	recordRepo ports.AttendanceRecordRepository
}

func NewListDepartmentAttendanceLogsUseCase(
	db ports.DB,
	deptRepo ports.DepartmentRepository,
	recordRepo ports.AttendanceRecordRepository,
) *ListDepartmentAttendanceLogsUseCase {
	return &ListDepartmentAttendanceLogsUseCase{
		db:         db,
		deptRepo:   deptRepo,
		recordRepo: recordRepo,
	}
}

func (uc *ListDepartmentAttendanceLogsUseCase) Execute(ctx context.Context, input ListDepartmentAttendanceLogsInput) (*ListDepartmentAttendanceLogsOutput, error) {
	department, err := uc.deptRepo.GetByUID(ctx, uc.db, input.DepartmentUID)
	if err != nil {
		return nil, err
	}
	if department == nil {
		return nil, ErrDepartmentNotFound
	}

	params := input.ListParams.Normalize(20, 100, "punchedAt", ports.SortOrderDesc)
	filter := ports.DepartmentAttendanceLogsFilter{
		EmployeeUID:      input.EmployeeUID,
		EmployeeName:     input.EmployeeName,
		EmployeeNameMode: input.EmployeeNameMode,
		DeviceUID:        input.DeviceUID,
		PunchType:        input.PunchType,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
	}

	total, err := uc.recordRepo.CountByDepartmentUID(ctx, uc.db, input.DepartmentUID, filter)
	if err != nil {
		return nil, err
	}

	records, err := uc.recordRepo.ListByDepartmentUID(ctx, uc.db, input.DepartmentUID, filter, params)
	if err != nil {
		return nil, err
	}

	logs := make([]DepartmentAttendanceLogItem, 0, len(records))
	for _, item := range records {
		logs = append(logs, DepartmentAttendanceLogItem{
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

	return &ListDepartmentAttendanceLogsOutput{
		DepartmentUID: input.DepartmentUID,
		Total:         total,
		Page:          params.Page,
		PageSize:      params.PageSize,
		TotalPages:    ports.TotalPages(total, params.PageSize),
		Logs:          logs,
	}, nil
}
