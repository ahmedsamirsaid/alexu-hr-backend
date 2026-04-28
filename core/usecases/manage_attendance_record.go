package usecases

import (
	"context"
	"errors"
	"fmt"
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
	EditedByUID string
	Reason      *string
}

type UpdateAttendanceLogInput struct {
	UID         string
	DeviceUID   string
	PunchedAt   time.Time
	PunchType   string
	EditedByUID string
	Reason      *string
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

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Get employee name for human-readable action sentence
	employee, _ := uc.employeeRepo.GetByUID(ctx, uc.db, record.EmployeeUID)
	employeeName := record.EmployeeUID
	if employee != nil {
		employeeName = employee.Name
	}

	actorName := audit.ActorFromContext(ctx)
	
	// Format time in Cairo timezone for audit log display
	cairoLoc, _ := time.LoadLocation("Africa/Cairo")
	if cairoLoc == nil {
		cairoLoc = time.FixedZone("Africa/Cairo", 2*60*60) // Fallback to UTC+2
	}
	localTime := record.PunchedAt.In(cairoLoc)
	
	actionSentence := fmt.Sprintf(
		"%s created attendance record for %s (Employee) — %s at %s",
		actorName, employeeName, string(record.PunchType), localTime.Format("Jan 2, 2006 3:04 PM"),
	)

	// Audit log will fire after successful creation
	// we will save the audit log only if the transaction commits successfully

	existingRecords, err := uc.recordRepo.ListByDate(ctx, tx, record.PunchedAt, &record.EmployeeUID)
	if err != nil {
		return nil, err
	}
	for _, existing := range existingRecords {
		if record.PunchType == domain.AttendancePunchTypeCheckOut &&
			existing.PunchType == domain.AttendancePunchTypeCheckIn &&
			record.PunchedAt.Before(existing.PunchedAt) {
			return nil, ErrAttendanceLogCheckoutBeforeCheckin
		}
	}

	created, err := uc.recordRepo.Create(ctx, tx, record)
	if err != nil {
		if isAttendanceRecordConflictError(err) {
			return nil, ErrAttendanceLogConflict
		}
		return nil, err
	}
	if !created {
		return nil, ErrAttendanceLogConflict
	}

	// Note: Audit logging is handled after successful transaction commit
	// No need to write to attendance_edit_history table - audit_logs table is used instead

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	// device name not uid 
	device, _ := uc.deviceRepo.GetByUID(ctx, uc.db, record.DeviceUID)

	uc.auditor.From(ctx).
		Did(audit.ActionCreate).
		On(audit.EntityAttendanceRecord, record.UID).
		WithMeta("action", actionSentence).
		WithMeta("employee_uid", record.EmployeeUID).
		WithMeta("employee_name", employeeName).
		WithMeta("device_uid", record.DeviceUID).
		WithMeta("device_name", device.Name).
		WithMeta("punched_at", record.PunchedAt.Format(time.RFC3339)).
		WithMeta("punch_type", string(record.PunchType)).
		WithMeta("reason", normalizeOptionalString(input.Reason)).
		SaveSync(ctx)

	return buildAttendanceLogOutput(record), nil
}

func (uc *CreateAttendanceLogUseCase) buildNewRecord(ctx context.Context, employeeUID, deviceUID string, punchedAt time.Time, punchType string) (*domain.AttendanceRecord, error) {
	// For manually created attendance logs, use the employee's university ID as device_user_id
	// to ensure the unique constraint works properly and prevents duplicate entries
	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, employeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}
	
	deviceUserID := employee.UniversityID
	if deviceUserID == "" {
		deviceUserID = employee.GovernmentID
	}
	if deviceUserID == "" {
		// Fallback to generating a UID if neither ID is available
		deviceUserID = domain.GenerateUID("dusr")
	}
	
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

	tx, err := uc.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	existing, err := uc.recordRepo.GetByUID(ctx, tx, input.UID)
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
	// Get employee name for human-readable action sentence
	employee, _ := uc.employeeRepo.GetByUID(ctx, tx, record.EmployeeUID)
	employeeName := record.EmployeeUID
	if employee != nil {
		employeeName = employee.Name
	}

	actorName := audit.ActorFromContext(ctx)
	
	// Format times in Cairo timezone for audit log display
	cairoLoc, _ := time.LoadLocation("Africa/Cairo")
	if cairoLoc == nil {
		cairoLoc = time.FixedZone("Africa/Cairo", 2*60*60) // Fallback to UTC+2
	}
	oldLocalTime := oldPunchedAt.In(cairoLoc)
	newLocalTime := record.PunchedAt.In(cairoLoc)
	
	actionSentence := fmt.Sprintf(
		"%s updated attendance record for %s (Employee) — changed from %s at %s to %s at %s",
		actorName, employeeName,
		string(oldPunchType), oldLocalTime.Format("Jan 2, 3:04 PM"),
		string(record.PunchType), newLocalTime.Format("Jan 2, 3:04 PM"),
	)

	// Audit log will fire after successful update
	// we will save the audit log only if the transaction commits successfully
	if err := uc.recordRepo.Update(ctx, tx, record); err != nil {
		if isAttendanceRecordConflictError(err) {
			return nil, ErrAttendanceLogConflict
		}
		return nil, err
	}

	// Note: Audit logging is handled below
	// No need to write to attendance_edit_history table - audit_logs table is used instead

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	uc.auditor.From(ctx).
		Did(audit.ActionUpdate).
		On(audit.EntityAttendanceRecord, record.UID).
		WithMeta("action", actionSentence).
		WithMeta("employee_uid", record.EmployeeUID).
		WithMeta("employee_name", employeeName).
		WithMeta("old_device_uid", oldDeviceUID).
		WithMeta("new_device_uid", record.DeviceUID).
		WithMeta("old_punched_at", oldPunchedAt.Format(time.RFC3339)).
		WithMeta("new_punched_at", record.PunchedAt.Format(time.RFC3339)).
		WithMeta("old_punch_type", string(oldPunchType)).
		WithMeta("new_punch_type", string(record.PunchType)).
		WithMeta("reason", normalizeOptionalString(input.Reason)).
		SaveSync(ctx)

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

func normalizeOptionalString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
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
