package domain

import "time"

// AuditLog records who did what and when on which entity.
// It is the domain model backing the audit trail feature.
type AuditLog struct {
	ID         int64
	UID        string
	ActorUID   string    // employee UID of the person who performed the action
	Action     string    // e.g. "approve", "update", "create"
	EntityType string    // e.g. "leave_request", "attendance_record"
	EntityUID  string    // UID of the affected entity
	Meta       *string   // optional JSON with extra context (old/new values, comments…)
	OccurredAt time.Time // when the action actually happened
	CreatedAt  time.Time // when the log row was inserted
}
