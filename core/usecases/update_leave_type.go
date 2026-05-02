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

var (
	ErrInvalidLeaveTypeDefaultBalance        = errors.New("default balance must be greater than or equal to zero")
	ErrInvalidLeaveTypeRecordingDeadlineDays = errors.New("recording deadline days must be greater than or equal to zero")
	ErrInvalidLeaveTypeAdvanceNoticeDays     = errors.New("advance notice days must be greater than or equal to zero")
)

type UpdateLeaveTypeInput struct {
	UID string

	DefaultBalanceSet        bool
	DefaultBalance           *int
	RecordingDeadlineDaysSet bool
	RecordingDeadlineDays    *int
	AdvanceNoticeDaysSet     bool
	AdvanceNoticeDays        *int
}

type UpdateLeaveTypeOutput struct {
	LeaveType *domain.LeaveType
}

type UpdateLeaveTypeUseCase struct {
	db            ports.DB
	leaveTypeRepo ports.LeaveTypeRepository
	auditor       audit.Auditor
}

func NewUpdateLeaveTypeUseCase(
	db ports.DB,
	leaveTypeRepo ports.LeaveTypeRepository,
	auditor audit.Auditor,
) *UpdateLeaveTypeUseCase {
	return &UpdateLeaveTypeUseCase{
		db:            db,
		leaveTypeRepo: leaveTypeRepo,
		auditor:       auditor,
	}
}

func (uc *UpdateLeaveTypeUseCase) Execute(ctx context.Context, input UpdateLeaveTypeInput) (*UpdateLeaveTypeOutput, error) {
	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return nil, err
	}
	if leaveType == nil {
		return nil, ErrLeaveTypeNotFound
	}

	// Store old values for audit log
	oldDefaultBalance := leaveType.DefaultBalance
	oldRecordingDeadlineDays := leaveType.RecordingDeadlineDays
	oldAdvanceNoticeDays := leaveType.AdvanceNoticeDays

	if input.DefaultBalanceSet {
		if input.DefaultBalance == nil || *input.DefaultBalance < 0 {
			return nil, ErrInvalidLeaveTypeDefaultBalance
		}
		leaveType.DefaultBalance = *input.DefaultBalance
	}

	if input.RecordingDeadlineDaysSet {
		if input.RecordingDeadlineDays != nil && *input.RecordingDeadlineDays < 0 {
			return nil, ErrInvalidLeaveTypeRecordingDeadlineDays
		}
		leaveType.RecordingDeadlineDays = input.RecordingDeadlineDays
	}

	if input.AdvanceNoticeDaysSet {
		if input.AdvanceNoticeDays != nil && *input.AdvanceNoticeDays < 0 {
			return nil, ErrInvalidLeaveTypeAdvanceNoticeDays
		}
		leaveType.AdvanceNoticeDays = input.AdvanceNoticeDays
	}

	if err := uc.leaveTypeRepo.Update(ctx, uc.db, leaveType); err != nil {
		if errors.Is(err, db.ErrLeaveTypeNotFound) {
			return nil, ErrLeaveTypeNotFound
		}
		return nil, err
	}

	// Audit log after successful update
	actorName := audit.ActorFromContext(ctx)
	actionSentence := fmt.Sprintf(
		"%s updated leave type '%s'",
		actorName, leaveType.NameEN,
	)
	
	auditBuilder := uc.auditor.From(ctx).Did(audit.ActionUpdate).On(audit.EntityLeaveType, input.UID).
		WithMeta("action", actionSentence).
		WithMeta("leave_type_name", leaveType.NameEN)
	
	if input.DefaultBalanceSet {
		auditBuilder.WithMeta("old_default_balance", oldDefaultBalance).
			WithMeta("new_default_balance", leaveType.DefaultBalance)
	}
	if input.RecordingDeadlineDaysSet {
		auditBuilder.WithMeta("old_recording_deadline_days", oldRecordingDeadlineDays).
			WithMeta("new_recording_deadline_days", leaveType.RecordingDeadlineDays)
	}
	if input.AdvanceNoticeDaysSet {
		auditBuilder.WithMeta("old_advance_notice_days", oldAdvanceNoticeDays).
			WithMeta("new_advance_notice_days", leaveType.AdvanceNoticeDays)
	}
	
	defer auditBuilder.Save(ctx)

	return &UpdateLeaveTypeOutput{LeaveType: leaveType}, nil
}
