package usecases

import (
	"context"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type AttendanceDeviceStatsOutput struct {
	Total       int `json:"total"`
	Online      int `json:"online"`
	Offline     int `json:"offline"`
	Deactivated int `json:"deactivated"`
}

type GetAttendanceDeviceStatsUseCase struct {
	db   ports.DB
	repo ports.AttendanceDeviceRepository
}

func NewGetAttendanceDeviceStatsUseCase(db ports.DB, repo ports.AttendanceDeviceRepository) *GetAttendanceDeviceStatsUseCase {
	return &GetAttendanceDeviceStatsUseCase{db: db, repo: repo}
}

func (uc *GetAttendanceDeviceStatsUseCase) Execute(ctx context.Context) (*AttendanceDeviceStatsOutput, error) {
	total, err := uc.repo.Count(ctx, uc.db)
	if err != nil {
		return nil, err
	}

	online, err := uc.repo.CountByStatus(ctx, uc.db, domain.AttendanceDeviceStatusOnline)
	if err != nil {
		return nil, err
	}

	offline, err := uc.repo.CountByStatus(ctx, uc.db, domain.AttendanceDeviceStatusOffline)
	if err != nil {
		return nil, err
	}

	deactivated, err := uc.repo.CountByStatus(ctx, uc.db, domain.AttendanceDeviceStatusDeactivated)
	if err != nil {
		return nil, err
	}

	return &AttendanceDeviceStatsOutput{
		Total:       total,
		Online:      online,
		Offline:     offline,
		Deactivated: deactivated,
	}, nil
}
