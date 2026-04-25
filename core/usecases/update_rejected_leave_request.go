package usecases

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type UpdateRejectedLeaveRequestInput struct {
	LeaveRequestUID   string
	ActorUserUID      string
	ActorEmployeeUID  string
	LeaveTypeUID      *string
	SubLeaveTypeUID   *string
	StartDate         *time.Time
	EndDate           *time.Time
	Notes             *string
	StudyDestination  *string
	Assignment        *string
	AssignmentCountry *string
	SpouseWorkCountry *string
	Documents         []SubmitLeaveRequestDocumentInput
}

type UpdateRejectedLeaveRequestOutput struct {
	LeaveRequest    *domain.LeaveRequest
	ApprovalRequest *domain.ApprovalRequest
	Documents       []SubmitLeaveRequestDocumentOutput
}

type UpdateRejectedLeaveRequestUseCase struct {
	db                  ports.DB
	employeeRepo        ports.EmployeeRepository
	leaveTypeRepo       ports.LeaveTypeRepository
	leaveRequestRepo    ports.LeaveRequestRepository
	leaveRequestDocRepo ports.LeaveRequestDocumentRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	approvalActionRepo  ports.ApprovalActionRepository
	workingDaysCalc     *WorkingDaysCalculator
	documentUploadURLUC *GenerateDocumentUploadURLUseCase
}

func NewUpdateRejectedLeaveRequestUseCase(
	db ports.DB,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	leaveRequestRepo ports.LeaveRequestRepository,
	leaveRequestDocRepo ports.LeaveRequestDocumentRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	workingDaysCalc *WorkingDaysCalculator,
	documentUploadURLUC *GenerateDocumentUploadURLUseCase,
) *UpdateRejectedLeaveRequestUseCase {
	return &UpdateRejectedLeaveRequestUseCase{
		db:                  db,
		employeeRepo:        employeeRepo,
		leaveTypeRepo:       leaveTypeRepo,
		leaveRequestRepo:    leaveRequestRepo,
		leaveRequestDocRepo: leaveRequestDocRepo,
		approvalRequestRepo: approvalRequestRepo,
		approvalActionRepo:  approvalActionRepo,
		workingDaysCalc:     workingDaysCalc,
		documentUploadURLUC: documentUploadURLUC,
	}
}

