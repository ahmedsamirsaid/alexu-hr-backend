package db

import (
	"context"
	"database/sql"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

// AuditLogRepository is the PostgreSQL-backed implementation of ports.AuditLogRepository.
type AuditLogRepository struct{}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{}
}

// Create inserts one audit log row.
func (r *AuditLogRepository) Create(ctx context.Context, q ports.Querier, log *domain.AuditLog) error {
	query := `
		INSERT INTO audit_logs (uid, actor_uid, action, entity_type, entity_uid, meta, occurred_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	log.CreatedAt = now
	if log.OccurredAt.IsZero() {
		log.OccurredAt = now
	}

	err := q.QueryRowContext(ctx, query,
		log.UID, log.ActorUID, log.Action,
		log.EntityType, log.EntityUID, log.Meta,
		log.OccurredAt, log.CreatedAt,
	).Scan(&log.ID)
	if err != nil {
		slog.Error("audit_log_repository.Create.exec_query",
			"error", err,
			"uid", log.UID,
			"action", log.Action,
			"entity_type", log.EntityType,
		)
		return err
	}

	return nil
}

// ListByEntity returns all audit entries for a specific entity, newest first.
func (r *AuditLogRepository) ListByEntity(ctx context.Context, q ports.Querier, entityType, entityUID string) ([]*domain.AuditLog, error) {
	query := `
		SELECT id, uid, actor_uid, action, entity_type, entity_uid, meta, occurred_at, created_at
		FROM audit_logs
		WHERE entity_type = $1 AND entity_uid = $2
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
		WHERE actor_uid = $1
		ORDER BY occurred_at DESC
		LIMIT $2 OFFSET $3`

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

	// Base query - use JOIN if actor name filtering is needed
	var query string
	if filters.ActorName != "" {
		query = `
			SELECT al.id, al.uid, al.actor_uid, al.action, al.entity_type, al.entity_uid, al.meta, al.occurred_at, al.created_at
			FROM audit_logs al
			LEFT JOIN employees e ON al.actor_uid = e.uid
			WHERE 1=1`
	} else {
		query = `
			SELECT id, uid, actor_uid, action, entity_type, entity_uid, meta, occurred_at, created_at
			FROM audit_logs
			WHERE 1=1`
	}

	args := []interface{}{}

	// Apply filters
	if filters.EntityType != "" {
		if filters.ActorName != "" {
			query += " AND al.entity_type = " + nextPlaceholder(args)
		} else {
			query += " AND entity_type = " + nextPlaceholder(args)
		}
		args = append(args, filters.EntityType)
	}

	if len(filters.EntityTypes) > 0 {
		startIdx := len(args) + 1
		if filters.ActorName != "" {
			query += " AND al.entity_type IN (" + buildPlaceholders(len(filters.EntityTypes), startIdx) + ")"
		} else {
			query += " AND entity_type IN (" + buildPlaceholders(len(filters.EntityTypes), startIdx) + ")"
		}
		for _, et := range filters.EntityTypes {
			args = append(args, et)
		}
	}

	if filters.ActorUID != "" {
		if filters.ActorName != "" {
			query += " AND al.actor_uid = " + nextPlaceholder(args)
		} else {
			query += " AND actor_uid = " + nextPlaceholder(args)
		}
		args = append(args, filters.ActorUID)
	}

	if filters.ActorName != "" {
		query += " AND e.name LIKE " + nextPlaceholder(args)
		args = append(args, "%"+filters.ActorName+"%")
	}

	if len(filters.ActionTypes) > 0 {
		startIdx := len(args) + 1
		if filters.ActorName != "" {
			query += " AND al.action IN (" + buildPlaceholders(len(filters.ActionTypes), startIdx) + ")"
		} else {
			query += " AND action IN (" + buildPlaceholders(len(filters.ActionTypes), startIdx) + ")"
		}
		for _, at := range filters.ActionTypes {
			args = append(args, at)
		}
	}

	if filters.SearchText != "" {
		searchPattern := "%" + filters.SearchText + "%"
		if filters.ActorName != "" {
			query += " AND (al.entity_uid LIKE " + nextPlaceholder(args) + " OR al.action LIKE " + nextPlaceholder(append(args, nil)) + " OR al.meta LIKE " + nextPlaceholder(append(args, nil, nil)) + ")"
		} else {
			query += " AND (entity_uid LIKE " + nextPlaceholder(args) + " OR action LIKE " + nextPlaceholder(append(args, nil)) + " OR meta LIKE " + nextPlaceholder(append(args, nil, nil)) + ")"
		}
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if filters.StartDate != nil {
		if filters.ActorName != "" {
			query += " AND al.occurred_at >= " + nextPlaceholder(args)
		} else {
			query += " AND occurred_at >= " + nextPlaceholder(args)
		}
		args = append(args, filters.StartDate.Format(time.RFC3339))
	}

	if filters.EndDate != nil {
		if filters.ActorName != "" {
			query += " AND al.occurred_at <= " + nextPlaceholder(args)
		} else {
			query += " AND occurred_at <= " + nextPlaceholder(args)
		}
		args = append(args, filters.EndDate.Format(time.RFC3339))
	}

	if filters.HideSystemActions {
		// Exclude system actions
		systemActions := []string{"deduct_balance", "annual_reset", "auto_reject", "system_adjustment", "cascade_delete", "auto_create_transaction"}
		startIdx := len(args) + 1
		if filters.ActorName != "" {
			query += " AND al.action NOT IN (" + buildPlaceholders(len(systemActions), startIdx) + ")"
		} else {
			query += " AND action NOT IN (" + buildPlaceholders(len(systemActions), startIdx) + ")"
		}
		for _, sa := range systemActions {
			args = append(args, sa)
		}
	}

	if filters.ActorName != "" {
		query += " ORDER BY al.occurred_at DESC LIMIT " + nextPlaceholder(args) + " OFFSET " + nextPlaceholder(append(args, nil))
	} else {
		query += " ORDER BY occurred_at DESC LIMIT " + nextPlaceholder(args) + " OFFSET " + nextPlaceholder(append(args, nil))
	}
	args = append(args, p.PageSize, p.Offset())

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("audit_log_repository.ListAll.query",
			"error", err,
			"entity_type", filters.EntityType,
			"actor_uid", filters.ActorUID,
			"actor_name", filters.ActorName,
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

// GetDistinctActionTypes returns all unique action types in the audit log.
func (r *AuditLogRepository) GetDistinctActionTypes(ctx context.Context, q ports.Querier) ([]string, error) {
	query := `
		SELECT DISTINCT action
		FROM audit_logs
		ORDER BY action ASC`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("audit_log_repository.GetDistinctActionTypes.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var actionTypes []string
	for rows.Next() {
		var action string
		if err := rows.Scan(&action); err != nil {
			slog.Error("audit_log_repository.GetDistinctActionTypes.scan", "error", err)
			return nil, err
		}
		actionTypes = append(actionTypes, action)
	}

	if err := rows.Err(); err != nil {
		slog.Error("audit_log_repository.GetDistinctActionTypes.rows_err", "error", err)
		return nil, err
	}

	return actionTypes, nil
}

// GetDistinctEntityTypes returns all unique entity types in the audit log.
func (r *AuditLogRepository) GetDistinctEntityTypes(ctx context.Context, q ports.Querier) ([]string, error) {
	query := `
		SELECT DISTINCT entity_type
		FROM audit_logs
		ORDER BY entity_type ASC`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("audit_log_repository.GetDistinctEntityTypes.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var entityTypes []string
	for rows.Next() {
		var entityType string
		if err := rows.Scan(&entityType); err != nil {
			slog.Error("audit_log_repository.GetDistinctEntityTypes.scan", "error", err)
			return nil, err
		}
		entityTypes = append(entityTypes, entityType)
	}

	if err := rows.Err(); err != nil {
		slog.Error("audit_log_repository.GetDistinctEntityTypes.rows_err", "error", err)
		return nil, err
	}

	return entityTypes, nil
}

// CountAll returns the total count of audit logs matching the filters.
func (r *AuditLogRepository) CountAll(ctx context.Context, q ports.Querier, filters ports.AuditLogFilters) (int, error) {
	query := `SELECT COUNT(*) FROM audit_logs WHERE 1=1`
	args := []interface{}{}

	// Apply the same filters as ListAll
	if filters.EntityType != "" {
		query += " AND entity_type = " + nextPlaceholder(args)
		args = append(args, filters.EntityType)
	}

	if len(filters.EntityTypes) > 0 {
		startIdx := len(args) + 1
		query += " AND entity_type IN (" + buildPlaceholders(len(filters.EntityTypes), startIdx) + ")"
		for _, et := range filters.EntityTypes {
			args = append(args, et)
		}
	}

	if filters.ActorUID != "" {
		query += " AND actor_uid = " + nextPlaceholder(args)
		args = append(args, filters.ActorUID)
	}

	if filters.ActorName != "" {
		// JOIN with employees table for actor name filtering
		query = `SELECT COUNT(*) FROM audit_logs al 
			LEFT JOIN employees e ON al.actor_uid = e.uid 
			WHERE 1=1`
		query += " AND e.name LIKE " + nextPlaceholder(args)
		args = append(args, "%"+filters.ActorName+"%")
	}

	if len(filters.ActionTypes) > 0 {
		startIdx := len(args) + 1
		query += " AND action IN (" + buildPlaceholders(len(filters.ActionTypes), startIdx) + ")"
		for _, at := range filters.ActionTypes {
			args = append(args, at)
		}
	}

	if filters.SearchText != "" {
		searchPattern := "%" + filters.SearchText + "%"
		query += " AND (entity_uid LIKE " + nextPlaceholder(args) + " OR action LIKE " + nextPlaceholder(append(args, nil)) + " OR meta LIKE " + nextPlaceholder(append(args, nil, nil)) + ")"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	if filters.StartDate != nil {
		query += " AND occurred_at >= " + nextPlaceholder(args)
		args = append(args, filters.StartDate.Format(time.RFC3339))
	}

	if filters.EndDate != nil {
		query += " AND occurred_at <= " + nextPlaceholder(args)
		args = append(args, filters.EndDate.Format(time.RFC3339))
	}

	if filters.HideSystemActions {
		// Exclude system actions
		systemActions := []string{"deduct_balance", "annual_reset", "auto_reject", "system_adjustment", "cascade_delete", "auto_create_transaction"}
		startIdx := len(args) + 1
		query += " AND action NOT IN (" + buildPlaceholders(len(systemActions), startIdx) + ")"
		for _, sa := range systemActions {
			args = append(args, sa)
		}
	}

	var count int
	err := q.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		slog.Error("audit_log_repository.CountAll.query", "error", err)
		return 0, err
	}

	return count, nil
}
