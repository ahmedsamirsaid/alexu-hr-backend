package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// AuditLogRepository is the SQLite-backed implementation of ports.AuditLogRepository.
type AuditLogRepository struct{}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{}
}

// Create inserts one audit log row.
func (r *AuditLogRepository) Create(ctx context.Context, q ports.Querier, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (uid, actor_uid, action, entity_type, entity_uid, meta, occurred_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	log.CreatedAt = now
	if log.OccurredAt.IsZero() {
		log.OccurredAt = now
	}

	result, err := q.ExecContext(ctx, query,
		log.UID, log.ActorUID, log.Action,
		log.EntityType, log.EntityUID, log.Meta,
		log.OccurredAt, log.CreatedAt,
	)
	if err != nil {
		slog.Error("audit_log_repository.Create.exec_query",
			"error", err,
			"uid", log.UID,
			"action", log.Action,
			"entity_type", log.EntityType,
		)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("audit_log_repository.Create.last_insert_id", "error", err, "uid", log.UID)
		return err
	}
	log.ID = id
	return nil
}

// ListByEntity returns all audit entries for a specific entity, newest first.
func (r *AuditLogRepository) ListByEntity(ctx context.Context, q ports.Querier, entityType, entityUID string) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, uid, actor_uid, action, entity_type, entity_uid, meta, occurred_at, created_at
		FROM audit_logs
		WHERE entity_type = ? AND entity_uid = ?
		ORDER BY occurred_at DESC`

	rows, err := q.QueryContext(ctx, query, entityType, entityUID)
	if err != nil {
		slog.Error("audit_log_repository.ListByEntity.query",
			"error", err,
			"entity_type", entityType,
			"entity_uid", entityUID,
		)
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

// ListByActor returns paginated audit entries for a specific actor, newest first.
func (r *AuditLogRepository) ListByActor(ctx context.Context, q ports.Querier, actorUID string, p ports.ListParams) ([]*domain.AuditLog, error) {
	p = p.Normalize(20, 100, "occurred_at", ports.SortOrderDesc)

	query := `
		SELECT id, uid, actor_uid, action, entity_type, entity_uid, meta, occurred_at, created_at
		FROM audit_logs
		WHERE actor_uid = ?
		ORDER BY occurred_at DESC
		LIMIT ? OFFSET ?`

	rows, err := q.QueryContext(ctx, query, actorUID, p.PageSize, p.Offset())
	if err != nil {
		slog.Error("audit_log_repository.ListByActor.query", "error", err, "actor_uid", actorUID)
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

// ListAll returns paginated audit entries with optional filters, newest first.
func (r *AuditLogRepository) ListAll(ctx context.Context, q ports.Querier, filters ports.AuditLogFilters, p ports.ListParams) ([]*domain.AuditLog, error) {
	p = p.Normalize(20, 100, "occurred_at", ports.SortOrderDesc)

	query := `
		SELECT id, uid, actor_uid, action, entity_type, entity_uid, meta, occurred_at, created_at
		FROM audit_logs
		WHERE 1=1`

	args := []interface{}{}

	// Apply filters
	if filters.EntityType != "" {
		query += " AND entity_type = ?"
		args = append(args, filters.EntityType)
	}

	if filters.ActorUID != "" {
		query += " AND actor_uid = ?"
		args = append(args, filters.ActorUID)
	}

	if filters.SearchText != "" {
		searchPattern := "%" + filters.SearchText + "%"
		query += " AND (entity_uid LIKE ? OR action LIKE ? OR meta LIKE ?)"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	query += " ORDER BY occurred_at DESC LIMIT ? OFFSET ?"
	args = append(args, p.PageSize, p.Offset())

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("audit_log_repository.ListAll.query",
			"error", err,
			"entity_type", filters.EntityType,
			"actor_uid", filters.ActorUID,
			"search_text", filters.SearchText,
		)
		return nil, err
	}
	defer rows.Close()
	return r.scanRows(rows)
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func (r *AuditLogRepository) scanRows(rows *sql.Rows) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	for rows.Next() {
		l, err := r.scanRow(rows)
		if err != nil {
			slog.Error("audit_log_repository.scanRows.scan", "error", err)
			return nil, err
		}
		logs = append(logs, l)
	}
	if err := rows.Err(); err != nil {
		slog.Error("audit_log_repository.scanRows.rows_err", "error", err)
		return nil, err
	}
	return logs, nil
}

func (r *AuditLogRepository) scanRow(rows *sql.Rows) (*domain.AuditLog, error) {
	var l domain.AuditLog
	var occurredAt, createdAt domain.Time
	err := rows.Scan(
		&l.ID, &l.UID, &l.ActorUID, &l.Action,
		&l.EntityType, &l.EntityUID, &l.Meta,
		&occurredAt, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	l.OccurredAt = occurredAt.Time
	l.CreatedAt = createdAt.Time
	return &l, nil
}
