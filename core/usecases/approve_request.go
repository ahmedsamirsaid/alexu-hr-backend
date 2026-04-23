package usecases

import (
	"context"
	"log/slog"
    "github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ApproveRequestInput struct {
	ApprovalRequestUID string
	ActorUserID        int64
	ActorEmployeeUID   string
	Comments           *string
}

type ApproveRequestOutput struct {
	ApprovalRequest *domain.ApprovalRequest
	LeaveRecord     *domain.LeaveRecord // Set if final approval created a leave record
}

type ApproveRequestUseCase struct {
	db                   ports.DB
	employeeRepo         ports.EmployeeRepository
	leaveTypeRepo        ports.LeaveTypeRepository
	leaveBalanceRepo     ports.LeaveBalanceRepository
	leaveRecordRepo      ports.LeaveRecordRepository
	balanceTxRepo        ports.LeaveBalanceTransactionRepository
	leaveRequestRepo     ports.LeaveRequestRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	roleRepo             ports.RoleRepository
	notificationService  ports.NotificationService
	userRepo             ports.UserRepository
	auditor              audit.Auditor
}

func NewApproveRequestUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	leaveBalanceRepo ports.LeaveBalanceRepository,
	leaveRecordRepo ports.LeaveRecordRepository,
	balanceTxRepo ports.LeaveBalanceTransactionRepository,
	leaveRequestRepo ports.LeaveRequestRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	roleRepo ports.RoleRepository,
	notificationService ports.NotificationService,
	userRepo ports.UserRepository,
	auditor audit.Auditor,
) *ApproveRequestUseCase {
	return &ApproveRequestUseCase{
		db:                   db,
		employeeRepo:         employeeRepo,
		leaveTypeRepo:        leaveTypeRepo,
		leaveBalanceRepo:     leaveBalanceRepo,
		leaveRecordRepo:      leaveRecordRepo,
		balanceTxRepo:        balanceTxRepo,
		leaveRequestRepo:     leaveRequestRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		roleRepo:             roleRepo,
		notificationService:  notificationService,
		userRepo:             userRepo,
		auditor:              auditor,
	}
}

