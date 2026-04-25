package usecases

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type GetAttendanceLogHistoryInput struct {
	AttendanceLogUID string
}

type AttendanceLogHistoryItemOutput struct {
	UID           string  `json:"uid"`
	FieldChanged  string  `json:"fieldChanged"`
	OldValue      *string `json:"oldValue,omitempty"`
	NewValue      *string `json:"newValue,omitempty"`
	Reason        *string `json:"reason,omitempty"`
	EditedByUID   string  `json:"editedByUid"`
	EditedByName  string  `json:"editedByName"`
	EditedByPhone *string `json:"editedByPhone,omitempty"`
	EditedAt      string  `json:"editedAt"`
}

type GetAttendanceLogHistoryOutput struct {
	AttendanceLogUID string                           `json:"attendanceLogUid"`
	EmployeeUID      string                           `json:"employeeUid"`
	DepartmentUID    *string                          `json:"departmentUid,omitempty"`
	History          []AttendanceLogHistoryItemOutput `json:"history"`
}

type GetAttendanceLogHistoryUseCase struct {
	db           ports.DB
	recordRepo   ports.AttendanceRecordRepository
	auditRepo    ports.AuditLogRepository
	employeeRepo ports.EmployeeRepository
	userRepo     ports.UserRepository
}

func NewGetAttendanceLogHistoryUseCase(
	db ports.DB,
	recordRepo ports.AttendanceRecordRepository,
	auditRepo ports.AuditLogRepository,
	employeeRepo ports.EmployeeRepository,
	userRepo ports.UserRepository,
) *GetAttendanceLogHistoryUseCase {
	return &GetAttendanceLogHistoryUseCase{
		db:           db,
		recordRepo:   recordRepo,
		auditRepo:    auditRepo,
		employeeRepo: employeeRepo,
		userRepo:     userRepo,
	}
}

func (uc *GetAttendanceLogHistoryUseCase) Execute(ctx context.Context, input GetAttendanceLogHistoryInput) (*GetAttendanceLogHistoryOutput, error) {
	record, err := uc.recordRepo.GetByUID(ctx, uc.db, input.AttendanceLogUID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, ErrAttendanceLogNotFound
	}

	employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, record.EmployeeUID)
	if err != nil {
		return nil, err
	}
	if employee == nil {
		return nil, ErrEmployeeNotFound
	}

	// Fetch audit logs for this attendance record
	auditLogs, err := uc.auditRepo.ListByEntity(ctx, uc.db, "attendance_record", record.UID)
	if err != nil {
		return nil, err
	}

	output := &GetAttendanceLogHistoryOutput{
		AttendanceLogUID: record.UID,
		EmployeeUID:      record.EmployeeUID,
		DepartmentUID:    employee.DepartmentUID,
		History:          make([]AttendanceLogHistoryItemOutput, 0, len(auditLogs)),
	}

	// Convert audit logs to attendance history format
	for _, auditLog := range auditLogs {
		historyItems := uc.convertAuditLogToHistory(ctx, auditLog)
		output.History = append(output.History, historyItems...)
	}

	return output, nil
}

