CREATE TABLE attendance_reminders_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE CASCADE,
    attendance_date TEXT NOT NULL,
    reminder_type TEXT NOT NULL CHECK (reminder_type IN ('missing_check_in', 'missing_check_out')),
    sent_at TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    UNIQUE(employee_uid, attendance_date, reminder_type)
);

INSERT INTO attendance_reminders_new (
    id, uid, employee_uid, attendance_date, reminder_type, sent_at, created_at, updated_at
)
SELECT
    id, uid, employee_uid, attendance_date, reminder_type, sent_at, created_at, updated_at
FROM attendance_reminders;

DROP TABLE attendance_reminders;

ALTER TABLE attendance_reminders_new RENAME TO attendance_reminders;

CREATE INDEX idx_attendance_reminders_employee_date ON attendance_reminders(employee_uid, attendance_date);
