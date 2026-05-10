package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/banumusa/backend/core/domain"
	"github.com/banumusa/backend/core/ports"
)

type AttendanceDeviceRepository struct{}

func NewAttendanceDeviceRepository() *AttendanceDeviceRepository {
	return &AttendanceDeviceRepository{}
}

func (r *AttendanceDeviceRepository) Create(ctx context.Context, q ports.Querier, device *domain.AttendanceDevice) error {
	query := `
		INSERT INTO attendance_devices (uid, ip, port, name, location, serial_number, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`

	now := time.Now()
	device.CreatedAt = now
	device.UpdatedAt = now

	err := q.QueryRowContext(ctx, query,
		device.UID,
		device.IP,
		device.Port,
		device.Name,
		device.Location,
		device.SerialNumber,
		string(device.Status),
		device.CreatedAt,
		device.UpdatedAt,
	).Scan(&device.ID)
	if err != nil {
		slog.Error("attendance_device_repository.Create.exec_query", "error", err, "uid", device.UID, "serial_number", device.SerialNumber)
		return err
	}

	return nil
}

func (r *AttendanceDeviceRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.AttendanceDevice, error) {
	query := `
		SELECT id, uid, ip, port, name, location, serial_number, status, created_at, updated_at
		FROM attendance_devices
		WHERE uid = $1`
	return r.scanDevice(q.QueryRowContext(ctx, query, uid))
}

func (r *AttendanceDeviceRepository) GetBySerialNumber(ctx context.Context, q ports.Querier, serialNumber string) (*domain.AttendanceDevice, error) {
	query := `
		SELECT id, uid, ip, port, name, location, serial_number, status, created_at, updated_at
		FROM attendance_devices
		WHERE serial_number = $1`
	return r.scanDevice(q.QueryRowContext(ctx, query, serialNumber))
}

func (r *AttendanceDeviceRepository) GetByAddress(ctx context.Context, q ports.Querier, ip string, port int) (*domain.AttendanceDevice, error) {
	query := `
		SELECT id, uid, ip, port, name, location, serial_number, status, created_at, updated_at
		FROM attendance_devices
		WHERE ip = $1 AND port = $2`
	return r.scanDevice(q.QueryRowContext(ctx, query, ip, port))
}

func (r *AttendanceDeviceRepository) Update(ctx context.Context, q ports.Querier, device *domain.AttendanceDevice) error {
	query := `
		UPDATE attendance_devices
		SET ip = $1, port = $2, name = $3, location = $4, status = $5, updated_at = $6
		WHERE uid = $7`

	device.UpdatedAt = time.Now()

	_, err := q.ExecContext(ctx, query,
		device.IP,
		device.Port,
		device.Name,
		device.Location,
		string(device.Status),
		device.UpdatedAt,
		device.UID,
	)
	if err != nil {
		slog.Error("attendance_device_repository.Update.exec", "error", err, "uid", device.UID)
	}

	return err
}

