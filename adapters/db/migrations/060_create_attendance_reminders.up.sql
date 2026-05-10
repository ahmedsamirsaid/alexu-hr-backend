CREATE TABLE attendance_reminders (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE CASCADE,
    attendance_date TEXT NOT NULL,
    reminder_type TEXT NOT NULL CHECK (reminder_type IN ('missing_check_out')),
    sent_at TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_uid, attendance_date, reminder_type)
);

CREATE INDEX idx_attendance_reminders_employee_date ON attendance_reminders(employee_uid, attendance_date);
