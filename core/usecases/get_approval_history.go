package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ApprovalHistoryItem struct {
	Action       *domain.ApprovalAction
	ActorName    string
	ActorUID     string
}

type GetApprovalHistoryOutput struct {
	ApprovalRequest *domain.ApprovalRequest
	History         []*ApprovalHistoryItem
}

type GetApprovalHistoryUseCase struct {
	db                  ports.DB
	approvalRequestRepo ports.ApprovalRequestRepository
	approvalActionRepo  ports.ApprovalActionRepository
	employeeRepo        ports.EmployeeRepository
}

func NewGetApprovalHistoryUseCase(
	db ports.DB,
	approvalRequestRepo ports.ApprovalRequestRepository,
	approvalActionRepo ports.ApprovalActionRepository,
	employeeRepo ports.EmployeeRepository,
) *GetApprovalHistoryUseCase {
	return &GetApprovalHistoryUseCase{
		db:                  db,
		approvalRequestRepo: approvalRequestRepo,
		approvalActionRepo:  approvalActionRepo,
		employeeRepo:        employeeRepo,
	}
}

func (uc *GetApprovalHistoryUseCase) Execute(ctx context.Context, approvalRequestUID string) (*GetApprovalHistoryOutput, error) {
	approvalRequest, err := uc.approvalRequestRepo.GetByUID(ctx, uc.db, approvalRequestUID)
	if err != nil {
		return nil, err
	}
	if approvalRequest == nil {
		return nil, ErrApprovalRequestNotFound
	}

	actions, err := uc.approvalActionRepo.ListByRequest(ctx, uc.db, approvalRequest.UID)
	if err != nil {
		return nil, err
	}

	var history []*ApprovalHistoryItem
	for _, action := range actions {
		item := &ApprovalHistoryItem{
			Action:   action,
			ActorUID: action.ActorUID,
		}

		// Fetch actor name
		employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, action.ActorUID)
		if err != nil {
			return nil, err
		}
		if employee != nil {
			item.ActorName = employee.Name
		}

		history = append(history, item)
	}

	return &GetApprovalHistoryOutput{
		ApprovalRequest: approvalRequest,
		History:         history,
	}, nil
}
