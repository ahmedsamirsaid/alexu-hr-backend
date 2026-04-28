package usecases

import (
	"context"
	"fmt"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type UpdateApprovalFlowInput struct {
	UID         string
	NameEN      *string
	NameAR      *string
	Description *string
	IsActive    *bool
}

type UpdateApprovalFlowOutput struct {
	Flow *domain.ApprovalFlow
}

type UpdateApprovalFlowUseCase struct {
	db       ports.DB
	flowRepo ports.ApprovalFlowRepository
	auditor  audit.Auditor
}

func NewUpdateApprovalFlowUseCase(
	db ports.DB,
	flowRepo ports.ApprovalFlowRepository,
	auditor audit.Auditor,
) *UpdateApprovalFlowUseCase {
	return &UpdateApprovalFlowUseCase{
		db:       db,
		flowRepo: flowRepo,
		auditor:  auditor,
	}
}

func (uc *UpdateApprovalFlowUseCase) Execute(ctx context.Context, input UpdateApprovalFlowInput) (*UpdateApprovalFlowOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	flow, err := uc.flowRepo.GetByUID(ctx, tx, input.UID)
	if err != nil {
		return nil, err
	}
	if flow == nil {
		return nil, ErrApprovalFlowNotFound
	}

	// Capture old values for audit metadata
	oldNameEN := flow.NameEN
	oldNameAR := ""
	if flow.NameAR != nil {
		oldNameAR = *flow.NameAR
	}
	oldDescription := ""
	if flow.Description != nil {
		oldDescription = *flow.Description
	}
	oldIsActive := flow.IsActive

	// Build audit metadata with field changes
	actorName := audit.ActorFromContext(ctx)
	changedFields := []string{}
	auditBuilder := uc.auditor.From(ctx).Did(audit.ActionUpdate).On(audit.EntityApprovalFlow, input.UID)

	if input.NameEN != nil && *input.NameEN != flow.NameEN {
		auditBuilder.WithMeta("old_name_en", oldNameEN).WithMeta("new_name_en", *input.NameEN)
		changedFields = append(changedFields, fmt.Sprintf("name from '%s' to '%s'", oldNameEN, *input.NameEN))
		flow.NameEN = *input.NameEN
	}
	if input.NameAR != nil {
		newNameAR := ""
		if input.NameAR != nil {
			newNameAR = *input.NameAR
		}
		if newNameAR != oldNameAR {
			auditBuilder.WithMeta("old_name_ar", oldNameAR).WithMeta("new_name_ar", newNameAR)
		}
		flow.NameAR = input.NameAR
	}
	if input.Description != nil {
		newDescription := ""
		if input.Description != nil {
			newDescription = *input.Description
		}
		if newDescription != oldDescription {
			auditBuilder.WithMeta("old_description", oldDescription).WithMeta("new_description", newDescription)
		}
		flow.Description = input.Description
	}
	if input.IsActive != nil && *input.IsActive != flow.IsActive {
		auditBuilder.WithMeta("old_is_active", oldIsActive).WithMeta("new_is_active", *input.IsActive)
		statusChange := "deactivated"
		if *input.IsActive {
			statusChange = "activated"
		}
		changedFields = append(changedFields, statusChange)
		flow.IsActive = *input.IsActive
	}

	// Build human-readable action sentence
	actionSentence := fmt.Sprintf("%s updated approval flow '%s'", actorName, flow.NameEN)
	if len(changedFields) > 0 {
		actionSentence = fmt.Sprintf("%s updated approval flow '%s': %s", actorName, flow.NameEN, changedFields[0])
	}
	auditBuilder.WithMeta("action", actionSentence).WithMeta("flow_name", flow.NameEN)

	if err := uc.flowRepo.Update(ctx, tx, flow); err != nil {
		return nil, err
	}

	// Audit log after successful update
	defer auditBuilder.Save(ctx)

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &UpdateApprovalFlowOutput{Flow: flow}, nil
}
