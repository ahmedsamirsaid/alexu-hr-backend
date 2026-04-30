package domain

import "time"

type ApprovalRequestStatus string

const (
	ApprovalRequestStatusPending   ApprovalRequestStatus = "pending"
	ApprovalRequestStatusApproved  ApprovalRequestStatus = "approved"
	ApprovalRequestStatusRejected  ApprovalRequestStatus = "rejected"
	ApprovalRequestStatusCancelled ApprovalRequestStatus = "cancelled"
)

type ApprovalRequest struct {
	ID              int64
	UID             string
	ApprovalFlowUID string
	RequesterUID    string
	CurrentStep     int
	MaxStep         int
	Status          ApprovalRequestStatus
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewApprovalRequest(approvalFlowUID, requesterUID string, maxStep int) *ApprovalRequest {
	return NewApprovalRequestWithState(approvalFlowUID, requesterUID, 1, maxStep, ApprovalRequestStatusPending)
}

func NewApprovalRequestWithState(approvalFlowUID, requesterUID string, currentStep, maxStep int, status ApprovalRequestStatus) *ApprovalRequest {
	return &ApprovalRequest{
		UID:             GenerateUID("apr"),
		ApprovalFlowUID: approvalFlowUID,
		RequesterUID:    requesterUID,
		CurrentStep:     currentStep,
		MaxStep:         maxStep,
		Status:          status,
	}
}

func (ar *ApprovalRequest) IsPending() bool {
	return ar.Status == ApprovalRequestStatusPending
}

func (ar *ApprovalRequest) IsAtFinalStep() bool {
	return ar.CurrentStep >= ar.MaxStep
}

func (ar *ApprovalRequest) Approve() {
	if ar.IsAtFinalStep() {
		ar.Status = ApprovalRequestStatusApproved
	} else {
		ar.CurrentStep++
	}
}

func (ar *ApprovalRequest) Reject() {
	ar.Status = ApprovalRequestStatusRejected
}

func (ar *ApprovalRequest) Cancel() {
	ar.Status = ApprovalRequestStatusCancelled
}
