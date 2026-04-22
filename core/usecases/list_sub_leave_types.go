package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ListSubLeaveTypesOutput struct {
	LeaveType     *domain.LeaveType
	SubLeaveTypes []*domain.SubLeaveType
}

type ListSubLeaveTypesUseCase struct {
	db            ports.DB
	leaveTypeRepo ports.LeaveTypeRepository
}

func NewListSubLeaveTypesUseCase(
	db ports.DB,
	leaveTypeRepo ports.LeaveTypeRepository,
) *ListSubLeaveTypesUseCase {
	return &ListSubLeaveTypesUseCase{
		db:            db,
		leaveTypeRepo: leaveTypeRepo,
	}
}

func (uc *ListSubLeaveTypesUseCase) Execute(ctx context.Context, leaveTypeUID string) (*ListSubLeaveTypesOutput, error) {
	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, leaveTypeUID)
	if err != nil {
		return nil, err
	}
	if leaveType == nil {
		return nil, ErrLeaveTypeNotFound
	}

	subLeaveTypes, err := uc.leaveTypeRepo.ListSubLeaveTypesByLeaveTypeUID(ctx, uc.db, leaveTypeUID)
	if err != nil {
		return nil, err
	}

	return &ListSubLeaveTypesOutput{
		LeaveType:     leaveType,
		SubLeaveTypes: subLeaveTypes,
	}, nil
}
