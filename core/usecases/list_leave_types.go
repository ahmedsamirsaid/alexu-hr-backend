package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListLeaveTypesOutput struct {
	LeaveTypes []*domain.LeaveType
}

type ListLeaveTypesUseCase struct {
	db            ports.DB
	leaveTypeRepo ports.LeaveTypeRepository
}

func NewListLeaveTypesUseCase(
	db ports.DB,
	leaveTypeRepo ports.LeaveTypeRepository,
) *ListLeaveTypesUseCase {
	return &ListLeaveTypesUseCase{
		db:            db,
		leaveTypeRepo: leaveTypeRepo,
	}
}

func (uc *ListLeaveTypesUseCase) Execute(ctx context.Context, activeOnly bool) (*ListLeaveTypesOutput, error) {
	leaveTypes, err := uc.leaveTypeRepo.List(ctx, uc.db, activeOnly)
	if err != nil {
		return nil, err
	}

	return &ListLeaveTypesOutput{LeaveTypes: leaveTypes}, nil
}