// convertAuditLogToHistory converts an audit log entry to one or more attendance history items
func (uc *GetAttendanceLogHistoryUseCase) convertAuditLogToHistory(ctx context.Context, auditLog *domain.AuditLog) []AttendanceLogHistoryItemOutput {
	items := make([]AttendanceLogHistoryItemOutput, 0)

	// Get actor name and phone
	actorName := auditLog.ActorUID
	var actorPhone *string
	if auditLog.ActorUID != "" && auditLog.ActorUID != "system_import" {
		user, err := uc.userRepo.GetByUID(ctx, uc.db, auditLog.ActorUID)
		if err == nil && user != nil {
			actorPhone = &user.Phone
			actorName = user.Phone
			// Try to get employee name
			if user.EmployeeUID != nil {
				employee, err := uc.employeeRepo.GetByUID(ctx, uc.db, *user.EmployeeUID)
				if err == nil && employee != nil {
					actorName = employee.Name
				}
			}
		}
	}

	// Parse metadata
	var meta map[string]interface{}
	if auditLog.Meta != nil {
		if err := json.Unmarshal([]byte(*auditLog.Meta), &meta); err == nil {
			// Handle different action types
			switch auditLog.Action {
			case "create":
				// For create action, show as "created_manually"
				reason := extractStringFromMeta(meta, "reason")
				items = append(items, AttendanceLogHistoryItemOutput{
					UID:           auditLog.UID,
					FieldChanged:  "created_manually",
					OldValue:      nil,
					NewValue:      nil,
					Reason:        reason,
					EditedByUID:   auditLog.ActorUID,
					EditedByName:  actorName,
					EditedByPhone: actorPhone,
					EditedAt:      auditLog.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
				})

			case "update":
				// For update action, extract field changes
				reason := extractStringFromMeta(meta, "reason")
				
				// Check for device_uid change
				if oldDeviceUID, newDeviceUID := extractStringFromMeta(meta, "old_device_uid"), extractStringFromMeta(meta, "new_device_uid"); oldDeviceUID != nil && newDeviceUID != nil {
					items = append(items, AttendanceLogHistoryItemOutput{
						UID:           auditLog.UID + "_device",
						FieldChanged:  "device_uid",
						OldValue:      oldDeviceUID,
						NewValue:      newDeviceUID,
						Reason:        reason,
						EditedByUID:   auditLog.ActorUID,
						EditedByName:  actorName,
						EditedByPhone: actorPhone,
						EditedAt:      auditLog.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
					})
				}

				// Check for punched_at change
				if oldPunchedAt, newPunchedAt := extractStringFromMeta(meta, "old_punched_at"), extractStringFromMeta(meta, "new_punched_at"); oldPunchedAt != nil && newPunchedAt != nil {
					items = append(items, AttendanceLogHistoryItemOutput{
						UID:           auditLog.UID + "_punched_at",
						FieldChanged:  "punched_at",
						OldValue:      oldPunchedAt,
						NewValue:      newPunchedAt,
						Reason:        reason,
						EditedByUID:   auditLog.ActorUID,
						EditedByName:  actorName,
						EditedByPhone: actorPhone,
						EditedAt:      auditLog.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
					})
				}

				// Check for punch_type change
				if oldPunchType, newPunchType := extractStringFromMeta(meta, "old_punch_type"), extractStringFromMeta(meta, "new_punch_type"); oldPunchType != nil && newPunchType != nil {
					items = append(items, AttendanceLogHistoryItemOutput{
						UID:           auditLog.UID + "_punch_type",
						FieldChanged:  "punch_type",
						OldValue:      oldPunchType,
						NewValue:      newPunchType,
						Reason:        reason,
						EditedByUID:   auditLog.ActorUID,
						EditedByName:  actorName,
						EditedByPhone: actorPhone,
						EditedAt:      auditLog.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
					})
				}

				// If no specific field changes found, show as "updated_manually"
				if len(items) == 0 {
					items = append(items, AttendanceLogHistoryItemOutput{
						UID:           auditLog.UID,
						FieldChanged:  "updated_manually",
						OldValue:      nil,
						NewValue:      nil,
						Reason:        reason,
						EditedByUID:   auditLog.ActorUID,
						EditedByName:  actorName,
						EditedByPhone: actorPhone,
						EditedAt:      auditLog.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
					})
				}

			case "delete":
				// For delete action
				items = append(items, AttendanceLogHistoryItemOutput{
					UID:           auditLog.UID,
					FieldChanged:  "deleted",
					OldValue:      nil,
					NewValue:      nil,
					Reason:        extractStringFromMeta(meta, "reason"),
					EditedByUID:   auditLog.ActorUID,
					EditedByName:  actorName,
					EditedByPhone: actorPhone,
					EditedAt:      auditLog.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
				})
			}
		}
	}

	// If no items were created (no metadata or parsing failed), create a generic entry
	if len(items) == 0 {
		items = append(items, AttendanceLogHistoryItemOutput{
			UID:           auditLog.UID,
			FieldChanged:  strings.ToLower(auditLog.Action),
			OldValue:      nil,
			NewValue:      nil,
			Reason:        nil,
			EditedByUID:   auditLog.ActorUID,
			EditedByName:  actorName,
			EditedByPhone: actorPhone,
			EditedAt:      auditLog.OccurredAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	return items
}

// extractStringFromMeta safely extracts a string value from metadata map
func extractStringFromMeta(meta map[string]interface{}, key string) *string {
	if val, ok := meta[key]; ok {
		if strVal, ok := val.(string); ok && strVal != "" {
			return &strVal
		}
	}
	return nil
}
