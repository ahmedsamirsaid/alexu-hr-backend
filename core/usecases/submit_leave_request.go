package usecases

import (
	"context"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type SubmitLeaveRequestInput struct {
	EmployeeUID  string
	LeaveTypeUID string
	StartDate    time.Time
	EndDate      time.Time
	Notes        *string
}

type SubmitLeaveRequestOutput struct {
	LeaveRequest    *domain.LeaveRequest
	ApprovalRequest *domain.ApprovalRequest
	LeaveRecord     *domain.LeaveRecord // Set if auto-approved (no approval flow)
}

type SubmitLeaveRequestUseCase struct {
	db                   ports.DB
	employeeRepo         ports.EmployeeRepository
	leaveTypeRepo        ports.LeaveTypeRepository
	leaveBalanceRepo     ports.LeaveBalanceRepository
	leaveRequestRepo     ports.LeaveRequestRepository
	leaveRecordRepo      ports.LeaveRecordRepository
	balanceTxRepo        ports.LeaveBalanceTransactionRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	workingDaysCalc      *WorkingDaysCalculator
	notificationService  ports.NotificationService
	roleRepo             ports.RoleRepository
	auditor              audit.Auditor
}

func NewSubmitLeaveRequestUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	leaveBalanceRepo ports.LeaveBalanceRepository,
	leaveRequestRepo ports.LeaveRequestRepository,
	leaveRecordRepo ports.LeaveRecordRepository,
	balanceTxRepo ports.LeaveBalanceTransactionRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	workingDaysCalc *WorkingDaysCalculator,
	notificationService ports.NotificationService,
	roleRepo ports.RoleRepository,
	auditor audit.Auditor,
) *SubmitLeaveRequestUseCase {
	return &SubmitLeaveRequestUseCase{
		db:                   db,
		employeeRepo:         employeeRepo,
		leaveTypeRepo:        leaveTypeRepo,
		leaveBalanceRepo:     leaveBalanceRepo,
		leaveRequestRepo:     leaveRequestRepo,
		leaveRecordRepo:      leaveRecordRepo,
		balanceTxRepo:        balanceTxRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		workingDaysCalc:      workingDaysCalc,
		notificationService:  notificationService,
		roleRepo:             roleRepo,
		auditor:              auditor,
	}
}

func (uc *SubmitLeaveRequestUseCase) Execute(ctx context.Context, input SubmitLeaveRequestInput) (*SubmitLeaveRequestOutput, error) {
	if input.EndDate.Before(input.StartDate) {
		return nil, ErrInvalidDateRange
	}

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get employee
	employee, err := uc.employeeRepo.GetByUID(ctx, tx, input.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	// Check employee has department
	if employee.DepartmentUID == nil {
		return nil, ErrNoDepartmentAssigned
	}

	// Get leave type
	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, tx, input.LeaveTypeUID)
	if err != nil {
		return nil, err
	}
	if leaveType == nil {
		return nil, ErrLeaveTypeNotFound
	}

	// Calculate working days
	totalWorkingDays, err := uc.workingDaysCalc.CalculateWorkingDays(ctx, tx, input.StartDate, input.EndDate)
	if err != nil {
		return nil, err
	}

	// if totalWorkingDays is 0, return error
	if totalWorkingDays == 0 {
		return nil, ErrNoWorkingDays
	}

	// Check for overlapping requests
	hasOverlap, err := uc.leaveRequestRepo.HasOverlapping(ctx, tx, input.EmployeeUID, input.StartDate, input.EndDate, nil)
	if err != nil {
		return nil, err
	}
	if hasOverlap {
		return nil, ErrOverlappingRequest
	}

	// Get or create balance and check sufficiency
	year := input.StartDate.Year()
	balance, err := uc.leaveBalanceRepo.GetByEmployeeAndTypeAndYear(ctx, tx, employee.ID, leaveType.ID, year)
	if err != nil {
		return nil, err
	}
	if balance == nil {
		// Create balance with default
		balance = domain.NewLeaveBalance(employee.ID, leaveType.ID, year, leaveType.DefaultBalance)
		if err := uc.leaveBalanceRepo.Create(ctx, tx, balance); err != nil {
			return nil, err
		}
	}

	remaining := balance.TotalDays - balance.UsedDays
	if totalWorkingDays > remaining {
		return nil, ErrInsufficientBalance
	}

	// Check if leave type has an approval flow
	// If no approval flow, auto-approve by creating leave record directly
	if !leaveType.RequiresApproval() {
		return uc.handleAutoApprove(ctx, tx, employee, leaveType, balance, input, totalWorkingDays)
	}

	// Normal approval flow
	return uc.handleApprovalFlow(ctx, tx, employee, leaveType, balance, input, totalWorkingDays)
}

