package usecases

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/banumusa/backend/core/ports"
)

type UpdateAttendanceDeviceInput struct {
	UID      string
	IP       string
	Port     int
	Name     string
	Location string
}

type UpdateAttendanceDeviceOutput struct {
	UID               string `json:"uid"`
	IP                string `json:"ip"`
	Port              int    `json:"port"`
	Name              string `json:"name"`
	Location          string `json:"location"`
	SerialNumber      string `json:"serialNumber"`
	Status            string `json:"status"`
	LastStatusKnownAt string `json:"lastStatusKnownAt"`
}

type UpdateAttendanceDeviceUseCase struct {
	db   ports.DB
	repo ports.AttendanceDeviceRepository
}

func NewUpdateAttendanceDeviceUseCase(db ports.DB, repo ports.AttendanceDeviceRepository) *UpdateAttendanceDeviceUseCase {
	return &UpdateAttendanceDeviceUseCase{db: db, repo: repo}
}

func (uc *UpdateAttendanceDeviceUseCase) Execute(ctx context.Context, input UpdateAttendanceDeviceInput) (*UpdateAttendanceDeviceOutput, error) {
	input.UID = strings.TrimSpace(input.UID)
	input.IP = strings.TrimSpace(input.IP)
	input.Name = strings.TrimSpace(input.Name)
	input.Location = strings.TrimSpace(input.Location)

	if input.UID == "" {
		return nil, ErrAttendanceDeviceNotFound
	}
	if input.Name == "" {
		return nil, ErrDeviceNameRequired
	}
	if input.IP == "" {
		return nil, ErrDeviceIPRequired
	}
	if net.ParseIP(input.IP) == nil {
		return nil, ErrDeviceInvalidIP
	}
	if input.Port < 1 || input.Port > 65535 {
		return nil, ErrDeviceInvalidPort
	}

	existing, err := uc.repo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrAttendanceDeviceNotFound
	}

	conflict, err := uc.repo.GetByAddress(ctx, uc.db, input.IP, input.Port)
	if err != nil {
		return nil, err
	}
	if conflict != nil && conflict.UID != input.UID {
		return nil, ErrDeviceAddressExists
	}

	existing.IP = input.IP
	existing.Port = input.Port
	existing.Name = input.Name
	existing.Location = input.Location

	if err := uc.repo.Update(ctx, uc.db, existing); err != nil {
		return nil, err
	}

	return &UpdateAttendanceDeviceOutput{
		UID:               existing.UID,
		IP:                existing.IP,
		Port:              existing.Port,
		Name:              existing.Name,
		Location:          existing.Location,
		SerialNumber:      existing.SerialNumber,
		Status:            string(existing.Status),
		LastStatusKnownAt: existing.UpdatedAt.Format(time.RFC3339),
	}, nil
}
