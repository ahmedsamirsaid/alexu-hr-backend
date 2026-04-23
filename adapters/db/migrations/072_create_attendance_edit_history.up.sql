CREATE TABLE attendance_edit_history (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    attendance_record_uid TEXT NOT NULL REFERENCES attendance_records(uid) ON DELETE CASCADE,
    field_changed TEXT NOT NULL,
    old_value TEXT,
    new_value TEXT,
    reason TEXT,
    edited_by_uid TEXT NOT NULL REFERENCES users(uid),
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_attendance_edit_history_record_uid
    ON attendance_edit_history(attendance_record_uid, created_at);