func (uc *SubmitLeaveRequestUseCase) handleAutoApprove(
	ctx context.Context,
	tx ports.Tx,
	employee *domain.Employee,
	leaveType *domain.LeaveType,
	balance *domain.LeaveBalance,
	input SubmitLeaveRequestInput,
	totalWorkingDays int,
) (*SubmitLeaveRequestOutput, error) {
	// Create leave record directly (no request/approval needed)
	leaveRecord := domain.NewLeaveRecord(
		employee.ID,
		leaveType.ID,
		input.StartDate,
		input.EndDate,
		totalWorkingDays,
		&employee.ID, // recorded_by is the employee themselves
		input.Notes,
		nil, // no leave_request_uid for auto-approved
	)
	if err := uc.leaveRecordRepo.Create(ctx, tx, leaveRecord); err != nil {
		return nil, err
	}

	// Deduct balance
	balance.UsedDays += totalWorkingDays
	if err := uc.leaveBalanceRepo.Update(ctx, tx, balance); err != nil {
		return nil, err
	}

	// Record balance transaction
	balanceTx := domain.NewLeaveBalanceTransaction(
		balance.ID,
		domain.TransactionTypeDeduct,
		totalWorkingDays,
		&leaveRecord.ID,
		&employee.ID,
		input.Notes,
	)
	if err := uc.balanceTxRepo.Create(ctx, tx, balanceTx); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &SubmitLeaveRequestOutput{
		LeaveRecord: leaveRecord,
	}, nil
}

func (uc *SubmitLeaveRequestUseCase) handleApprovalFlow(
	ctx context.Context,
	tx ports.Tx,
	employee *domain.Employee,
	leaveType *domain.LeaveType,
	balance *domain.LeaveBalance,
	input SubmitLeaveRequestInput,
	totalWorkingDays int,
) (*SubmitLeaveRequestOutput, error) {
	// Count steps in the approval flow
	maxStep, err := uc.approvalFlowStepRepo.CountByFlow(ctx, tx, *leaveType.ApprovalFlowUID)
	if err != nil {
		return nil, err
	}
	if maxStep == 0 {
		return nil, ErrApprovalFlowHasNoSteps
	}

	// Create approval request
	approvalRequest := domain.NewApprovalRequest(*leaveType.ApprovalFlowUID, input.EmployeeUID, maxStep)
	if err := uc.approvalRequestRepo.Create(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	// Create leave request
	leaveRequest := domain.NewLeaveRequest(
		input.EmployeeUID,
		input.LeaveTypeUID,
		input.StartDate,
		input.EndDate,
		totalWorkingDays,
		input.Notes,
		approvalRequest.UID,
	)
	if err := uc.leaveRequestRepo.Create(ctx, tx, leaveRequest); err != nil {
		return nil, err
	}

	// Audit log for leave request submission
	defer uc.auditor.Actor(input.EmployeeUID).
		Did(audit.ActionSubmit).
		On(audit.EntityLeaveRequest, leaveRequest.UID).
		Save(ctx)

	// Record submit action
	submitAction := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeSubmit,
		nil, // step_order is nil for submit
		input.EmployeeUID,
		nil,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, submitAction); err != nil {
		return nil, err
	}

	// Get step 1 approvers for notification
	var step1RoleUID string
	var departmentUID string
	step1, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, tx, *leaveType.ApprovalFlowUID, 1)
	if err == nil && step1 != nil {
		step1RoleUID = step1.RoleUID
		departmentUID = *employee.DepartmentUID
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	// Send notification to step 1 approvers (after commit, non-blocking)
	if step1RoleUID != "" {
		go uc.notifyApprovers(step1RoleUID, departmentUID, employee.Name, leaveType.NameAR, leaveRequest.UID)
	}

	return &SubmitLeaveRequestOutput{
		LeaveRequest:    leaveRequest,
		ApprovalRequest: approvalRequest,
	}, nil
}

func (uc *SubmitLeaveRequestUseCase) notifyApprovers(roleUID, departmentUID, employeeName, leaveTypeName, requestUID string) {
	// Get users with the required role for this department
	approvers, err := uc.roleRepo.GetUsersByRoleAndDepartment(context.Background(), uc.db, roleUID, &departmentUID)
	if err != nil {
		slog.Error("submit_leave_request.notifyApprovers.get_approvers", "error", err)
		return
	}

	if len(approvers) == 0 {
		return
	}

	// Collect user UIDs
	userUIDs := make([]string, len(approvers))
	for i, user := range approvers {
		userUIDs[i] = user.UID
	}

	// Send notification
	title := "طلب إجازة جديد"
	body := "طلب " + employeeName + " " + leaveTypeName
	data := ports.NotificationData{
		"type":       "pending_approval",
		"requestUid": requestUID,
	}

	_, err = uc.notificationService.SendToUsers(userUIDs, title, body, data)
	if err != nil {
		slog.Error("submit_leave_request.notifyApprovers.send", "error", err)
	}
}
