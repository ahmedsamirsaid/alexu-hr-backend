package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListAllLeaveRecordsInput struct {
	Search       string
	LeaveTypeUID *string
	StartDate    *time.Time
	EndDate      *time.Time
	Page         int
	PageSize     int
}

type AllLeaveRecordInfo struct {
	UID             string
	EmployeeUID     string
	EmployeeName    string
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

type ListAllLeaveRecordsOutput struct {
	Records  []AllLeaveRecordInfo
	Total    int
	Page     int
	PageSize int
}

type ListAllLeaveRecordsUseCase struct {
	db              ports.DB
	leaveTypeRepo   ports.LeaveTypeRepository
	leaveRecordRepo ports.LeaveRecordRepository
}

func NewListAllLeaveRecordsUseCase(
	db ports.DB,
	leaveTypeRepo ports.LeaveTypeRepository,
	leaveRecordRepo ports.LeaveRecordRepository,
) *ListAllLeaveRecordsUseCase {
	return &ListAllLeaveRecordsUseCase{
		db:              db,
		leaveTypeRepo:   leaveTypeRepo,
		leaveRecordRepo: leaveRecordRepo,
	}
}

func (uc *ListAllLeaveRecordsUseCase) Execute(ctx context.Context, input ListAllLeaveRecordsInput) (*ListAllLeaveRecordsOutput, error) {
	// Load leave types for mapping (include inactive for historical records)
	leaveTypes, err := uc.leaveTypeRepo.List(ctx, uc.db, false)
	if err != nil {
		return nil, err
	}

	leaveTypeMap := make(map[int64]*domain.LeaveType)
	leaveTypeUIDMap := make(map[string]*domain.LeaveType)
	for _, lt := range leaveTypes {
		leaveTypeMap[lt.ID] = lt
		leaveTypeUIDMap[lt.UID] = lt
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

	// Build filter
	filter := ports.ListAllLeaveRecordsFilter{
		Search:    input.Search,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
	}

	// Resolve leave type UID to ID if provided
	if input.LeaveTypeUID != nil {
		lt := leaveTypeUIDMap[*input.LeaveTypeUID]
		if lt == nil {
			return nil, ErrLeaveTypeNotFound
		}
		filter.LeaveTypeID = &lt.ID
	}

	// Get total count
	total, err := uc.leaveRecordRepo.CountAll(ctx, uc.db, filter)
	if err != nil {
		return nil, err
	}

	// Get paginated records
	records, err := uc.leaveRecordRepo.ListAllPaginated(ctx, uc.db, filter, pageSize, offset)
	if err != nil {
		return nil, err
	}

	// Map to output
	var result []AllLeaveRecordInfo
	for _, rec := range records {
		lt := leaveTypeMap[rec.LeaveRecord.LeaveTypeID]
		if lt == nil {
			continue
		}

		result = append(result, AllLeaveRecordInfo{
			UID:             rec.LeaveRecord.UID,
			EmployeeUID:     rec.EmployeeUID,
			EmployeeName:    rec.EmployeeName,
			LeaveTypeUID:    lt.UID,
			LeaveTypeCode:   lt.Code,
			LeaveTypeNameEN: lt.NameEN,
			LeaveTypeNameAR: lt.NameAR,
			StartDate:       rec.LeaveRecord.StartDate,
			EndDate:         rec.LeaveRecord.EndDate,
			Days:            rec.LeaveRecord.Days,
			RecordedAt:      rec.LeaveRecord.RecordedAt,
			Notes:           rec.LeaveRecord.Notes,
		})
	}

	return &ListAllLeaveRecordsOutput{
		Records:  result,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}