func (uc *ApproveRequestUseCase) Execute(ctx context.Context, input ApproveRequestInput) (*ApproveRequestOutput, error) {
	defer uc.auditor.Actor(input.ActorEmployeeUID).
		Did(audit.ActionApprove).
		On(audit.EntityApprovalRequest, input.ApprovalRequestUID).
		WithMeta("comments", input.Comments).
		Save(ctx)
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, tx, input.ApprovalRequestUID)
	if err != nil {
		return nil, err
	}
	if approvalRequest == nil {
		return nil, ErrApprovalRequestNotFound
	}

	if !approvalRequest.IsPending() {
		return nil, ErrRequestNotPending
	}

	// Get the step configuration for current step
	step, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, tx, approvalRequest.ApprovalFlowUID, approvalRequest.CurrentStep)
	if err != nil {
		return nil, err
	}
	if step == nil {
		return nil, ErrApprovalFlowStepNotFound
	}

	// Get the requester to check their department
	requester, err := uc.employeeRepo.GetByUID(ctx, tx, approvalRequest.RequesterUID)
	if err != nil {
		return nil, err
	}
	if requester == nil || requester.DepartmentUID == nil {
		return nil, ErrNoDepartmentAssigned
	}

	// Check if actor is authorized to approve
	isAuthorized, err := uc.roleRepo.IsUserAuthorizedApprover(ctx, tx, input.ActorUserID, step.RoleUID, *requester.DepartmentUID)
	if err != nil {
		return nil, err
	}
	if !isAuthorized {
		return nil, ErrNotAuthorizedApprover
	}

	// Record the approve action
	currentStep := approvalRequest.CurrentStep
	action := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeApprove,
		&currentStep,
		input.ActorEmployeeUID,
		input.Comments,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, action); err != nil {
		return nil, err
	}

	// Advance approval status
	isFinalApproval := approvalRequest.IsAtFinalStep()
	approvalRequest.Approve()
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	var leaveRecord *domain.LeaveRecord

	// If final approval, create leave record and deduct balance
	if isFinalApproval {
		leaveRequest, err := uc.leaveRequestRepo.GetByApprovalRequestUID(ctx, tx, approvalRequest.UID)
		if err != nil {
			return nil, err
		}
		if leaveRequest == nil {
			return nil, ErrLeaveRequestNotFound
		}

		// Get leave type
		leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, tx, leaveRequest.LeaveTypeUID)
		if err != nil {
			return nil, err
		}
		if leaveType == nil {
			return nil, ErrLeaveTypeNotFound
		}

		// Get or create balance
		year := leaveRequest.StartDate.Year()
		balance, err := uc.leaveBalanceRepo.GetByEmployeeAndTypeAndYear(ctx, tx, requester.ID, leaveType.ID, year)
		if err != nil {
			return nil, err
		}
		if balance == nil {
			balance = domain.NewLeaveBalance(requester.ID, leaveType.ID, year, leaveType.DefaultBalance)
			if err := uc.leaveBalanceRepo.Create(ctx, tx, balance); err != nil {
				return nil, err
			}
		}

		// Re-check balance at approval time
		oldUsedDays := balance.UsedDays
		oldRemainingDays := balance.TotalDays - balance.UsedDays
		remaining := balance.TotalDays - balance.UsedDays
		if leaveRequest.Days > remaining {
			// Reject due to insufficient balance at approval time
			approvalRequest.Status = domain.ApprovalRequestStatusRejected
			if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
				return nil, err
			}
			leaveRequest.SetDecided()
			if err := uc.leaveRequestRepo.Update(ctx, tx, leaveRequest); err != nil {
				return nil, err
			}
			return nil, ErrInsufficientBalance
		}

		// Create leave record
		leaveRecord = domain.NewLeaveRecord(
			requester.ID,
			leaveType.ID,
			leaveRequest.StartDate,
			leaveRequest.EndDate,
			leaveRequest.Days,
			&requester.ID, // recorded_by is the approver's employee ID
			leaveRequest.Notes,
			&leaveRequest.UID,
		)
		if err := uc.leaveRecordRepo.Create(ctx, tx, leaveRecord); err != nil {
			return nil, err
		}

		// Deduct balance
		balance.UsedDays += leaveRequest.Days
		if err := uc.leaveBalanceRepo.Update(ctx, tx, balance); err != nil {
			return nil, err
		}

		// Audit log for balance deduction
		defer uc.auditor.Actor(input.ActorEmployeeUID).
			Did(audit.ActionDeductBalance).
			On(audit.EntityLeaveBalance, balance.UID).
			WithMeta("employee_uid", approvalRequest.RequesterUID).
			WithMeta("leave_type_uid", leaveRequest.LeaveTypeUID).
			WithMeta("leave_type_name", leaveType.NameAR).
			WithMeta("deduction_amount", leaveRequest.Days).
			WithMeta("old_used_days", oldUsedDays).
			WithMeta("new_used_days", balance.UsedDays).
			WithMeta("old_remaining_days", oldRemainingDays).
			WithMeta("new_remaining_days", balance.TotalDays-balance.UsedDays).
			WithMeta("start_date", leaveRequest.StartDate.Format("2006-01-02")).
			WithMeta("end_date", leaveRequest.EndDate.Format("2006-01-02")).
			WithMeta("year", year).
			WithMeta("leave_request_uid", leaveRequest.UID).
			Save(ctx)

		// Record balance transaction
		balanceTx := domain.NewLeaveBalanceTransaction(
			balance.ID,
			domain.TransactionTypeDeduct,
			leaveRequest.Days,
			&leaveRecord.ID,
			&requester.ID,
			leaveRequest.Notes,
		)
		if err := uc.balanceTxRepo.Create(ctx, tx, balanceTx); err != nil {
			return nil, err
		}

		// Set decided_at on leave request
		leaveRequest.SetDecided()
		if err := uc.leaveRequestRepo.Update(ctx, tx, leaveRequest); err != nil {
			return nil, err
		}
	}

	// Get data for notification before commit
	var requesterUID string
	var leaveTypeName string
	var leaveRequestUID string
	var nextStepRoleUID string
	var departmentUID string
	var requesterName string

	lr, _ := uc.leaveRequestRepo.GetByApprovalRequestUID(ctx, tx, approvalRequest.UID)
	if lr != nil {
		leaveRequestUID = lr.UID
		lt, _ := uc.leaveTypeRepo.GetByUID(ctx, tx, lr.LeaveTypeUID)
		if lt != nil {
			leaveTypeName = lt.NameAR
		}
	}

	if isFinalApproval {
		requesterUID = approvalRequest.RequesterUID
	} else {
		// Intermediate approval - get next step approvers
		// Note: approvalRequest.CurrentStep has already been incremented by Approve()
		nextStep, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, tx, approvalRequest.ApprovalFlowUID, approvalRequest.CurrentStep)
		if err == nil && nextStep != nil {
			nextStepRoleUID = nextStep.RoleUID
			departmentUID = *requester.DepartmentUID
			requesterName = requester.Name
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Send notification after commit (non-blocking, uses background context)
	if isFinalApproval && requesterUID != "" {
		go uc.notifyRequesterApproved(requesterUID, leaveTypeName, leaveRequestUID)
	} else if nextStepRoleUID != "" {
		go uc.notifyNextStepApprovers(nextStepRoleUID, departmentUID, requesterName, leaveTypeName, leaveRequestUID)
	}

	return &ApproveRequestOutput{
		ApprovalRequest: approvalRequest,
		LeaveRecord:     leaveRecord,
	}, nil
}

