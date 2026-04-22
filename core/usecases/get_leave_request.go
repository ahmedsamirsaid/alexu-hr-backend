package usecases

import (
	"context"
	"strings"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetLeaveRequestOutput struct {
	LeaveRequest    *domain.LeaveRequest
	ApprovalRequest *domain.ApprovalRequest
	Employee        *domain.Employee
	LeaveType       *domain.LeaveType
	SubLeaveType    *domain.SubLeaveType
	Documents       []SubmitLeaveRequestDocumentOutput
}

type GetLeaveRequestUseCase struct {
	db                  ports.DB
	leaveRequestRepo    ports.LeaveRequestRepository
	leaveRequestDocRepo ports.LeaveRequestDocumentRepository
	approvalRequestRepo ports.ApprovalRequestRepository
	employeeRepo        ports.EmployeeRepository
	leaveTypeRepo       ports.LeaveTypeRepository
	documentDownloadUC  *GenerateDocumentDownloadURLUseCase
}

func NewGetLeaveRequestUseCase(
	db ports.DB,
	leaveRequestRepo ports.LeaveRequestRepository,
	leaveRequestDocRepo ports.LeaveRequestDocumentRepository,
	approvalRequestRepo ports.ApprovalRequestRepository,
	employeeRepo ports.EmployeeRepository,
	leaveTypeRepo ports.LeaveTypeRepository,
	documentDownloadUC *GenerateDocumentDownloadURLUseCase,
) *GetLeaveRequestUseCase {
	return &GetLeaveRequestUseCase{
		db:                  db,
		leaveRequestRepo:    leaveRequestRepo,
		leaveRequestDocRepo: leaveRequestDocRepo,
		approvalRequestRepo: approvalRequestRepo,
		employeeRepo:        employeeRepo,
		leaveTypeRepo:       leaveTypeRepo,
		documentDownloadUC:  documentDownloadUC,
	}
}

func (uc *GetLeaveRequestUseCase) Execute(ctx context.Context, uid string) (*GetLeaveRequestOutput, error) {
	var leaveRequest *domain.LeaveRequest
	var err error

	// Try to find by leave request UID first
	leaveRequest, err = uc.leaveRequestRepo.GetByUID(ctx, uc.db, uid)
	if err != nil {
		return nil, err
	}

	// If not found and UID looks like an approval request UID, try that
	if leaveRequest == nil && strings.HasPrefix(uid, "apr_") {
		leaveRequest, err = uc.leaveRequestRepo.GetByApprovalRequestUID(ctx, uc.db, uid)
		if err != nil {
			return nil, err
		}
	}

	if leaveRequest == nil {
		return nil, ErrLeaveRequestNotFound
	}

	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, uc.db, leaveRequest.ApprovalRequestUID)
	if err != nil {
		return nil, err
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, leaveRequest.EmployeeUID)
	if err != nil {
		return nil, err
	}

	leaveType, err := uc.leaveTypeRepo.GetByUID(ctx, uc.db, leaveRequest.LeaveTypeUID)
	if err != nil {
		return nil, err
	}

	var subLeaveType *domain.SubLeaveType
	if leaveRequest.SubLeaveTypeUID != nil && *leaveRequest.SubLeaveTypeUID != "" {
		subLeaveType, err = uc.leaveTypeRepo.GetSubLeaveTypeByUID(ctx, uc.db, *leaveRequest.SubLeaveTypeUID)
		if err != nil {
			return nil, err
		}
	}

	documents, err := uc.leaveRequestDocRepo.ListByLeaveRequestUID(ctx, uc.db, leaveRequest.UID)
	if err != nil {
		return nil, err
	}

	documentOutputs := make([]SubmitLeaveRequestDocumentOutput, 0, len(documents))
	for _, document := range documents {
		presigned, err := uc.documentDownloadUC.Execute(ctx, GenerateDocumentDownloadURLInput{
			ObjectKey:        document.ObjectKey,
			DownloadFileName: document.FileName,
		})
		if err != nil {
			return nil, err
		}

		documentOutputs = append(documentOutputs, SubmitLeaveRequestDocumentOutput{
			FileName:  document.FileName,
			URL:       presigned.URL,
			Method:    presigned.Method,
			Bucket:    presigned.Bucket,
			ObjectKey: presigned.ObjectKey,
		})
	}

	return &GetLeaveRequestOutput{
		LeaveRequest:    leaveRequest,
		ApprovalRequest: approvalRequest,
		Employee:        employee,
		LeaveType:       leaveType,
		SubLeaveType:    subLeaveType,
		Documents:       documentOutputs,
	}, nil
}
