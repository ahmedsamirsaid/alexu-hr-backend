package domain

import "time"

type ApprovalFlowStep struct {
	ID              int64
	UID             string
	ApprovalFlowUID string
	StepOrder       int
	RoleUID         string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewApprovalFlowStep(approvalFlowUID string, stepOrder int, roleUID string) *ApprovalFlowStep {
	return &ApprovalFlowStep{
		UID:             GenerateUID("afs"),
		ApprovalFlowUID: approvalFlowUID,
		StepOrder:       stepOrder,
		RoleUID:         roleUID,
	}
}
