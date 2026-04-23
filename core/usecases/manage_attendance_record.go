package usecases

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/banumusa/backend/core/audit"
	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

var (
	ErrAttendanceLogNotFound              = errors.New("attendance log not found")
	ErrAttendanceLogEmployeeRequired      = errors.New("employeeUid is required")
	ErrAttendanceLogDeviceRequired        = errors.New("deviceUid is required")
	ErrAttendanceLogDeviceUserIDReq       = errors.New("deviceUserId is required")
	ErrAttendanceLogPunchedAtReq          = errors.New("punchedAt is required")
	ErrAttendanceLogPunchTypeInvalid      = errors.New("punchType must be one of: check_in, check_out, break_start, break_end, unknown")
	ErrAttendanceLogConflict              = errors.New("attendance log already exists for the same employee, date, and punch type")
	ErrAttendanceLogCheckoutBeforeCheckin = errors.New("check-out time cannot be earlier than the existing check-in time on the same date")
)

type CreateAttendanceLogInput struct {
	EmployeeUID string
	DeviceUID   string
	PunchedAt   time.Time
	PunchType   string
}

type UpdateAttendanceLogInput struct {
	UID       string
	DeviceUID string
	PunchedAt time.Time
	PunchType string
}

type AttendanceLogOutput struct {
	UID          string  `json:"uid"`
	EmployeeUID  string  `json:"employeeUid"`
	DeviceUID    string  `json:"deviceUid"`
	DeviceUserID string  `json:"deviceUserId"`
	PunchedAt    string  `json:"punchedAt"`
	PunchType    string  `json:"punchType"`
	RawPayload   *string `json:"rawPayload,omitempty"`
	CreatedAt    string  `json:"createdAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type CreateAttendanceLogUseCase struct {
	db           ports.DB
	recordRepo   ports.AttendanceRecordRepository
	employeeRepo ports.EmployeeRepository
	deviceRepo   ports.AttendanceDeviceRepository
	auditor      audit.Auditor
}

func NewCreateAttendanceLogUseCase(
	db ports.DB,
	recordRepo ports.AttendanceRecordRepository,
	employeeRepo ports.EmployeeRepository,
	deviceRepo ports.AttendanceDeviceRepository,
	auditor audit.Auditor,
) *CreateAttendanceLogUseCase {
	return &CreateAttendanceLogUseCase{
		db:           db,
		recordRepo:   recordRepo,
		employeeRepo: employeeRepo,
		deviceRepo:   deviceRepo,
		auditor:      auditor,
	}
}

func (uc *CreateAttendanceLogUseCase) Execute(ctx context.Context, input CreateAttendanceLogInput) (*AttendanceLogOutput, error) {
	record, err := uc.buildNewRecord(ctx, input.EmployeeUID, input.DeviceUID, input.PunchedAt, input.PunchType)
	if err != nil {
		return nil, err
	}

	// Audit log will fire after successful creation
	defer uc.auditor.From(ctx).
		Did(audit.ActionCreate).
		On(audit.EntityAttendanceRecord, record.UID).
		WithMeta("employee_uid", record.EmployeeUID).
		WithMeta("device_uid", record.DeviceUID).
		WithMeta("punched_at", record.PunchedAt.Format(time.RFC3339)).
		WithMeta("punch_type", string(record.PunchType)).
		Save(ctx)

	existingRecords, err := uc.recordRepo.ListByDate(ctx, uc.db, record.PunchedAt, &record.EmployeeUID)
	if err != nil {
		return nil, err
	}
	for _, existing := range existingRecords {
		if existing.PunchType == record.PunchType {
			return nil, ErrAttendanceLogConflict
		}
		if record.PunchType == domain.AttendancePunchTypeCheckOut &&
			existing.PunchType == domain.AttendancePunchTypeCheckIn &&
			record.PunchedAt.Before(existing.PunchedAt) {
			return nil, ErrAttendanceLogCheckoutBeforeCheckin
		}
	}

	created, err := uc.recordRepo.Create(ctx, uc.db, record)
	if err != nil {
		if isAttendanceRecordConflictError(err) {
			return nil, ErrAttendanceLogConflict
		}
		return nil, err
	}
	if !created {
		return nil, ErrAttendanceLogConflict
	}

	return buildAttendanceLogOutput(record), nil
}

func (uc *CreateAttendanceLogUseCase) buildNewRecord(ctx context.Context, employeeUID, deviceUID string, punchedAt time.Time, punchType string) (*domain.AttendanceRecord, error) {
	deviceUserID := domain.GenerateUID("dusr")
	return uc.buildRecord(ctx, employeeUID, deviceUID, deviceUserID, punchedAt, punchType, nil, nil)
}

func (uc *CreateAttendanceLogUseCase) buildRecord(ctx context.Context, employeeUID, deviceUID, deviceUserID string, punchedAt time.Time, punchType string, rawPayload *string, existing *domain.AttendanceRecord) (*domain.AttendanceRecord, error) {
	employeeUID = strings.TrimSpace(employeeUID)
	deviceUID = strings.TrimSpace(deviceUID)
	deviceUserID = strings.TrimSpace(deviceUserID)
	punchType = strings.TrimSpace(strings.ToLower(punchType))

	if employeeUID == "" {
		return nil, ErrAttendanceLogEmployeeRequired
	}
	if deviceUID == "" {
		return nil, ErrAttendanceLogDeviceRequired
	}
	if deviceUserID == "" {
		return nil, ErrAttendanceLogDeviceUserIDReq
	}
	if punchedAt.IsZero() {
		return nil, ErrAttendanceLogPunchedAtReq
	}

	parsedPunchType, err := parseAttendancePunchType(punchType)
	if err != nil {
		return nil, err
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, employeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	device, err := uc.deviceRepo.GetByUID(ctx, uc.db, deviceUID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrAttendanceDeviceNotFound
	}

	if rawPayload != nil {
		trimmed := strings.TrimSpace(*rawPayload)
		rawPayload = &trimmed
	}

	if existing != nil {
		existing.EmployeeUID = employeeUID
		existing.DeviceUID = deviceUID
		existing.DeviceUserID = deviceUserID
		existing.PunchedAt = punchedAt.UTC()
		existing.PunchType = parsedPunchType
		existing.RawPayload = rawPayload
		return existing, nil
	}

	return domain.NewAttendanceRecord(employeeUID, deviceUID, deviceUserID, punchedAt.UTC(), parsedPunchType, rawPayload), nil
}

type UpdateAttendanceLogUseCase struct {
	db           ports.DB
	recordRepo   ports.AttendanceRecordRepository
	employeeRepo ports.EmployeeRepository
	deviceRepo   ports.AttendanceDeviceRepository
	auditor      audit.Auditor
}

func NewUpdateAttendanceLogUseCase(
	db ports.DB,
	recordRepo ports.AttendanceRecordRepository,
	employeeRepo ports.EmployeeRepository,
	deviceRepo ports.AttendanceDeviceRepository,
	auditor audit.Auditor,
) *UpdateAttendanceLogUseCase {
	return &UpdateAttendanceLogUseCase{
		db:           db,
		recordRepo:   recordRepo,
		employeeRepo: employeeRepo,
		deviceRepo:   deviceRepo,
		auditor:      auditor,
	}
}

func (uc *UpdateAttendanceLogUseCase) Execute(ctx context.Context, input UpdateAttendanceLogInput) (*AttendanceLogOutput, error) {
	input.UID = strings.TrimSpace(input.UID)
	if input.UID == "" {
		return nil, ErrAttendanceLogNotFound
	}

	existing, err := uc.recordRepo.GetByUID(ctx, uc.db, input.UID)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrAttendanceLogNotFound
	}

	// Capture old values for audit metadata
	oldDeviceUID := existing.DeviceUID
	oldPunchedAt := existing.PunchedAt
	oldPunchType := existing.PunchType

	record, err := uc.buildUpdatedRecord(ctx, existing, input.DeviceUID, input.PunchedAt, input.PunchType)
	if err != nil {
		return nil, err
	}

	// Audit log will fire after successful update with field changes
	defer uc.auditor.From(ctx).
		Did(audit.ActionUpdate).
		On(audit.EntityAttendanceRecord, record.UID).
		WithMeta("old_device_uid", oldDeviceUID).
		WithMeta("new_device_uid", record.DeviceUID).
		WithMeta("old_punched_at", oldPunchedAt.Format(time.RFC3339)).
		WithMeta("new_punched_at", record.PunchedAt.Format(time.RFC3339)).
		WithMeta("old_punch_type", string(oldPunchType)).
		WithMeta("new_punch_type", string(record.PunchType)).
		Save(ctx)

	if err := uc.recordRepo.Update(ctx, uc.db, record); err != nil {
		if isAttendanceRecordConflictError(err) {
			return nil, ErrAttendanceLogConflict
		}
		return nil, err
	}

	return buildAttendanceLogOutput(record), nil
}

func (uc *UpdateAttendanceLogUseCase) buildUpdatedRecord(ctx context.Context, existing *domain.AttendanceRecord, deviceUID string, punchedAt time.Time, punchType string) (*domain.AttendanceRecord, error) {
	deviceUID = strings.TrimSpace(deviceUID)
	punchType = strings.TrimSpace(strings.ToLower(punchType))

	if deviceUID == "" {
		return nil, ErrAttendanceLogDeviceRequired
	}
	if punchedAt.IsZero() {
		return nil, ErrAttendanceLogPunchedAtReq
	}

	parsedPunchType, err := parseAttendancePunchType(punchType)
	if err != nil {
		return nil, err
	}

	device, err := uc.deviceRepo.GetByUID(ctx, uc.db, deviceUID)
	if err != nil {
		return nil, err
	}
	if device == nil {
		return nil, ErrAttendanceDeviceNotFound
	}

	existing.DeviceUID = deviceUID
	existing.PunchedAt = punchedAt.UTC()
	existing.PunchType = parsedPunchType
	return existing, nil
}

func parseAttendancePunchType(value string) (domain.AttendancePunchType, error) {
	switch domain.AttendancePunchType(value) {
	case domain.AttendancePunchTypeCheckIn,
		domain.AttendancePunchTypeCheckOut,
		domain.AttendancePunchTypeBreakStart,
		domain.AttendancePunchTypeBreakEnd,
		domain.AttendancePunchTypeUnknown:
		return domain.AttendancePunchType(value), nil
	default:
		return "", ErrAttendanceLogPunchTypeInvalid
	}
}

func buildAttendanceLogOutput(record *domain.AttendanceRecord) *AttendanceLogOutput {
	return &AttendanceLogOutput{
		UID:          record.UID,
		EmployeeUID:  record.EmployeeUID,
		DeviceUID:    record.DeviceUID,
		DeviceUserID: record.DeviceUserID,
		PunchedAt:    record.PunchedAt.Format(time.RFC3339),
		PunchType:    string(record.PunchType),
		RawPayload:   record.RawPayload,
		CreatedAt:    record.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    record.UpdatedAt.Format(time.RFC3339),
	}
}

func isAttendanceRecordConflictError(err error) bool {
	return strings.Contains(strings.ToLower(err.Error()), "unique constraint failed")
}
