package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/ports"
)

var ErrAttendanceDeviceNotFound = errors.New("device not found")

type DeleteAttendanceDeviceUseCase struct {
	db   ports.DB
	repo ports.AttendanceDeviceRepository
}

func NewDeleteAttendanceDeviceUseCase(db ports.DB, repo ports.AttendanceDeviceRepository) *DeleteAttendanceDeviceUseCase {
	return &DeleteAttendanceDeviceUseCase{db: db, repo: repo}
}

func (uc *DeleteAttendanceDeviceUseCase) Execute(ctx context.Context, uid string) error {
	if uid == "" {
		return ErrAttendanceDeviceNotFound
	}

	existing, err := uc.repo.GetByUID(ctx, uc.db, uid)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrAttendanceDeviceNotFound
	}

	return uc.repo.Delete(ctx, uc.db, uid)
}
