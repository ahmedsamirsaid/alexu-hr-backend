package usecases

import (
	"context"
	"errors"
	"fmt"

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

	// Build human-readable action sentence
	actorName := audit.ActorFromContext(ctx)
	actionSentence := fmt.Sprintf(
		"%s deleted (deactivated) attendance device '%s' at %s:%d",
		actorName, existing.Name, existing.IP, existing.Port,
	)

	// Audit log will fire after successful deletion (status change to deactivated)
	defer uc.auditor.From(ctx).
		Did(audit.ActionDelete).
		On(audit.EntityAttendanceDevice, uid).
		WithMeta("action", actionSentence).
		WithMeta("device_name", existing.Name).
		WithMeta("old_status", string(existing.Status)).
		WithMeta("new_status", string(domain.AttendanceDeviceStatusDeactivated)).
		Save(ctx)

	return uc.repo.UpdateStatus(ctx, uc.db, uid, domain.AttendanceDeviceStatusDeactivated)
}
