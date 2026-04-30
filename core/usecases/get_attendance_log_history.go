package usecases

import (
	"context"

	"github.com/banumusa/backend/core/ports"
)

type GetAttendanceLogHistoryInput struct {
	AttendanceLogUID string
}

type AttendanceLogHistoryItemOutput struct {
	UID           string  `json:"uid"`
	FieldChanged  string  `json:"fieldChanged"`
	OldValue      *string `json:"oldValue,omitempty"`
	NewValue      *string `json:"newValue,omitempty"`
	Reason        *string `json:"reason,omitempty"`
	EditedByUID   string  `json:"editedByUid"`
	EditedByEmployeeUID *string `json:"editedByEmployeeUid,omitempty"`
	EditedByName  string  `json:"editedByName"`
	EditedByPhone *string `json:"editedByPhone,omitempty"`
	EditedAt      string  `json:"editedAt"`
}

type GetAttendanceLogHistoryOutput struct {
	AttendanceLogUID string                           `json:"attendanceLogUid"`
	EmployeeUID      string                           `json:"employeeUid"`
	DepartmentUID    *string                          `json:"departmentUid,omitempty"`
	History          []AttendanceLogHistoryItemOutput `json:"history"`
}

type GetAttendanceLogHistoryUseCase struct {
	db              ports.DB
	recordRepo      ports.AttendanceRecordRepository
	editHistoryRepo ports.AttendanceEditHistoryRepository
	employeeRepo    ports.EmployeeRepository
}

func NewGetAttendanceLogHistoryUseCase(
	db ports.DB,
	recordRepo ports.AttendanceRecordRepository,
	editHistoryRepo ports.AttendanceEditHistoryRepository,
	employeeRepo ports.EmployeeRepository,
) *GetAttendanceLogHistoryUseCase {
	return &GetAttendanceLogHistoryUseCase{
		db:              db,
		recordRepo:      recordRepo,
		editHistoryRepo: editHistoryRepo,
		employeeRepo:    employeeRepo,
	}
}

func (uc *GetAttendanceLogHistoryUseCase) Execute(ctx context.Context, input GetAttendanceLogHistoryInput) (*GetAttendanceLogHistoryOutput, error) {
	record, err := uc.recordRepo.GetByUID(ctx, uc.db, input.AttendanceLogUID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrAttendanceLogNotFound
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, record.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	historyItems, err := uc.editHistoryRepo.ListByAttendanceRecordUID(ctx, uc.db, record.UID)
	if err != nil {
		return nil, err
	}

	output := &GetAttendanceLogHistoryOutput{
		AttendanceLogUID: record.UID,
		EmployeeUID:      record.EmployeeUID,
		DepartmentUID:    employee.DepartmentUID,
		History:          make([]AttendanceLogHistoryItemOutput, 0, len(historyItems)),
	}

	for _, item := range historyItems {
		output.History = append(output.History, AttendanceLogHistoryItemOutput{
			UID:           item.History.UID,
			FieldChanged:  item.History.FieldChanged,
			OldValue:      item.History.OldValue,
			NewValue:      item.History.NewValue,
			Reason:        item.History.Reason,
			EditedByUID:   item.History.EditedByUID,
			EditedByEmployeeUID: item.EditedByEmployeeUID,
			EditedByName:  item.EditedByName,
			EditedByPhone: item.EditedByPhone,
			EditedAt:      item.History.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return output, nil
}
