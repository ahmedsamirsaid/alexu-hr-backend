package usecases

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var (
	ErrLeaveRequestDocumentsRequired    = errors.New("documents array is required when documents_attached is true")
	ErrLeaveRequestDocumentsUnsupported = errors.New("document attachments are only supported for leave requests that require approval")
	ErrLeaveRequestOutsideDeadline      = errors.New("leave request is outside the allowed recording deadline")
)

const autoApprovedLeaveFlowUID = "apf_leave_default"

type SubmitLeaveRequestInput struct {
	UserUID           string
	EmployeeUID       string
	LeaveTypeUID      string
	SubLeaveTypeUID   *string
	OtherSubLeaveName *string
	StartDate         time.Time
	EndDate           time.Time
	Notes             *string
	StudyDestination  *string
	Assignment        *string
	AssignmentCountry *string
	SpouseWorkCountry *string
	DocumentsAttached bool
	Documents         []SubmitLeaveRequestDocumentInput
}

type SubmitLeaveRequestDocumentInput struct {
	FileName    string
	ContentType string
}

type SubmitLeaveRequestDocumentOutput struct {
	FileName  string `json:"fileName"`
	URL       string `json:"url"`
	Method    string `json:"method"`
	Bucket    string `json:"bucket"`
	ObjectKey string `json:"objectKey"`
}

type SubmitLeaveRequestOutput struct {
	LeaveRequest    *domain.LeaveRequest
	ApprovalRequest *domain.ApprovalRequest
	LeaveRecord     *domain.LeaveRecord // Set if auto-approved (no approval flow)
	Documents       []SubmitLeaveRequestDocumentOutput
}

type SubmitLeaveRequestUseCase struct {
	db                   ports.DB
	userRepo             ports.UserRepository
	employeeRepo         ports.EmployeeRepository
	leaveTypeRepo        ports.LeaveTypeRepository
	leaveBalanceRepo     ports.LeaveBalanceRepository
	leaveRequestRepo     ports.LeaveRequestRepository
	leaveRecordRepo      ports.LeaveRecordRepository
	balanceTxRepo        ports.LeaveBalanceTransactionRepository
	approvalRequestRepo  ports.ApprovalRequestRepository
	approvalActionRepo   ports.ApprovalActionRepository
	approvalFlowStepRepo ports.ApprovalFlowStepRepository
	leaveRequestDocRepo  ports.LeaveRequestDocumentRepository
	workingDaysCalc      *WorkingDaysCalculator
	documentUploadURLUC  *GenerateDocumentUploadURLUseCase
	notificationService  ports.NotificationService
	roleRepo             ports.RoleRepository
}

