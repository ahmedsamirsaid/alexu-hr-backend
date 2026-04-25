package usecases

import (
	"context"
	"errors"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type ActivateAttendanceDeviceUseCase struct {
	db      ports.DB
	repo    ports.AttendanceDeviceRepository
	auditor audit.Auditor
}

func NewActivateAttendanceDeviceUseCase(db ports.DB, repo ports.AttendanceDeviceRepository, auditor audit.Auditor) *ActivateAttendanceDeviceUseCase {
	return &ActivateAttendanceDeviceUseCase{db: db, repo: repo, auditor: auditor}
}

func (uc *ActivateAttendanceDeviceUseCase) Execute(ctx context.Context, uid string) error {
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

	if existing.Status != domain.AttendanceDeviceStatusDeactivated {
		return ErrAttendanceDeviceAlreadyActive
	}

	// Audit log will fire after successful activation
	defer uc.auditor.From(ctx).
		Did(audit.ActionActivate).
		On(audit.EntityAttendanceDevice, uid).
		Save(ctx)

	return uc.repo.UpdateStatus(ctx, uc.db, uid, domain.AttendanceDeviceStatusOffline)
}

var ErrAttendanceDeviceAlreadyActive = errors.New("device already active")
