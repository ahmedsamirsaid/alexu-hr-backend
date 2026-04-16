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
	db           ports.DB
	employeeRepo ports.EmployeeRepository
	deptRepo     ports.DepartmentRepository
	recordRepo   ports.AttendanceRecordRepository
	shiftRepo    ports.ShiftRepository
	leaveRepo    ports.LeaveRecordRepository
}

func NewListDailyEmployeeAttendanceLogsUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	recordRepo ports.AttendanceRecordRepository,
	deptRepo ports.DepartmentRepository,
	shiftRepo ports.ShiftRepository,
	leaveRepo ports.LeaveRecordRepository,
) *ListDailyEmployeeAttendanceLogsUseCase {
	return &ListDailyEmployeeAttendanceLogsUseCase{
		db:           db,
		employeeRepo: employeeRepo,
		deptRepo:     deptRepo,
		recordRepo:   recordRepo,
		shiftRepo:    shiftRepo,
		leaveRepo:    leaveRepo,
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

	groups, err := uc.recordRepo.ListDailyByEmployeeUID(ctx, uc.db, input.EmployeeUID, filter, allParams)
	if err != nil {
		return nil, err
	}

	records, err := buildDailyAttendanceLogItems(ctx, uc.db, uc.employeeRepo, uc.deptRepo, uc.shiftRepo, groups)
	if err != nil {
		return nil, err
	}

	records, err = appendEmployeeAbsenceItems(ctx, uc.db, uc.leaveRepo, records, employee, rangeStart, rangeEnd)
	if err != nil {
		return nil, err
	}
	sortDailyAttendanceItems(records, params)

	total := len(records)
	pagedRecords := paginateDailyAttendanceItems(records, params)

	return &ListDailyEmployeeAttendanceLogsOutput{
		EmployeeUID: input.EmployeeUID,
		Total:       total,
		Page:        params.Page,
		PageSize:    params.PageSize,
		TotalPages:  ports.TotalPages(total, params.PageSize),
		Records:     pagedRecords,
	}, nil
}
