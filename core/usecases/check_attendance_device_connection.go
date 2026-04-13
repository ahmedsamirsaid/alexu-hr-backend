package usecases

import (
	"context"
	cryptorand "crypto/rand"
	"errors"
	"math/big"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var ErrCheckDeviceUIDRequired = errors.New("device uid is required")

type CheckAttendanceDeviceConnectionOutput struct {
	UID       string `json:"uid"`
	Status    string `json:"status"`
	Reachable bool   `json:"reachable"`
	CheckedAt string `json:"checkedAt"`
}

type CheckAttendanceDeviceConnectionUseCase struct {
	db   ports.DB
	repo ports.AttendanceDeviceRepository
}

func NewCheckAttendanceDeviceConnectionUseCase(db ports.DB, repo ports.AttendanceDeviceRepository) *CheckAttendanceDeviceConnectionUseCase {
	return &CheckAttendanceDeviceConnectionUseCase{db: db, repo: repo}
}

func (uc *CheckAttendanceDeviceConnectionUseCase) Execute(ctx context.Context, uid string) (*CheckAttendanceDeviceConnectionOutput, error) {
	if uid == "" {
		return nil, ErrCheckDeviceUIDRequired
	}

	device, err := uc.repo.GetByUID(ctx, uc.db, uid)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrAttendanceDeviceNotFound
	}

	status := uc.checkConnectivity(ctx, device.IP, device.Port)
	checkedAt := time.Now()
	_ = uc.repo.UpdateStatus(ctx, uc.db, device.UID, status)

	return &CheckAttendanceDeviceConnectionOutput{
		UID:       device.UID,
		Status:    string(status),
		Reachable: status == domain.AttendanceDeviceStatusOnline,
		CheckedAt: checkedAt.Format(time.RFC3339),
	}, nil
}

func (uc *CheckAttendanceDeviceConnectionUseCase) checkConnectivity(ctx context.Context, ip string, port int) domain.AttendanceDeviceStatus {
	// TODO: Replace with real device reachability check.
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(2))
	if err == nil && n.Int64() == 0 {
		return domain.AttendanceDeviceStatusOnline
	}
	return domain.AttendanceDeviceStatusOffline
}
