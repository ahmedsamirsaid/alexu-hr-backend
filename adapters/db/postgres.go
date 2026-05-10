package db

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/banumusa/backend/core/ports"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDB struct {
	db *sql.DB
}

func NewPostgresDB(host, port, user, password, dbname, sslmode string) (*PostgresDB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode)

	slog.Info("postgres.NewPostgresDB", "driver", "pgx", "dsn", dsn)
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		slog.Error("postgres.NewPostgresDB.open", "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(time.Hour)
	db.SetConnMaxIdleTime(30 * time.Minute)

	if err := db.PingContext(context.Background()); err != nil {
		slog.Error("postgres.NewPostgresDB.ping", "error", err)
		db.Close()
		return nil, err
	}

	return &PostgresDB{db: db}, nil
}

func (p *PostgresDB) Close() error {
	return p.db.Close()
}

func (p *PostgresDB) DB() *sql.DB {
	return p.db
}

func (p *PostgresDB) BeginTx(ctx context.Context, opts *sql.TxOptions) (ports.Tx, error) {
	tx, err := p.db.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &PostgresTx{tx: tx}, nil
}

func (p *PostgresDB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	slog.Debug("postgres.QueryContext", "query_len", len(query), "args", args)
	rows, err := p.db.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("postgres.QueryContext.error", "error", err, "query", query)
	}
	return rows, err
}

func (p *PostgresDB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return p.db.QueryRowContext(ctx, query, args...)
}

func (p *PostgresDB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return p.db.ExecContext(ctx, query, args...)
}

type PostgresTx struct {
	tx *sql.Tx
}

func (t *PostgresTx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.tx.QueryContext(ctx, query, args...)
}

func (t *PostgresTx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.tx.QueryRowContext(ctx, query, args...)
}

func (t *PostgresTx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.tx.ExecContext(ctx, query, args...)
}

func (t *PostgresTx) Commit() error {
	return t.tx.Commit()
}

func (t *PostgresTx) Rollback() error {
	return t.tx.Rollback()
}