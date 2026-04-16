CREATE TABLE attendance_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE CASCADE,
    device_uid TEXT NOT NULL,
    device_user_id TEXT NOT NULL,
    punched_at TEXT NOT NULL,
    punch_type TEXT NOT NULL CHECK (punch_type IN ('check_in', 'check_out', 'break_start', 'break_end', 'unknown')),
    raw_payload TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    UNIQUE(device_uid, device_user_id, punched_at, punch_type)
);

CREATE INDEX idx_attendance_records_employee_date ON attendance_records(employee_uid, punched_at);
CREATE INDEX idx_attendance_records_device_date ON attendance_records(device_uid, punched_at);

CREATE TABLE work_hours_configs (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    work_day_start TEXT NOT NULL,
    work_day_end TEXT NOT NULL,
    late_grace_minutes INTEGER NOT NULL,
    early_grace_minutes INTEGER NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT OR IGNORE INTO work_hours_configs (id, work_day_start, work_day_end, late_grace_minutes, early_grace_minutes)
VALUES (1, '09:00', '17:00', 15, 15);
