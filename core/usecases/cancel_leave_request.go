package usecases

import (
	"context"
	"fmt"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type CancelLeaveRequestInput struct {
	LeaveRequestUID  string
	ActorEmployeeUID string // Employee cancelling the request
}

type CancelLeaveRequestUseCase struct {
	db                  ports.DB
	employeeRepo        ports.EmployeeRepository
	leaveRequestRepo    ports.LeaveRequestRepository
	leaveTypeRepo       ports.LeaveTypeRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	approvalActionRepo  ports.ApprovalActionRepository
	auditor             audit.Auditor
}

func NewCancelLeaveRequestUseCase(
	db ports.DB,
	leaveRequestRepo ports.LeaveRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	auditor audit.Auditor,
) *CancelLeaveRequestUseCase {
	return &CancelLeaveRequestUseCase{
		db:                  db,
		leaveRequestRepo:    leaveRequestRepo,
		approvalRequestRepo: approvalRequestRepo,
		approvalActionRepo:  approvalActionRepo,
		auditor:             auditor,
	}
}

// NewCancelLeaveRequestUseCaseWithRepos creates the use case with optional extra repos for richer audit metadata.
func NewCancelLeaveRequestUseCaseWithRepos(
	db ports.DB,
	leaveRequestRepo ports.LeaveRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	auditor audit.Auditor,
) *CancelLeaveRequestUseCase {
	return &CancelLeaveRequestUseCase{
		db:                  db,
		leaveRequestRepo:    leaveRequestRepo,
		approvalRequestRepo: approvalRequestRepo,
		approvalActionRepo:  approvalActionRepo,
		employeeRepo:        employeeRepo,
		leaveTypeRepo:       leaveTypeRepo,
		auditor:             auditor,
	}
}

func (uc *CancelLeaveRequestUseCase) Execute(ctx context.Context, input CancelLeaveRequestInput) error {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	leaveRequest, err := uc.leaveRequestRepo.GetByUID(ctx, tx, input.LeaveRequestUID)
	if err != nil {
		return err
	}
	if leaveRequest == nil {
		return ErrLeaveRequestNotFound
	}

	// Check if requester is the owner
	if leaveRequest.EmployeeUID != input.ActorEmployeeUID {
		return ErrNotRequestOwner
	}

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, tx, leaveRequest.ApprovalRequestUID)
	if err != nil {
		return err
	}
	if approvalRequest == nil {
		return ErrApprovalRequestNotFound
	}

	// Can only cancel pending requests
	if !approvalRequest.IsPending() {
		if approvalRequest.Status == domain.ApprovalRequestStatusApproved {
			return ErrCannotCancelApproved
		}
		return ErrRequestNotPending
	}

	// Resolve names for the human-readable audit action sentence
	actorName := input.ActorEmployeeUID
	if uc.employeeRepo != nil {
		if emp, err := uc.employeeRepo.GetByUID(ctx, tx, input.ActorEmployeeUID); err == nil && emp != nil {
			actorName = emp.Name
		}
	}
	leaveTypeName := "Leave"
	if uc.leaveTypeRepo != nil {
		if lt, err := uc.leaveTypeRepo.GetByUID(ctx, tx, leaveRequest.LeaveTypeUID); err == nil && lt != nil {
			leaveTypeName = lt.NameEN
		}
	}
	startDate := leaveRequest.StartDate.Format("Jan 2, 2006")
	endDate := leaveRequest.EndDate.Format("Jan 2, 2006")

	actionSentence := fmt.Sprintf(
		"%s (Employee) cancelled their %s leave request for %s to %s (%d days)",
		actorName, leaveTypeName, startDate, endDate, leaveRequest.Days,
	)

	// Audit log — deferred, fires on function return
	defer uc.auditor.Actor(input.ActorEmployeeUID).
		Did(audit.ActionCancel).
		On(audit.EntityLeaveRequest, leaveRequest.UID).
		WithMeta("action", actionSentence).
		WithMeta("leave_type", leaveTypeName).
		WithMeta("start_date", startDate).
		WithMeta("end_date", endDate).
		WithMeta("days", leaveRequest.Days).
		WithMeta("old_status", "pending").
		WithMeta("new_status", "cancelled").
		WithMeta("approval_request_uid", approvalRequest.UID).
		Save(ctx)

	// Cancel the approval request
	approvalRequest.Cancel()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return err
	}

	// Set decided_at on leave request
	leaveRequest.SetDecided()
	if err := uc.leaveRequestRepo.Update(ctx, tx, leaveRequest); err != nil {
		return err
	}

	// Record cancel action
	action := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeCancel,
		nil,
		input.ActorEmployeeUID,
		nil,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}
