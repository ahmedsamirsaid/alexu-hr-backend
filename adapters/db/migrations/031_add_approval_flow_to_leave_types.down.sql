DROP INDEX IF EXISTS idx_leave_types_approval_flow;

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
CREATE TABLE leave_types_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    code TEXT UNIQUE NOT NULL,
    name_en TEXT NOT NULL,
    name_ar TEXT NOT NULL,
    default_balance INTEGER NOT NULL,
    max_consecutive INTEGER,
    recording_deadline_days INTEGER,
    advance_notice_days INTEGER,
    is_active INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT INTO leave_types_new (id, uid, code, name_en, name_ar, default_balance, max_consecutive, recording_deadline_days, advance_notice_days, is_active, created_at, updated_at)
SELECT id, uid, code, name_en, name_ar, default_balance, max_consecutive, recording_deadline_days, advance_notice_days, is_active, created_at, updated_at
FROM leave_types;

DROP TABLE leave_types;
ALTER TABLE leave_types_new RENAME TO leave_types;
