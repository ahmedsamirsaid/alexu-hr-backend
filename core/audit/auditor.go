// Package audit provides AOP-style audit logging for Go — as simple as Spring Boot's @Audit.
//
// # One-line usage inside any use case
//
//	// With a known actor UID:
//	defer uc.auditor.Actor(actorUID).Did(audit.ActionApprove).On(audit.EntityApprovalRequest, requestUID).Save(ctx)
//
//	// Actor auto-extracted from context (set by JWT middleware):
//	defer uc.auditor.From(ctx).Did(audit.ActionUpdate).On(audit.EntityAttendanceRecord, uid).Save(ctx)
//
//	// With optional metadata:
//	defer uc.auditor.Actor(actorUID).
//	    Did(audit.ActionUpdate).
//	    On(audit.EntityAttendanceRecord, uid).
//	    WithMeta("punch_type", newPunchType).
//	    WithMeta("punched_at", newTime).
//	    Save(ctx)
//
// Save is non-blocking (fire-and-forget goroutine).
// Use SaveSync when you need guaranteed persistence before returning.
package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// ---------------------------------------------------------------------------
// Action constants  — extend freely
// ---------------------------------------------------------------------------

const (
	ActionCreate        = "create"
	ActionUpdate        = "update"
	ActionDelete        = "delete"
	ActionApprove       = "approve"
	ActionReject        = "reject"
	ActionCancel        = "cancel"
	ActionSubmit        = "submit"
	ActionLogin         = "login"
	ActionLogout        = "logout"
	ActionImport        = "import"
	ActionExport        = "export"
	ActionAssign        = "assign"
	ActionUnassign      = "unassign"
	ActionActivate      = "activate"
	ActionDeactivate    = "deactivate"
	ActionDeductBalance = "deduct_balance"
	ActionRecordLeave   = "record_leave"
	ActionAddStep       = "add_step"
	ActionUpdateStep    = "update_step"
	ActionDeleteStep    = "delete_step"
)

// ---------------------------------------------------------------------------
// Entity type constants — extend freely
// ---------------------------------------------------------------------------

const (
	EntityAttendanceRecord = "attendance_record"
	EntityLeaveRequest     = "leave_request"
	EntityLeaveRecord      = "leave_record"
	EntityApprovalRequest  = "approval_request"
	EntityEmployee         = "employee"
	EntityUser             = "user"
	EntityDepartment       = "department"
	EntityRole             = "role"
	EntityApprovalFlow     = "approval_flow"
	EntityApprovalFlowStep = "approval_flow_step"
	EntityLeaveType        = "leave_type"
	EntityShift            = "shift"
	EntityAttendanceDevice = "attendance_device"
	EntityLeaveBalance     = "leave_balance"
	EntityHoliday          = "holiday"
)

// ---------------------------------------------------------------------------
// Context helpers — used by JWT middleware to inject the actor automatically
// ---------------------------------------------------------------------------

type contextKey string

const actorContextKey contextKey = "audit_actor_uid"

// WithActor injects the acting employee UID into the context.
// Call this from your auth middleware after validating the JWT.
func WithActor(ctx context.Context, employeeUID string) context.Context {
	return context.WithValue(ctx, actorContextKey, employeeUID)
}

// ActorFromContext extracts the acting employee UID from context.
// Returns an empty string if none was set.
func ActorFromContext(ctx context.Context) string {
	uid, _ := ctx.Value(actorContextKey).(string)
	return uid
}

// ---------------------------------------------------------------------------
// Auditor — the interface you inject into use cases
// ---------------------------------------------------------------------------

// Auditor is the single dependency you add to a use case to get audit logging.
// Wire it up once in main.go; use it everywhere with one-liner defers.
type Auditor interface {
	// Actor starts a fluent chain using a known employee UID.
	Actor(actorUID string) *Builder

	// From starts a fluent chain extracting the actor from context
	// (populated by the JWT middleware via WithActor).
	From(ctx context.Context) *Builder
}

// ---------------------------------------------------------------------------
// Builder — fluent chain (mirrors Spring's AuditEvent builder feel)
// ---------------------------------------------------------------------------

// Builder constructs an audit entry with a fluent API.
// Call Save(ctx) or SaveSync(ctx) at the end.
type Builder struct {
	impl       *auditorImpl
	actorUID   string
	action     string
	entityType string
	entityUID  string
	meta       map[string]any
}