func NewSubmitLeaveRequestUseCase(
	db ports.DB,
	userRepo ports.UserRepository,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	leaveBalanceRepo ports.LeaveBalanceRepository,
	leaveRequestRepo ports.LeaveRequestRepository,
	leaveRecordRepo ports.LeaveRecordRepository,
	balanceTxRepo ports.LeaveBalanceTransactionRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	approvalFlowStepRepo ports.ApprovalFlowStepRepository,
	leaveRequestDocRepo ports.LeaveRequestDocumentRepository,
	workingDaysCalc *WorkingDaysCalculator,
	documentUploadURLUC *GenerateDocumentUploadURLUseCase,
	notificationService ports.NotificationService,
	roleRepo ports.RoleRepository,
) *SubmitLeaveRequestUseCase {
	return &SubmitLeaveRequestUseCase{
		db:                   db,
		userRepo:             userRepo,
		employeeRepo:         employeeRepo,
		leaveTypeRepo:        leaveTypeRepo,
		leaveBalanceRepo:     leaveBalanceRepo,
		leaveRequestRepo:     leaveRequestRepo,
		leaveRecordRepo:      leaveRecordRepo,
		balanceTxRepo:        balanceTxRepo,
		approvalRequestRepo:  approvalRequestRepo,
		approvalActionRepo:   approvalActionRepo,
		approvalFlowStepRepo: approvalFlowStepRepo,
		leaveRequestDocRepo:  leaveRequestDocRepo,
		workingDaysCalc:      workingDaysCalc,
		documentUploadURLUC:  documentUploadURLUC,
		notificationService:  notificationService,
		roleRepo:             roleRepo,
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

	if err := validateLeaveRequestRecordingDeadline(leaveType, input.StartDate, time.Now()); err != nil {
		return nil, err
	}

	if input.SubLeaveTypeUID != nil && *input.SubLeaveTypeUID != "" {
		subLeaveType, err := uc.leaveTypeRepo.GetSubLeaveTypeByUID(ctx, tx, *input.SubLeaveTypeUID)
		if err != nil {
			return nil, err
		}
		if subLeaveType == nil {
			return nil, ErrSubLeaveTypeNotFound
		}
		if subLeaveType.LeaveTypeUID != input.LeaveTypeUID {
			return nil, ErrSubLeaveTypeDoesNotBelongToLeaveType
		}
		if !isOtherSubLeaveType(subLeaveType) {
			input.OtherSubLeaveName = nil
		}
	} else {
		input.OtherSubLeaveName = nil
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

	if leaveType.MaxConsecutive != nil && totalWorkingDays > *leaveType.MaxConsecutive {
		return nil, ErrExceedsConsecutiveDays
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
	balance, _, err := ensureLeaveBalance(ctx, tx, uc.leaveBalanceRepo, employee, leaveType, year, input.StartDate)
	if err != nil {
		return nil, err
	}

	remaining := balance.TotalDays - balance.UsedDays
	if totalWorkingDays > remaining {
		return nil, ErrInsufficientBalance
	}

	if input.DocumentsAttached && len(input.Documents) == 0 {
		return nil, ErrLeaveRequestDocumentsRequired
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
	if input.DocumentsAttached {
		return nil, ErrLeaveRequestDocumentsUnsupported
	}

	approvalRequest := domain.NewApprovalRequestWithState(
		autoApprovedLeaveFlowUID,
		input.EmployeeUID,
		0,
		0,
		domain.ApprovalRequestStatusApproved,
	)
	if err := uc.approvalRequestRepo.Create(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	leaveRequest := domain.NewLeaveRequest(
		input.EmployeeUID,
		input.LeaveTypeUID,
		input.SubLeaveTypeUID,
		input.OtherSubLeaveName,
		input.StartDate,
		input.EndDate,
		totalWorkingDays,
		input.Notes,
		input.StudyDestination,
		input.Assignment,
		input.AssignmentCountry,
		input.SpouseWorkCountry,
		approvalRequest.UID,
	)
	if err := uc.leaveRequestRepo.Create(ctx, tx, leaveRequest); err != nil {
		return nil, err
	}
	leaveRequest.SetDecided()
	if err := uc.leaveRequestRepo.Update(ctx, tx, leaveRequest); err != nil {
		return nil, err
	}

	submitAction := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeSubmit,
		nil,
		input.EmployeeUID,
		nil,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, submitAction); err != nil {
		return nil, err
	}

	documents, err := uc.createLeaveRequestDocuments(ctx, tx, leaveRequest.UID, input)
	if err != nil {
		return nil, err
	}

	leaveRecord := domain.NewLeaveRecord(
		employee.ID,
		leaveType.ID,
		input.StartDate,
		input.EndDate,
		totalWorkingDays,
		&employee.ID, // recorded_by is the employee themselves
		input.Notes,
		&leaveRequest.UID,
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
		LeaveRequest:    leaveRequest,
		ApprovalRequest: approvalRequest,
		LeaveRecord:     leaveRecord,
		Documents:       documents,
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
	requesterUser, err := uc.userRepo.GetByUID(ctx, tx, input.UserUID)
	if err != nil {
		return nil, err
	}
	if requesterUser == nil {
		return nil, ErrUserNotFound
	}

	resolver := newApprovalChainResolver(uc.employeeRepo, uc.userRepo, uc.approvalFlowStepRepo, uc.roleRepo)
	chain, err := resolver.resolveForSubmit(ctx, tx, *leaveType.ApprovalFlowUID, requesterUser.ID, employee)
	if err != nil {
		return nil, err
	}
	if len(chain.effectiveSteps) == 0 && chain.matchedFlowStep == nil {
		return nil, ErrApprovalFlowHasNoSteps
	}
	if len(chain.effectiveSteps) == 0 && chain.matchedFlowStep != nil {
		return uc.handleTopRoleAutoApprove(ctx, tx, employee, leaveType, balance, input, totalWorkingDays)
	}

	maxStep := len(chain.allSteps)
	if maxStep == 0 {
		return nil, ErrApprovalFlowHasNoSteps
	}

	currentStep := 1
	if chain.matchedFlowStep != nil {
		currentStep = chain.matchedFlowStep.StepOrder + 1
	}

	// Create approval request
	approvalRequest := domain.NewApprovalRequestWithState(
		*leaveType.ApprovalFlowUID,
		input.EmployeeUID,
		currentStep,
		maxStep,
		domain.ApprovalRequestStatusPending,
	)
	if err := uc.approvalRequestRepo.Create(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	// Create leave request
	leaveRequest := domain.NewLeaveRequest(
		input.EmployeeUID,
		input.LeaveTypeUID,
		input.SubLeaveTypeUID,
		input.OtherSubLeaveName,
		input.StartDate,
		input.EndDate,
		totalWorkingDays,
		input.Notes,
		input.StudyDestination,
		input.Assignment,
		input.AssignmentCountry,
		input.SpouseWorkCountry,
		approvalRequest.UID,
	)
	if err := uc.leaveRequestRepo.Create(ctx, tx, leaveRequest); err != nil {
		return nil, err
	}

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

	documents, err := uc.createLeaveRequestDocuments(ctx, tx, leaveRequest.UID, input)
	if err != nil {
		return nil, err
	}

	// Notify approvers at the raw current step.
	var step1RoleUID string
	var departmentUID string
	if step1, err := uc.approvalFlowStepRepo.GetByFlowAndStep(ctx, tx, *leaveType.ApprovalFlowUID, approvalRequest.CurrentStep); err == nil && step1 != nil {
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
		Documents:       documents,
	}, nil
}

func (uc *SubmitLeaveRequestUseCase) handleTopRoleAutoApprove(
	ctx context.Context,
	tx ports.Tx,
	employee *domain.Employee,
	leaveType *domain.LeaveType,
	balance *domain.LeaveBalance,
	input SubmitLeaveRequestInput,
	totalWorkingDays int,
) (*SubmitLeaveRequestOutput, error) {
	approvalRequest := domain.NewApprovalRequestWithState(
		*leaveType.ApprovalFlowUID,
		input.EmployeeUID,
		0,
		0,
		domain.ApprovalRequestStatusApproved,
	)
	if err := uc.approvalRequestRepo.Create(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	leaveRequest := domain.NewLeaveRequest(
		input.EmployeeUID,
		input.LeaveTypeUID,
		input.SubLeaveTypeUID,
		input.OtherSubLeaveName,
		input.StartDate,
		input.EndDate,
		totalWorkingDays,
		input.Notes,
		input.StudyDestination,
		input.Assignment,
		input.AssignmentCountry,
		input.SpouseWorkCountry,
		approvalRequest.UID,
	)
	if err := uc.leaveRequestRepo.Create(ctx, tx, leaveRequest); err != nil {
		return nil, err
	}
	leaveRequest.SetDecided()
	if err := uc.leaveRequestRepo.Update(ctx, tx, leaveRequest); err != nil {
		return nil, err
	}

	submitAction := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeSubmit,
		nil,
		input.EmployeeUID,
		nil,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, submitAction); err != nil {
		return nil, err
	}

	documents, err := uc.createLeaveRequestDocuments(ctx, tx, leaveRequest.UID, input)
	if err != nil {
		return nil, err
	}

	leaveRecord := domain.NewLeaveRecord(
		employee.ID,
		leaveType.ID,
		input.StartDate,
		input.EndDate,
		totalWorkingDays,
		&employee.ID,
		input.Notes,
		&leaveRequest.UID,
	)
	if err := uc.leaveRecordRepo.Create(ctx, tx, leaveRecord); err != nil {
		return nil, err
	}

	balance.UsedDays += totalWorkingDays
	if err := uc.leaveBalanceRepo.Update(ctx, tx, balance); err != nil {
		return nil, err
	}

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
		LeaveRequest:    leaveRequest,
		ApprovalRequest: approvalRequest,
		LeaveRecord:     leaveRecord,
		Documents:       documents,
	}, nil
}

func isOtherSubLeaveType(subLeaveType *domain.SubLeaveType) bool {
	if subLeaveType == nil {
		return false
	}

	nameEN := strings.TrimSpace(strings.ToLower(subLeaveType.NameEN))
	nameAR := strings.TrimSpace(subLeaveType.NameAR)

	return nameEN == "other" || nameAR == "أخرى" || nameAR == "أخرى."
}

func validateLeaveRequestRecordingDeadline(leaveType *domain.LeaveType, startDate, now time.Time) error {
	if leaveType == nil || leaveType.RecordingDeadlineDays == nil {
		return nil
	}

	startDay := time.Date(startDate.Year(), startDate.Month(), startDate.Day(), 0, 0, 0, 0, time.UTC)
	nowUTC := now.UTC()
	submitDay := time.Date(nowUTC.Year(), nowUTC.Month(), nowUTC.Day(), 0, 0, 0, 0, time.UTC)
	deadlineDays := *leaveType.RecordingDeadlineDays

	switch leaveType.Code {
	case leaveTypeCodeCasual:
		if submitDay.After(startDay) {
			daysLate := int(submitDay.Sub(startDay).Hours() / 24)
			if daysLate > deadlineDays {
				return ErrLeaveRequestOutsideDeadline
			}
		}
	default:
		if !startDay.After(submitDay) {
			return ErrLeaveRequestOutsideDeadline
		}

		daysAhead := int(startDay.Sub(submitDay).Hours() / 24)
		if daysAhead < deadlineDays {
			return ErrLeaveRequestOutsideDeadline
		}
	}

	return nil
}

func (uc *SubmitLeaveRequestUseCase) createLeaveRequestDocuments(
	ctx context.Context,
	tx ports.Tx,
	leaveRequestUID string,
	input SubmitLeaveRequestInput,
) ([]SubmitLeaveRequestDocumentOutput, error) {
	if !input.DocumentsAttached {
		return nil, nil
	}

	outputs := make([]SubmitLeaveRequestDocumentOutput, 0, len(input.Documents))
	for _, document := range input.Documents {
		presigned, err := uc.documentUploadURLUC.Execute(ctx, GenerateDocumentUploadURLInput{
			UserUID:     input.UserUID,
			FileName:    document.FileName,
			ContentType: document.ContentType,
		})
		if err != nil {
			return nil, err
		}

		doc := domain.NewLeaveRequestDocument(leaveRequestUID, document.FileName, presigned.ObjectKey)
		if err := uc.leaveRequestDocRepo.Create(ctx, tx, doc); err != nil {
			return nil, err
		}

		outputs = append(outputs, SubmitLeaveRequestDocumentOutput{
			FileName:  document.FileName,
			URL:       presigned.URL,
			Method:    presigned.Method,
			Bucket:    presigned.Bucket,
			ObjectKey: presigned.ObjectKey,
		})
	}

	return outputs, nil
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
