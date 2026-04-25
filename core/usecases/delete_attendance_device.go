package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var ErrAttendanceDeviceNotFound = errors.New("device not found")

type DeleteAttendanceDeviceUseCase struct {
	db      ports.DB
	repo    ports.AttendanceDeviceRepository
	auditor audit.Auditor
}

func NewDeleteAttendanceDeviceUseCase(db ports.DB, repo ports.AttendanceDeviceRepository, auditor audit.Auditor) *DeleteAttendanceDeviceUseCase {
	return &DeleteAttendanceDeviceUseCase{db: db, repo: repo, auditor: auditor}
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

	// Audit log will fire after successful deletion (status change to deactivated)
	defer uc.auditor.From(ctx).
		Did(audit.ActionDelete).
		On(audit.EntityAttendanceDevice, uid).
		Save(ctx)

	return uc.repo.UpdateStatus(ctx, uc.db, uid, domain.AttendanceDeviceStatusDeactivated)
}
