package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type LeaveRequestDocumentRepository interface {
	Create(ctx context.Context, q Querier, doc *domain.LeaveRequestDocument) error
	ListByLeaveRequestUID(ctx context.Context, q Querier, leaveRequestUID string) ([]*domain.LeaveRequestDocument, error)
	GetByLeaveRequestUIDAndFileName(ctx context.Context, q Querier, leaveRequestUID, fileName string) (*domain.LeaveRequestDocument, error)
	UpdateObjectKey(ctx context.Context, q Querier, leaveRequestUID, fileName, objectKey string) error
}
