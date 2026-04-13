package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
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
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	device.CreatedAt = now
	device.UpdatedAt = now

	result, err := q.ExecContext(ctx, query,
		device.UID,
		device.IP,
		device.Port,
		device.Name,
		device.Location,
		device.SerialNumber,
		string(device.Status),
		device.CreatedAt,
		device.UpdatedAt,
	)
	if err != nil {
		slog.Error("attendance_device_repository.Create.exec_query", "error", err, "uid", device.UID, "serial_number", device.SerialNumber)
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		slog.Error("attendance_device_repository.Create.last_insert_id", "error", err, "uid", device.UID)
		return err
	}
	device.ID = id
	return nil
}

func (r *AttendanceDeviceRepository) GetByUID(ctx context.Context, q ports.Querier, uid string) (*domain.AttendanceDevice, error) {
	query := `
		SELECT id, uid, ip, port, name, location, serial_number, status, created_at, updated_at
		FROM attendance_devices
		WHERE uid = ?`
	return r.scanDevice(q.QueryRowContext(ctx, query, uid))
}

func (r *AttendanceDeviceRepository) GetBySerialNumber(ctx context.Context, q ports.Querier, serialNumber string) (*domain.AttendanceDevice, error) {
	query := `
		SELECT id, uid, ip, port, name, location, serial_number, status, created_at, updated_at
		FROM attendance_devices
		WHERE serial_number = ?`
	return r.scanDevice(q.QueryRowContext(ctx, query, serialNumber))
}

func (r *AttendanceDeviceRepository) GetByAddress(ctx context.Context, q ports.Querier, ip string, port int) (*domain.AttendanceDevice, error) {
	query := `
		SELECT id, uid, ip, port, name, location, serial_number, status, created_at, updated_at
		FROM attendance_devices
		WHERE ip = ? AND port = ?`
	return r.scanDevice(q.QueryRowContext(ctx, query, ip, port))
}

func (r *AttendanceDeviceRepository) List(ctx context.Context, q ports.Querier, limit, offset int) ([]*domain.AttendanceDevice, error) {
	query := `
		SELECT id, uid, ip, port, name, location, serial_number, status, created_at, updated_at
		FROM attendance_devices
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := q.QueryContext(ctx, query, limit, offset)
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

func (r *AttendanceDeviceRepository) CountByStatus(ctx context.Context, q ports.Querier, status domain.AttendanceDeviceStatus) (int, error) {
	query := `SELECT COUNT(*) FROM attendance_devices WHERE status = ?`
	var count int
	if err := q.QueryRowContext(ctx, query, string(status)).Scan(&count); err != nil {
		slog.Error("attendance_device_repository.CountByStatus.scan", "error", err, "status", status)
		return 0, err
	}
	return count, nil
}

func (r *AttendanceDeviceRepository) UpdateStatus(ctx context.Context, q ports.Querier, uid string, status domain.AttendanceDeviceStatus) error {
	query := `UPDATE attendance_devices SET status = ?, updated_at = ? WHERE uid = ?`
	_, err := q.ExecContext(ctx, query, string(status), time.Now(), uid)
	if err != nil {
		slog.Error("attendance_device_repository.UpdateStatus.exec", "error", err, "uid", uid)
	}
	return err
}

func (r *AttendanceDeviceRepository) Delete(ctx context.Context, q ports.Querier, uid string) error {
	query := `DELETE FROM attendance_devices WHERE uid = ?`
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
