package ports

import (
	"context"

	"github.com/banumusa/backend/core/domain"
)

type AttendanceDeviceListFilter struct {
	Status      *domain.AttendanceDeviceStatus
	Search      string
	SearchField string
}

type AttendanceDeviceRepository interface {
	Create(ctx context.Context, q Querier, device *domain.AttendanceDevice) error
	GetByUID(ctx context.Context, q Querier, uid string) (*domain.AttendanceDevice, error)
	GetBySerialNumber(ctx context.Context, q Querier, serialNumber string) (*domain.AttendanceDevice, error)
	GetByAddress(ctx context.Context, q Querier, ip string, port int) (*domain.AttendanceDevice, error)
	List(ctx context.Context, q Querier, filter AttendanceDeviceListFilter, limit, offset int) ([]*domain.AttendanceDevice, error)
	ListAll(ctx context.Context, q Querier) ([]*domain.AttendanceDevice, error)
	Count(ctx context.Context, q Querier) (int, error)
	CountFiltered(ctx context.Context, q Querier, filter AttendanceDeviceListFilter) (int, error)
	CountByStatus(ctx context.Context, q Querier, status domain.AttendanceDeviceStatus) (int, error)
	UpdateStatus(ctx context.Context, q Querier, uid string, status domain.AttendanceDeviceStatus) error
	Delete(ctx context.Context, q Querier, uid string) error
}
