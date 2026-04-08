package db

import (
	"context"
	"database/sql"
	"log/slog"

	"github.com/banumusa/backend/core/ports"
	_ "modernc.org/sqlite"
)

type SQLiteDB struct {
	db *sql.DB
}

func NewSQLiteDB(dsn string) (*SQLiteDB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		slog.Error("sqlite.NewSQLiteDB.open_connection", "error", err, "dsn", dsn)
		return nil, err
	}

	// Enable WAL mode for better concurrency (allows concurrent reads while writing)
	_, err = db.Exec("PRAGMA journal_mode = WAL")
	if err != nil {
		slog.Error("sqlite.NewSQLiteDB.enable_wal_mode", "error", err)
		db.Close()
		return nil, err
	}

	// Set busy timeout to 5 seconds (SQLite will retry for this duration before returning SQLITE_BUSY)
	_, err = db.Exec("PRAGMA busy_timeout = 5000")
	if err != nil {
		slog.Error("sqlite.NewSQLiteDB.set_busy_timeout", "error", err)
		db.Close()
		return nil, err
	}

	// Enable foreign keys
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		slog.Error("sqlite.NewSQLiteDB.enable_foreign_keys", "error", err)
		db.Close()
		return nil, err
	}

	return &SQLiteDB{db: db}, nil
}

func (s *SQLiteDB) Close() error {
	return s.db.Close()
}

func (s *SQLiteDB) DB() *sql.DB {
	return s.db
}

func (s *SQLiteDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Tx, error) {
	tx, err := s.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &SQLiteTx{tx: tx}, nil
}

func (s *SQLiteDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, query, args...)
}

func (s *SQLiteDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(ctx, query, args...)
}

func (s *SQLiteDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return s.db.ExecContext(ctx, query, args...)
}

type SQLiteTx struct {
	tx *sql.Tx
}

func (t *SQLiteTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *SQLiteTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *SQLiteTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *SQLiteTx) Commit() error {
	return t.tx.Commit()
}

func (t *SQLiteTx) Rollback() error {
	return t.tx.Rollback()
}
