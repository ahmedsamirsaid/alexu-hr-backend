package usecases

import (
	"context"
	"errors"
	"fmt"

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

	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionSentence := fmt.Sprintf(
		"%s activated attendance device '%s' at %s:%d",
		actorName, existing.Name, existing.IP, existing.Port,
	)

	// Audit log will fire after successful activation
	defer uc.auditor.From(ctx).
		Did(audit.ActionActivate).
		On(audit.EntityAttendanceDevice, uid).
		WithMeta("action", actionSentence).
		WithMeta("device_name", existing.Name).
		WithMeta("old_status", string(existing.Status)).
		WithMeta("new_status", string(domain.AttendanceDeviceStatusOffline)).
		Save(ctx)

	return uc.repo.UpdateStatus(ctx, uc.db, uid, domain.AttendanceDeviceStatusOffline)
}

var ErrAttendanceDeviceAlreadyActive = errors.New("device already active")
