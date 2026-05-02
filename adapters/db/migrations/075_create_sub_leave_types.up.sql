CREATE TABLE sub_leave_types (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    leave_type_uid TEXT NOT NULL REFERENCES leave_types(uid) ON DELETE RESTRICT,
    name_en TEXT NOT NULL,
    name_ar TEXT NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX idx_sub_leave_types_leave_type_uid ON sub_leave_types(leave_type_uid);
