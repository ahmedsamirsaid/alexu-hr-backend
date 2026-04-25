package usecases

import (
	"context"
	"errors"
	"net"
	"strings"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var (
	ErrDeviceNameRequired         = errors.New("name is required")
	ErrDeviceIPRequired           = errors.New("ip is required")
	ErrDeviceInvalidIP            = errors.New("ip must be a valid IPv4/IPv6 address")
	ErrDeviceInvalidPort          = errors.New("port must be between 1 and 65535")
	ErrDeviceSerialNumberRequired = errors.New("serialNumber is required")
	ErrDeviceSerialNumberExists   = errors.New("a device with this serial number already exists")
	ErrDeviceAddressExists        = errors.New("a device with this ip and port already exists")
)

type RegisterAttendanceDeviceInput struct {
	IP           string
	Port         int
	Name         string
	Location     string
	SerialNumber string
}

type RegisterAttendanceDeviceOutput struct {
	UID string `json:"uid"`
}

type RegisterAttendanceDeviceUseCase struct {
	db      ports.DB
	repo    ports.AttendanceDeviceRepository
	auditor audit.Auditor
}

func NewRegisterAttendanceDeviceUseCase(db ports.DB, repo ports.AttendanceDeviceRepository, auditor audit.Auditor) *RegisterAttendanceDeviceUseCase {
	return &RegisterAttendanceDeviceUseCase{db: db, repo: repo, auditor: auditor}
}

func (uc *RegisterAttendanceDeviceUseCase) Execute(ctx context.Context, input RegisterAttendanceDeviceInput) (*RegisterAttendanceDeviceOutput, error) {
	input.IP = strings.TrimSpace(input.IP)
	input.Name = strings.TrimSpace(input.Name)
	input.Location = strings.TrimSpace(input.Location)
	input.SerialNumber = strings.TrimSpace(input.SerialNumber)

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
	if input.SerialNumber == "" {
		return nil, ErrDeviceSerialNumberRequired
	}

	byAddress, err := uc.repo.GetByAddress(ctx, uc.db, input.IP, input.Port)
	if err != nil {
		return nil, err
	}
	if byAddress != nil {
		return nil, ErrDeviceAddressExists
	}

	existing, err := uc.repo.GetBySerialNumber(ctx, uc.db, input.SerialNumber)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDeviceSerialNumberExists
	}

	device := domain.NewAttendanceDevice(input.IP, input.Port, input.Name, input.Location, input.SerialNumber)
	
	// Audit log will fire after successful device registration
	defer uc.auditor.From(ctx).
		Did(audit.ActionCreate).
		On(audit.EntityAttendanceDevice, device.UID).
		WithMeta("ip", device.IP).
		WithMeta("port", device.Port).
		WithMeta("serial_number", device.SerialNumber).
		WithMeta("name", device.Name).
		WithMeta("location", device.Location).
		Save(ctx)
	
	if err := uc.repo.Create(ctx, uc.db, device); err != nil {
		return nil, err
	}

	return &RegisterAttendanceDeviceOutput{UID: device.UID}, nil
}
