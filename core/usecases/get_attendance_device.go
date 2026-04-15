package usecases

import (
	"context"
	"strings"
	"time"

	"github.com/banumusa/backend/core/ports"
)

type GetAttendanceDeviceOutput struct {
	UID               string `json:"uid"`
	IP                string `json:"ip"`
	Port              int    `json:"port"`
	Name              string `json:"name"`
	Location          string `json:"location"`
	SerialNumber      string `json:"serialNumber"`
	Status            string `json:"status"`
	LastStatusKnownAt string `json:"lastStatusKnownAt"`
}

type GetAttendanceDeviceUseCase struct {
	db   ports.DB
	repo ports.AttendanceDeviceRepository
}

func NewGetAttendanceDeviceUseCase(db ports.DB, repo ports.AttendanceDeviceRepository) *GetAttendanceDeviceUseCase {
	return &GetAttendanceDeviceUseCase{db: db, repo: repo}
}

func (uc *GetAttendanceDeviceUseCase) Execute(ctx context.Context, uid string) (*GetAttendanceDeviceOutput, error) {
	uid = strings.TrimSpace(uid)
	if uid == "" {
		return nil, ErrAttendanceDeviceNotFound
	}

	device, err := uc.repo.GetByUID(ctx, uc.db, uid)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrAttendanceDeviceNotFound
	}

	return &GetAttendanceDeviceOutput{
		UID:               device.UID,
		IP:                device.IP,
		Port:              device.Port,
		Name:              device.Name,
		Location:          device.Location,
		SerialNumber:      device.SerialNumber,
		Status:            string(device.Status),
		LastStatusKnownAt: device.UpdatedAt.Format(time.RFC3339),
	}, nil
}
