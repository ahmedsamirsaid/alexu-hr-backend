package domain

import "time"

type ApprovalActionType string

const (
	ApprovalActionTypeSubmit  ApprovalActionType = "submit"
	ApprovalActionTypeApprove ApprovalActionType = "approve"
	ApprovalActionTypeReject  ApprovalActionType = "reject"
	ApprovalActionTypeCancel  ApprovalActionType = "cancel"
)

type ApprovalAction struct {
	ID                 int64
	UID                string
	ApprovalRequestUID string
	Action             ApprovalActionType
	StepOrder          *int
	ActorUID           string
	Comments           *string
	ActedAt            time.Time
	CreatedAt          time.Time
}

func NewApprovalAction(approvalRequestUID string, action ApprovalActionType, stepOrder *int, actorUID string, comments *string) *ApprovalAction {
	return &ApprovalAction{
		UID:                GenerateUID("apa"),
		ApprovalRequestUID: approvalRequestUID,
		Action:             action,
		StepOrder:          stepOrder,
		ActorUID:           actorUID,
		Comments:           comments,
		ActedAt:            time.Now(),
	}
}
