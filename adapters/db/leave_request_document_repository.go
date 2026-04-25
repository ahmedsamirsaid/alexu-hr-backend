package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type LeaveRequestDocumentRepository struct{}

func NewLeaveRequestDocumentRepository() *LeaveRequestDocumentRepository {
	return &LeaveRequestDocumentRepository{}
}

func (r *LeaveRequestDocumentRepository) Create(ctx context.Context, q ports.Querier, doc *domain.LeaveRequestDocument) error {
	query := `
		INSERT INTO leave_request_documents (leave_request_uid, file_name, object_key)
		VALUES (?, ?, ?)`

	_, err := q.ExecContext(ctx, query, doc.LeaveRequestUID, doc.FileName, doc.ObjectKey)
	if err != nil {
		slog.Error("leave_request_document_repository.Create.exec_query", "error", err, "leave_request_uid", doc.LeaveRequestUID, "object_key", doc.ObjectKey)
	}
	return err
}

func (r *LeaveRequestDocumentRepository) ListByLeaveRequestUID(ctx context.Context, q ports.Querier, leaveRequestUID string) ([]*domain.LeaveRequestDocument, error) {
	query := `
		SELECT leave_request_uid, file_name, object_key
		FROM leave_request_documents
		WHERE leave_request_uid = ?
		ORDER BY rowid ASC`

	rows, err := q.QueryContext(ctx, query, leaveRequestUID)
	if err != nil {
		slog.Error("leave_request_document_repository.ListByLeaveRequestUID.query", "error", err, "leave_request_uid", leaveRequestUID)
		return nil, err
	}
	defer rows.Close()

	var docs []*domain.LeaveRequestDocument
	for rows.Next() {
		doc := &domain.LeaveRequestDocument{}
		if err := rows.Scan(&doc.LeaveRequestUID, &doc.FileName, &doc.ObjectKey); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return nil, nil
			}
			slog.Error("leave_request_document_repository.ListByLeaveRequestUID.scan", "error", err, "leave_request_uid", leaveRequestUID)
			return nil, err
		}
		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		slog.Error("leave_request_document_repository.ListByLeaveRequestUID.rows", "error", err, "leave_request_uid", leaveRequestUID)
		return nil, err
	}

	return docs, nil
}

func (r *LeaveRequestDocumentRepository) GetByLeaveRequestUIDAndFileName(ctx context.Context, q ports.Querier, leaveRequestUID, fileName string) (*domain.LeaveRequestDocument, error) {
	query := `
		SELECT leave_request_uid, file_name, object_key
		FROM leave_request_documents
		WHERE leave_request_uid = ? AND file_name = ?`

	row := q.QueryRowContext(ctx, query, leaveRequestUID, fileName)
	doc := &domain.LeaveRequestDocument{}
	if err := row.Scan(&doc.LeaveRequestUID, &doc.FileName, &doc.ObjectKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("leave_request_document_repository.GetByLeaveRequestUIDAndFileName.scan", "error", err, "leave_request_uid", leaveRequestUID, "file_name", fileName)
		return nil, err
	}

	return doc, nil
}

func (r *LeaveRequestDocumentRepository) UpdateObjectKey(ctx context.Context, q ports.Querier, leaveRequestUID, fileName, objectKey string) error {
	query := `
		UPDATE leave_request_documents
		SET object_key = ?
		WHERE leave_request_uid = ? AND file_name = ?`

	_, err := q.ExecContext(ctx, query, objectKey, leaveRequestUID, fileName)
	if err != nil {
		slog.Error("leave_request_document_repository.UpdateObjectKey.exec_query", "error", err, "leave_request_uid", leaveRequestUID, "file_name", fileName)
	}
	return err
}
