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

		// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionParams := map[string]interface{}{
		"Actor":  actorName,
		"Device": existing.Name,
		"IP":     existing.IP,
		"Port":   existing.Port,
	}

	// Audit log will fire after successful activation
	defer uc.auditor.From(ctx).
		Did(audit.ActionActivate).
		On(audit.EntityAttendanceDevice, uid).
		WithMeta("action_key", "audit.sentence.activate_attendance_device").
		WithMeta("action_params", actionParams).
		WithMeta("device_name", existing.Name).
		WithMeta("old_status", string(existing.Status)).
		WithMeta("new_status", string(domain.AttendanceDeviceStatusOffline)).
		Save(ctx)
	return uc.repo.UpdateStatus(ctx, uc.db, uid, domain.AttendanceDeviceStatusOffline)
}

var ErrAttendanceDeviceAlreadyActive = errors.New("device already active")
