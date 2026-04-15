PRAGMA foreign_keys = OFF;

CREATE TABLE attendance_devices_old (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    ip TEXT NOT NULL,
    port INTEGER NOT NULL,
    name TEXT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    serial_number TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'offline' CHECK (status IN ('online', 'offline')),
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT INTO attendance_devices_old (
    id,
    uid,
    ip,
    port,
    name,
    location,
    serial_number,
    status,
    created_at,
    updated_at
)
SELECT
    id,
    uid,
    ip,
    port,
    name,
    location,
    serial_number,
    CASE
        WHEN status = 'deactivated' THEN 'offline'
        ELSE status
    END,
    created_at,
    updated_at
FROM attendance_devices;

DROP TABLE attendance_devices;
ALTER TABLE attendance_devices_old RENAME TO attendance_devices;

CREATE UNIQUE INDEX idx_attendance_devices_serial_number ON attendance_devices(serial_number);
CREATE UNIQUE INDEX idx_attendance_devices_ip_port ON attendance_devices(ip, port);
CREATE INDEX idx_attendance_devices_status ON attendance_devices(status);

PRAGMA foreign_keys = ON;