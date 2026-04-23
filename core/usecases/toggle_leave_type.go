package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/adapters/db"
	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ToggleLeaveTypeInput struct {
	UID      string
	IsActive bool
}

type ToggleLeaveTypeOutput struct {
	LeaveType *domain.LeaveType
}

type ToggleLeaveTypeUseCase struct {
	db            ports.DB
	leaveTypeRepo ports.LeaveTypeRepository
	auditor       audit.Auditor
}

func NewToggleLeaveTypeUseCase(
	db ports.DB,
	leaveTypeRepo ports.LeaveTypeRepository,
	auditor audit.Auditor,
) *ToggleLeaveTypeUseCase {
	return &ToggleLeaveTypeUseCase{
		db:            db,
		leaveTypeRepo: leaveTypeRepo,
		auditor:       auditor,
	}
}

func (uc *ToggleLeaveTypeUseCase) Execute(ctx context.Context, input ToggleLeaveTypeInput) (*ToggleLeaveTypeOutput, error) {
	err := uc.leaveTypeRepo.SetActive(ctx, uc.db, input.UID, input.IsActive)
	if err != nil {
		if errors.Is(err, db.ErrLeaveTypeNotFound) {
			return nil, ErrLeaveTypeNotFound
		}
		return nil, err
	}

	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return nil, err
	}

	// Audit log after successful toggle
	action := audit.ActionActivate
	if !input.IsActive {
		action = audit.ActionDeactivate
	}
	defer uc.auditor.From(ctx).Did(action).On(audit.EntityLeaveType, input.UID).
		WithMeta("is_active", input.IsActive).
		Save(ctx)

	return &ToggleLeaveTypeOutput{LeaveType: leaveType}, nil
}
