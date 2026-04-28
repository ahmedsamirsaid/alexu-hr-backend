package usecases

import (
	"context"
	"errors"
	"fmt"

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

	// Audit log after successful toggle with human-readable action sentence
	action := audit.ActionActivate
	actionVerb := "activated"
	if !input.IsActive {
		action = audit.ActionDeactivate
		actionVerb = "deactivated"
	}
	
	actorName := audit.ActorFromContext(ctx)
	actionSentence := fmt.Sprintf(
		"%s %s leave type '%s'",
		actorName, actionVerb, leaveType.NameEN,
	)
	
	defer uc.auditor.From(ctx).Did(action).On(audit.EntityLeaveType, input.UID).
		WithMeta("action", actionSentence).
		WithMeta("leave_type_name", leaveType.NameEN).
		WithMeta("old_is_active", !input.IsActive).
		WithMeta("new_is_active", input.IsActive).
		Save(ctx)

	return &ToggleLeaveTypeOutput{LeaveType: leaveType}, nil
}
