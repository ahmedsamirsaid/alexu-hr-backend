package usecases

import (
	"context"
	"log/slog"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

const SystemActorUID = "system"

type AutoRejectExpiredRequestsOutput struct {
	RejectedCount int
}

type AutoRejectExpiredRequestsUseCase struct {
	db                  ports.DB
	leaveRequestRepo    ports.LeaveRequestRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	approvalActionRepo  ports.ApprovalActionRepository
	graceDays           int
}

func NewAutoRejectExpiredRequestsUseCase(
	db ports.DB,
	leaveRequestRepo ports.LeaveRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	graceDays int,
) *AutoRejectExpiredRequestsUseCase {
	return &AutoRejectExpiredRequestsUseCase{
		db:                  db,
		leaveRequestRepo:    leaveRequestRepo,
		approvalRequestRepo: approvalRequestRepo,
		approvalActionRepo:  approvalActionRepo,
		graceDays:           graceDays,
	}
}

func (uc *AutoRejectExpiredRequestsUseCase) Execute(ctx context.Context) (*AutoRejectExpiredRequestsOutput, error) {
	expiredRequests, err := uc.leaveRequestRepo.FindExpiredPending(ctx, uc.db, uc.graceDays)
	if err != nil {
		slog.Error("auto_reject_expired_requests.Execute.find_expired", "error", err)
		return nil, err
	}

	if len(expiredRequests) == 0 {
		return &AutoRejectExpiredRequestsOutput{RejectedCount: 0}, nil
	}

	slog.Info("auto_reject_expired_requests.Execute.found_expired", "count", len(expiredRequests))

	rejectedCount := 0
	for _, leaveRequest := range expiredRequests {
		rejected, err := uc.rejectRequest(ctx, leaveRequest)
		if err != nil {
			slog.Error("auto_reject_expired_requests.Execute.reject_request", "error", err, "leave_request_uid", leaveRequest.UID)
			continue
		}
		if rejected {
			rejectedCount++
		}
	}

	slog.Info("auto_reject_expired_requests.Execute.completed", "rejected_count", rejectedCount)
	return &AutoRejectExpiredRequestsOutput{RejectedCount: rejectedCount}, nil
}

func (uc *AutoRejectExpiredRequestsUseCase) rejectRequest(ctx context.Context, leaveRequest *domain.LeaveRequest) (bool, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer tx.Rollback()

	// Get the approval request
	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, tx, leaveRequest.ApprovalRequestUID)
	if err != nil {
		return false, err
	}
	if approvalRequest == nil {
		slog.Warn("auto_reject_expired_requests.rejectRequest.approval_request_not_found", "approval_request_uid", leaveRequest.ApprovalRequestUID)
		return false, nil
	}

	// Double-check it's still pending (race condition protection)
	if !approvalRequest.IsPending() {
		slog.Debug("auto_reject_expired_requests.rejectRequest.not_pending", "approval_request_uid", approvalRequest.UID, "status", approvalRequest.Status)
		return false, nil
	}

	// Record the system rejection action
	comments := "Auto-rejected: leave start date has passed"
	action := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeReject,
		nil, // no step order for system action
		SystemActorUID,
		&comments,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return false, err
	}

	// Update approval request status to rejected
	approvalRequest.Reject()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return false, err
	}

	// Update leave request decided_at
	leaveRequest.SetDecided()
	if err := uc.leaveRequestRepo.Update(ctx, tx, leaveRequest); err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, err
	}

	slog.Info("auto_reject_expired_requests.rejectRequest.success", "leave_request_uid", leaveRequest.UID)
	return true, nil
}
