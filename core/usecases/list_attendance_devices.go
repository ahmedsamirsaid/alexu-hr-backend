package usecases

import (
	"context"
	"time"

	"github.com/banumusa/backend/core/ports"
)

type ListAttendanceDevicesInput struct {
	Page     int
	PageSize int
}

type AttendanceDeviceListItem struct {
	UID          string `json:"uid"`
	IP           string `json:"ip"`
	Port         int    `json:"port"`
	Name         string `json:"name"`
	Location     string `json:"location"`
	SerialNumber string `json:"serialNumber"`
	Status       string `json:"status"`
	LastStatusKnownAt string `json:"lastStatusKnownAt"`
}

type ListAttendanceDevicesOutput struct {
	Devices    []AttendanceDeviceListItem `json:"devices"`
	Total      int                        `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"pageSize"`
	TotalPages int                        `json:"totalPages"`
}

type ListAttendanceDevicesUseCase struct {
	db   ports.DB
	repo ports.AttendanceDeviceRepository
}

func NewListAttendanceDevicesUseCase(db ports.DB, repo ports.AttendanceDeviceRepository) *ListAttendanceDevicesUseCase {
	return &ListAttendanceDevicesUseCase{db: db, repo: repo}
}

func (uc *ListAttendanceDevicesUseCase) Execute(ctx context.Context, input ListAttendanceDevicesInput) (*ListAttendanceDevicesOutput, error) {
	if input.Page < 1 {
		input.Page = 1
	}
	if input.PageSize < 1 {
		input.PageSize = 20
	}
	if input.PageSize > 100 {
		input.PageSize = 100
	}

	offset := (input.Page - 1) * input.PageSize

	devices, err := uc.repo.List(ctx, uc.db, input.PageSize, offset)
	if err != nil {
		return nil, err
	}

	total, err := uc.repo.Count(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	items := make([]AttendanceDeviceListItem, len(devices))
	for i, dev := range devices {
		items[i] = AttendanceDeviceListItem{
			UID:          dev.UID,
			IP:           dev.IP,
			Port:         dev.Port,
			Name:         dev.Name,
			Location:     dev.Location,
			SerialNumber: dev.SerialNumber,
			Status:       string(dev.Status),
			LastStatusKnownAt: dev.UpdatedAt.Format(time.RFC3339),
		}
	}

	totalPages := (total + input.PageSize - 1) / input.PageSize

	return &ListAttendanceDevicesOutput{
		Devices:    items,
		Total:      total,
		Page:       input.Page,
		PageSize:   input.PageSize,
		TotalPages: totalPages,
	}, nil
}
