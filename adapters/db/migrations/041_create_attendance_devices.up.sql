CREATE TABLE attendance_devices (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    ip TEXT NOT NULL,
    port INTEGER NOT NULL,
    name TEXT NOT NULL,
    location TEXT NOT NULL DEFAULT '',
    serial_number TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'offline' CHECK (status IN ('online', 'offline')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_attendance_devices_serial_number ON attendance_devices(serial_number);
CREATE UNIQUE INDEX idx_attendance_devices_ip_port ON attendance_devices(ip, port);
CREATE INDEX idx_attendance_devices_status ON attendance_devices(status);