func (uc *UpdateRejectedLeaveRequestUseCase) Execute(ctx context.Context, input UpdateRejectedLeaveRequestInput) (*UpdateRejectedLeaveRequestOutput, error) {
	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	leaveRequest, err := uc.leaveRequestRepo.GetByUID(ctx, tx, input.LeaveRequestUID)
	if err != nil {
		return nil, err
	}
	if leaveRequest == nil {
		return nil, ErrLeaveRequestNotFound
	}
	if leaveRequest.EmployeeUID != input.ActorEmployeeUID {
		return nil, ErrNotRequestOwner
	}

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, tx, leaveRequest.ApprovalRequestUID)
	if err != nil {
		return nil, err
	}
	if approvalRequest == nil {
		return nil, ErrApprovalRequestNotFound
	}
	if approvalRequest.Status != domain.ApprovalRequestStatusRejected {
		return nil, ErrRequestNotRejected
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, tx, leaveRequest.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil || employee.DepartmentUID == nil {
		return nil, ErrNoDepartmentAssigned
	}

	if input.LeaveTypeUID != nil && *input.LeaveTypeUID != "" {
		leaveRequest.LeaveTypeUID = *input.LeaveTypeUID
	}
	if input.SubLeaveTypeUID != nil {
		leaveRequest.SubLeaveTypeUID = input.SubLeaveTypeUID
	}
	if input.StartDate != nil {
		leaveRequest.StartDate = *input.StartDate
	}
	if input.EndDate != nil {
		leaveRequest.EndDate = *input.EndDate
	}
	if input.Notes != nil {
		leaveRequest.Notes = input.Notes
	}
	if input.StudyDestination != nil {
		leaveRequest.StudyDestination = input.StudyDestination
	}
	if input.Assignment != nil {
		leaveRequest.Assignment = input.Assignment
	}
	if input.AssignmentCountry != nil {
		leaveRequest.AssignmentCountry = input.AssignmentCountry
	}
	if input.SpouseWorkCountry != nil {
		leaveRequest.SpouseWorkCountry = input.SpouseWorkCountry
	}

	if leaveRequest.EndDate.Before(leaveRequest.StartDate) {
		return nil, ErrInvalidDateRange
	}

	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, tx, leaveRequest.LeaveTypeUID)
	if err != nil {
		return nil, err
	}
	if leaveType == nil {
		return nil, ErrLeaveTypeNotFound
	}
	if !leaveType.RequiresApproval() {
		return nil, ErrLeaveTypeNoApprovalRequired
	}
	if leaveType.ApprovalFlowUID == nil || *leaveType.ApprovalFlowUID != approvalRequest.ApprovalFlowUID {
		return nil, ErrApprovalFlowMismatch
	}

	if leaveRequest.SubLeaveTypeUID != nil && *leaveRequest.SubLeaveTypeUID != "" {
		subLeaveType, err := uc.leaveTypeRepo.GetSubLeaveTypeByUID(ctx, tx, *leaveRequest.SubLeaveTypeUID)
		if err != nil {
			return nil, err
		}
		if subLeaveType == nil {
			return nil, ErrSubLeaveTypeNotFound
		}
		if subLeaveType.LeaveTypeUID != leaveRequest.LeaveTypeUID {
			return nil, ErrSubLeaveTypeDoesNotBelongToLeaveType
		}
	}

	totalWorkingDays, err := uc.workingDaysCalc.CalculateWorkingDays(ctx, tx, leaveRequest.StartDate, leaveRequest.EndDate)
	if err != nil {
		return nil, err
	}
	if totalWorkingDays == 0 {
		return nil, ErrNoWorkingDays
	}
	leaveRequest.Days = totalWorkingDays

	hasOverlap, err := uc.leaveRequestRepo.HasOverlapping(ctx, tx, leaveRequest.EmployeeUID, leaveRequest.StartDate, leaveRequest.EndDate, &leaveRequest.UID)
	if err != nil {
		return nil, err
	}
	if hasOverlap {
		return nil, ErrOverlappingRequest
	}

	updatedDocs := make([]SubmitLeaveRequestDocumentOutput, 0, len(input.Documents))
	existingDocs, err := uc.leaveRequestDocRepo.ListByLeaveRequestUID(ctx, tx, leaveRequest.UID)
	if err != nil {
		return nil, err
	}

	for _, document := range input.Documents {
		existingDoc := findLeaveRequestDocumentByBaseName(existingDocs, document.FileName)
		if existingDoc == nil {
			return nil, ErrLeaveRequestDocumentNotFound
		}

		presigned, err := uc.documentUploadURLUC.Execute(ctx, GenerateDocumentUploadURLInput{
			UserUID:     input.ActorUserUID,
			FileName:    document.FileName,
			ContentType: document.ContentType,
		})
		if err != nil {
			return nil, err
		}

		if err := uc.leaveRequestDocRepo.UpdateObjectKey(ctx, tx, leaveRequest.UID, existingDoc.FileName, presigned.ObjectKey); err != nil {
			return nil, err
		}

		updatedDocs = append(updatedDocs, SubmitLeaveRequestDocumentOutput{
			FileName:  document.FileName,
			URL:       presigned.URL,
			Method:    presigned.Method,
			Bucket:    presigned.Bucket,
			ObjectKey: presigned.ObjectKey,
		})
	}

	leaveRequest.DecidedAt = nil
	if err := uc.leaveRequestRepo.Update(ctx, tx, leaveRequest); err != nil {
		return nil, err
	}

	approvalRequest.Status = domain.ApprovalRequestStatusPending
	if err := uc.approvalRequestRepo.Update(ctx, tx, approvalRequest); err != nil {
		return nil, err
	}

	actions, err := uc.approvalActionRepo.ListByRequest(ctx, tx, approvalRequest.UID)
	if err != nil {
		return nil, err
	}

	var stepOrder *int
	if len(actions) > 0 {
		stepOrder = actions[len(actions)-1].StepOrder
	}

	resubmitAction := domain.NewApprovalAction(
		approvalRequest.UID,
		domain.ApprovalActionTypeSubmit,
		stepOrder,
		input.ActorEmployeeUID,
		nil,
	)
	if err := uc.approvalActionRepo.Create(ctx, tx, resubmitAction); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &UpdateRejectedLeaveRequestOutput{
		LeaveRequest:    leaveRequest,
		ApprovalRequest: approvalRequest,
		Documents:       updatedDocs,
	}, nil
}

func findLeaveRequestDocumentByBaseName(documents []*domain.LeaveRequestDocument, fileName string) *domain.LeaveRequestDocument {
	target := normalizeDocumentBaseName(fileName)
	for _, document := range documents {
		if normalizeDocumentBaseName(document.FileName) == target {
			return document
		}
	}
	return nil
}

func normalizeDocumentBaseName(fileName string) string {
	trimmed := strings.TrimSpace(fileName)
	ext := filepath.Ext(trimmed)
	if ext == "" {
		return strings.ToLower(trimmed)
	}
	return strings.ToLower(strings.TrimSpace(strings.TrimSuffix(trimmed, ext)))
}