// Did sets the action verb (use the Action* constants or any custom string).
func (b *Builder) Did(action string) *Builder {
	b.action = action
	return b
}

// On sets the entity type and UID that was affected.
func (b *Builder) On(entityType, entityUID string) *Builder {
	b.entityType = entityType
	b.entityUID = entityUID
	return b
}

// WithMeta attaches a single key-value pair of extra context.
// Call multiple times to add several fields.
func (b *Builder) WithMeta(key string, value any) *Builder {
	if b.meta == nil {
		b.meta = make(map[string]any)
	}
	b.meta[key] = value
	return b
}

// WithDetails replaces all metadata at once from a map.
func (b *Builder) WithDetails(details map[string]any) *Builder {
	b.meta = details
	return b
}

// Save persists the audit entry asynchronously (non-blocking fire-and-forget).
// This is the recommended default — safe to use with defer, zero latency impact.
// Failures are logged via slog but never returned.
func (b *Builder) Save(ctx context.Context) {
	// Use a background context so the goroutine outlives the HTTP request.
	go b.impl.persist(context.Background(), b)
}

// SaveSync persists the audit entry synchronously.
// Use this when you need the log row committed before the function returns.
func (b *Builder) SaveSync(ctx context.Context) {
	b.impl.persist(ctx, b)
}

// ---------------------------------------------------------------------------
// Real implementation
// ---------------------------------------------------------------------------

type auditorImpl struct {
	db   ports.DB
	repo ports.AuditLogRepository
}

// NewAuditor creates a production Auditor backed by the given repository.
func NewAuditor(db ports.DB, repo ports.AuditLogRepository) Auditor {
	return &auditorImpl{db: db, repo: repo}
}

func (a *auditorImpl) Actor(actorUID string) *Builder {
	return &Builder{impl: a, actorUID: actorUID}
}

func (a *auditorImpl) From(ctx context.Context) *Builder {
	return &Builder{impl: a, actorUID: ActorFromContext(ctx)}
}

func (a *auditorImpl) persist(ctx context.Context, b *Builder) {
	// Silently skip if db/repo were never wired (e.g. NoopAuditor path).
	if a.db == nil || a.repo == nil {
		return
	}

	if shouldSkipAuditLog(b.action, b.actorUID) {
		return
	}

	var metaJSON *string
	if len(b.meta) > 0 {
		raw, err := json.Marshal(b.meta)
		if err == nil {
			s := string(raw)
			metaJSON = &s
		}
	}

	entry := &domain.AuditLog{
		UID:        domain.GenerateUID("aud"),
		ActorUID:   b.actorUID,
		Action:     b.action,
		EntityType: b.entityType,
		EntityUID:  b.entityUID,
		Meta:       metaJSON,
		OccurredAt: time.Now(),
	}

	if err := a.repo.Create(ctx, a.db, entry); err != nil {
		slog.Error("auditor.persist",
			"error", err,
			"action", b.action,
			"entity_type", b.entityType,
			"entity_uid", b.entityUID,
			"actor_uid", b.actorUID,
		)
	}
}

func shouldSkipAuditLog(action, actorUID string) bool {
	if isSystemAction(action) {
		return true
	}

	if actorUID == "system" || strings.HasPrefix(actorUID, "system_") {
		return true
	}

	return false
}

func isSystemAction(action string) bool {
	systemActions := map[string]bool{
		"deduct_balance":          true,
		"annual_reset":            true,
		"auto_reject":             true,
		"system_adjustment":       true,
		"cascade_delete":          true,
		"auto_create_transaction": true,
	}

	return systemActions[action]
}

// ---------------------------------------------------------------------------
// NoopAuditor — use in tests to skip DB writes
// ---------------------------------------------------------------------------

// Noop returns an Auditor that silently discards every audit entry.
// Use it in unit tests so you don't need a DB:
//
//	uc := NewMyUseCase(..., audit.Noop())
func Noop() Auditor { return &noopAuditor{} }

type noopAuditor struct{}

func (n *noopAuditor) Actor(_ string) *Builder {
	return &Builder{impl: &auditorImpl{}}
}
func (n *noopAuditor) From(_ context.Context) *Builder {
	return &Builder{impl: &auditorImpl{}}
}
