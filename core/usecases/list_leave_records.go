package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListLeaveRecordsInput struct {
	EmployeeUID  string
	LeaveTypeUID *string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	PageSize     int
}

type LeaveRecordInfo struct {
	UID             string
	LeaveTypeUID    string
	LeaveTypeCode   string
	LeaveTypeNameEN string
	LeaveTypeNameAR string
	StartDate       time.Time
	EndDate         time.Time
	Days            int
	RecordedAt      time.Time
	Notes           *string
}

type ListLeaveRecordsOutput struct {
	Records  []LeaveRecordInfo
	Total    int
	Page     int
	PageSize int
}

type ListLeaveRecordsUseCase struct {
	db              ports.DB
	employeeRepo    ports.EmployeeRepository
	leaveTypeRepo   ports.LeaveTypeRepository
	leaveRecordRepo ports.LeaveRecordRepository
}

func NewListLeaveRecordsUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	leaveRecordRepo ports.LeaveRecordRepository,
) *ListLeaveRecordsUseCase {
	return &ListLeaveRecordsUseCase{
		db:              db,
		employeeRepo:    employeeRepo,
		leaveTypeRepo:   leaveTypeRepo,
		leaveRecordRepo: leaveRecordRepo,
	}
}

func (uc *ListLeaveRecordsUseCase) Execute(ctx context.Context, input ListLeaveRecordsInput) (*ListLeaveRecordsOutput, error) {
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	leaveTypes, err := uc.leaveTypeRepo.List(ctx, uc.db, false) // activeOnly=false to show historical records
	if err != nil {
		return nil, err
	}

	leaveTypeMap := make(map[int64]*domain.LeaveType)
	for _, lt := range leaveTypes {
		leaveTypeMap[lt.ID] = lt
	}

	// Default pagination values
	page := input.Page
	if page < 1 {
		page = 1
	}
	pageSize := input.PageSize
	if pageSize < 1 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	var records []*domain.LeaveRecord
	var total int

	if input.LeaveTypeUID != nil {
		leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, *input.LeaveTypeUID)
		if err != nil {
			return nil, err
		}
		if leaveType == nil {
			return nil, ErrLeaveTypeNotFound
		}
		records, err = uc.leaveRecordRepo.ListByEmployeeAndType(ctx, uc.db, employee.ID, leaveType.ID)
		if err != nil {
			return nil, err
		}
		total = len(records)
	} else if input.StartDate != nil && input.EndDate != nil {
		records, err = uc.leaveRecordRepo.ListByEmployeeAndDateRange(ctx, uc.db, employee.ID, *input.StartDate, *input.EndDate)
		if err != nil {
			return nil, err
		}
		total = len(records)
	} else {
		// Use paginated query for the default case
		total, err = uc.leaveRecordRepo.CountByEmployee(ctx, uc.db, employee.ID)
		if err != nil {
			return nil, err
		}
		records, err = uc.leaveRecordRepo.ListByEmployeePaginated(ctx, uc.db, employee.ID, pageSize, offset)
		if err != nil {
			return nil, err
		}
	}

	var result []LeaveRecordInfo
	for _, rec := range records {
		lt := leaveTypeMap[rec.LeaveTypeID]
		if lt == nil {
			continue
		}

		result = append(result, LeaveRecordInfo{
			UID:             rec.UID,
			LeaveTypeUID:    lt.UID,
			LeaveTypeCode:   lt.Code,
			LeaveTypeNameEN: lt.NameEN,
			LeaveTypeNameAR: lt.NameAR,
			StartDate:       rec.StartDate,
			EndDate:         rec.EndDate,
			Days:            rec.Days,
			RecordedAt:      rec.RecordedAt,
			Notes:           rec.Notes,
		})
	}

	return &ListLeaveRecordsOutput{
		Records:  result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