func (r *AttendanceDeviceRepository) List(ctx context.Context, q ports.Querier, filter ports.AttendanceDeviceListFilter, limit, offset int) ([]*domain.AttendanceDevice, error) {
	whereClause, whereArgs := buildAttendanceDeviceWhereClause(filter)

	query := fmt.Sprintf(`
		SELECT id, uid, ip, port, name, location, serial_number, status, created_at, updated_at
		FROM attendance_devices
		%s
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`, whereClause)

	args := append(whereArgs, limit, offset)

	rows, err := q.QueryContext(ctx, query, args...)
	if err != nil {
		slog.Error("attendance_device_repository.List.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var devices []*domain.AttendanceDevice
	for rows.Next() {
		dev, err := r.scanDeviceRow(rows)
		if err != nil {
			slog.Error("attendance_device_repository.List.scan_row", "error", err)
			return nil, err
		}
		devices = append(devices, dev)
	}

	if err := rows.Err(); err != nil {
		slog.Error("attendance_device_repository.List.rows_err", "error", err)
		return nil, err
	}

	return devices, nil
}

func (r *AttendanceDeviceRepository) ListAll(ctx context.Context, q ports.Querier) ([]*domain.AttendanceDevice, error) {
	query := `
		SELECT id, uid, ip, port, name, location, serial_number, status, created_at, updated_at
		FROM attendance_devices
		ORDER BY created_at DESC`

	rows, err := q.QueryContext(ctx, query)
	if err != nil {
		slog.Error("attendance_device_repository.ListAll.query", "error", err)
		return nil, err
	}
	defer rows.Close()

	var devices []*domain.AttendanceDevice
	for rows.Next() {
		dev, err := r.scanDeviceRow(rows)
		if err != nil {
			slog.Error("attendance_device_repository.ListAll.scan_row", "error", err)
			return nil, err
		}
		devices = append(devices, dev)
	}

	if err := rows.Err(); err != nil {
		slog.Error("attendance_device_repository.ListAll.rows_err", "error", err)
		return nil, err
	}

	return devices, nil
}

func (r *AttendanceDeviceRepository) Count(ctx context.Context, q ports.Querier) (int, error) {
	query := `SELECT COUNT(*) FROM attendance_devices`
	var count int
	if err := q.QueryRowContext(ctx, query).Scan(&count); err != nil {
		slog.Error("attendance_device_repository.Count.scan", "error", err)
		return 0, err
	}
	return count, nil
}

func (r *AttendanceDeviceRepository) CountFiltered(ctx context.Context, q ports.Querier, filter ports.AttendanceDeviceListFilter) (int, error) {
	whereClause, whereArgs := buildAttendanceDeviceWhereClause(filter)
	query := fmt.Sprintf(`SELECT COUNT(*) FROM attendance_devices %s`, whereClause)

	var count int
	if err := q.QueryRowContext(ctx, query, whereArgs...).Scan(&count); err != nil {
		slog.Error("attendance_device_repository.CountFiltered.scan", "error", err)
		return 0, err
	}

	return count, nil
}

func (r *AttendanceDeviceRepository) CountByStatus(ctx context.Context, q ports.Querier, status domain.AttendanceDeviceStatus) (int, error) {
	query := `SELECT COUNT(*) FROM attendance_devices WHERE status = $1`
	var count int
	if err := q.QueryRowContext(ctx, query, string(status)).Scan(&count); err != nil {
		slog.Error("attendance_device_repository.CountByStatus.scan", "error", err, "status", status)
		return 0, err
	}
	return count, nil
}

func (r *AttendanceDeviceRepository) UpdateStatus(ctx context.Context, q ports.Querier, uid string, status domain.AttendanceDeviceStatus) error {
	query := `UPDATE attendance_devices SET status = $1, updated_at = $2 WHERE uid = $3`
	_, err := q.ExecContext(ctx, query, string(status), time.Now(), uid)
	if err != nil {
		slog.Error("attendance_device_repository.UpdateStatus.exec", "error", err, "uid", uid)
	}
	return err
}

func (r *AttendanceDeviceRepository) Delete(ctx context.Context, q ports.Querier, uid string) error {
	query := `DELETE FROM attendance_devices WHERE uid = $1`
	_, err := q.ExecContext(ctx, query, uid)
	if err != nil {
		slog.Error("attendance_device_repository.Delete.exec", "error", err, "uid", uid)
	}
	return err
}

func (r *AttendanceDeviceRepository) scanDevice(row *sql.Row) (*domain.AttendanceDevice, error) {
	var dev domain.AttendanceDevice
	var createdAt, updatedAt domain.Time
	var status string

	err := row.Scan(
		&dev.ID,
		&dev.UID,
		&dev.IP,
		&dev.Port,
		&dev.Name,
		&dev.Location,
		&dev.SerialNumber,
		&status,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		slog.Error("attendance_device_repository.scanDevice.scan_row", "error", err)
		return nil, err
	}

	dev.Status = domain.AttendanceDeviceStatus(status)
	dev.CreatedAt = createdAt.Time
	dev.UpdatedAt = updatedAt.Time
	return &dev, nil
}

func (r *AttendanceDeviceRepository) scanDeviceRow(rows *sql.Rows) (*domain.AttendanceDevice, error) {
	var dev domain.AttendanceDevice
	var createdAt, updatedAt domain.Time
	var status string

	err := rows.Scan(
		&dev.ID,
		&dev.UID,
		&dev.IP,
		&dev.Port,
		&dev.Name,
		&dev.Location,
		&dev.SerialNumber,
		&status,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	dev.Status = domain.AttendanceDeviceStatus(status)
	dev.CreatedAt = createdAt.Time
	dev.UpdatedAt = updatedAt.Time
	return &dev, nil
}

func buildAttendanceDeviceWhereClause(filter ports.AttendanceDeviceListFilter) (string, []any) {
	clauses := make([]string, 0, 2)
	args := make([]any, 0, 5)
	param := 1

	if filter.Status != nil {
		clauses = append(clauses, fmt.Sprintf("status = $%d", param))
		args = append(args, string(*filter.Status))
		param++
	}

	if search := strings.TrimSpace(filter.Search); search != "" {
		like := "%" + search + "%"
		clauses = append(clauses, fmt.Sprintf("(name LIKE $%d OR location LIKE $%d OR ip LIKE $%d OR serial_number LIKE $%d OR CAST(port AS TEXT) LIKE $%d)", param, param+1, param+2, param+3, param+4))
		args = append(args, like, like, like, like, like)
		param += 5
	}

	if len(clauses) == 0 {
		return "", nil
	}

	return "WHERE " + strings.Join(clauses, " AND "), args
}
