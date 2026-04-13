package usecases

import (
	"context"
	cryptorand "crypto/rand"
	"math/big"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type CheckAllAttendanceDevicesConnectionOutput struct {
	CheckedCount int    `json:"checkedCount"`
	Online       int    `json:"online"`
	Offline      int    `json:"offline"`
	CheckedAt    string `json:"checkedAt"`
}

type CheckAllAttendanceDevicesConnectionUseCase struct {
	db   ports.DB
	repo ports.AttendanceDeviceRepository
}

func NewCheckAllAttendanceDevicesConnectionUseCase(db ports.DB, repo ports.AttendanceDeviceRepository) *CheckAllAttendanceDevicesConnectionUseCase {
	return &CheckAllAttendanceDevicesConnectionUseCase{db: db, repo: repo}
}

func (uc *CheckAllAttendanceDevicesConnectionUseCase) Execute(ctx context.Context) (*CheckAllAttendanceDevicesConnectionOutput, error) {
	devices, err := uc.repo.ListAll(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	online := 0
	offline := 0
	checkedAt := time.Now()

	for _, device := range devices {
		status := uc.checkConnectivity()
		if err := uc.repo.UpdateStatus(ctx, uc.db, device.UID, status); err != nil {
			return nil, err
		}

		if status == domain.AttendanceDeviceStatusOnline {
			online++
		} else {
			offline++
		}
	}

	return &CheckAllAttendanceDevicesConnectionOutput{
		CheckedCount: len(devices),
		Online:       online,
		Offline:      offline,
		CheckedAt:    checkedAt.Format(time.RFC3339),
	}, nil
}

func (uc *CheckAllAttendanceDevicesConnectionUseCase) checkConnectivity() domain.AttendanceDeviceStatus {
	// TODO: Replace with real device reachability check when Integration with actual devices is implemented.
	n, err := cryptorand.Int(cryptorand.Reader, big.NewInt(2))
	if err == nil && n.Int64() == 0 {
		return domain.AttendanceDeviceStatusOnline
	}
	return domain.AttendanceDeviceStatusOffline
}
