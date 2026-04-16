CREATE TABLE attendance_exceptions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE CASCADE,
    attendance_date TEXT NOT NULL,
    exception_type TEXT NOT NULL CHECK (exception_type IN ('missed_punch_in', 'missed_punch_out', 'late_arrival', 'early_departure', 'absence')),
    check_in TEXT,
    check_out TEXT,
    grace_minutes INTEGER,
    minutes_delta INTEGER,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    UNIQUE(employee_uid, attendance_date, exception_type)
);

CREATE INDEX idx_attendance_exceptions_employee_date ON attendance_exceptions(employee_uid, attendance_date);
CREATE INDEX idx_attendance_exceptions_type_date ON attendance_exceptions(exception_type, attendance_date);