func (uc *ApproveRequestUseCase) notifyRequesterApproved(employeeUID, leaveTypeName, requestUID string) {
	// Get the user linked to this employee
	user, err := uc.userRepo.GetByEmployeeUID(context.Background(), uc.db, employeeUID)
	if err != nil || user == nil {
		slog.Debug("approve_request.notifyRequesterApproved.no_user", "employee_uid", employeeUID)
		return
	}

	title := "تمت الموافقة على طلب الإجازة"
	body := "تمت الموافقة على طلب " + leaveTypeName
	data := ports.NotificationData{
		"type":       "request_approved",
		"requestUid": requestUID,
	}

	_, err = uc.notificationService.SendToUser(user.UID, title, body, data)
	if err != nil {
		slog.Error("approve_request.notifyRequesterApproved.send", "error", err)
	}
}

func (uc *ApproveRequestUseCase) notifyNextStepApprovers(roleUID, departmentUID, requesterName, leaveTypeName, requestUID string) {
	// Get users with the required role for this department
	approvers, err := uc.roleRepo.GetUsersByRoleAndDepartment(context.Background(), uc.db, roleUID, &departmentUID)
	if err != nil {
		slog.Error("approve_request.notifyNextStepApprovers.get_approvers", "error", err, "role_uid", roleUID)
		return
	}

	if len(approvers) == 0 {
		slog.Debug("approve_request.notifyNextStepApprovers.no_approvers", "role_uid", roleUID, "department_uid", departmentUID)
		return
	}

	// Collect user UIDs
	userUIDs := make([]string, len(approvers))
	for i, user := range approvers {
		userUIDs[i] = user.UID
	}

	title := "طلب إجازة بانتظار الموافقة"
	body := "طلب " + leaveTypeName + " من " + requesterName
	data := ports.NotificationData{
		"type":       "pending_approval",
		"requestUid": requestUID,
	}

	_, err = uc.notificationService.SendToUsers(userUIDs, title, body, data)
	if err != nil {
		slog.Error("approve_request.notifyNextStepApprovers.send", "error", err)
	}
}
