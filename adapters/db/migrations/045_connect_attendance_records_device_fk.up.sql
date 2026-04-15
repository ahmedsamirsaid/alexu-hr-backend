PRAGMA foreign_keys = OFF;

DROP INDEX IF EXISTS idx_attendance_records_device_date;
DROP INDEX IF EXISTS idx_attendance_records_employee_date;

CREATE TABLE attendance_records_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE CASCADE,
    device_uid TEXT NOT NULL REFERENCES attendance_devices(uid) ON DELETE CASCADE,
    device_user_id TEXT NOT NULL,
    punched_at TEXT NOT NULL,
    punch_type TEXT NOT NULL CHECK (punch_type IN ('check_in', 'check_out', 'break_start', 'break_end', 'unknown')),
    raw_payload TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    UNIQUE(device_uid, device_user_id, punched_at, punch_type)
);

INSERT INTO attendance_records_new (
    id, uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at
)
SELECT
    id, uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at
FROM attendance_records;

DROP TABLE attendance_records;
ALTER TABLE attendance_records_new RENAME TO attendance_records;

CREATE INDEX idx_attendance_records_employee_date ON attendance_records(employee_uid, punched_at);
CREATE INDEX idx_attendance_records_device_date ON attendance_records(device_uid, punched_at);

PRAGMA foreign_keys = ON;
