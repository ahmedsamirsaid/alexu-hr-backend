DROP INDEX IF EXISTS idx_leave_records_request;

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
CREATE TABLE leave_records_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    leave_type_id INTEGER NOT NULL REFERENCES leave_types(id) ON DELETE RESTRICT,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    days INTEGER NOT NULL,
    recorded_at TEXT NOT NULL,
    recorded_by INTEGER REFERENCES employees(id),
    notes TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT INTO leave_records_new (id, uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, created_at, updated_at)
SELECT id, uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, created_at, updated_at
FROM leave_records;

DROP TABLE leave_records;
ALTER TABLE leave_records_new RENAME TO leave_records;

CREATE INDEX idx_leave_records_employee_id ON leave_records(employee_id);
CREATE INDEX idx_leave_records_dates ON leave_records(start_date, end_date);
CREATE INDEX idx_leave_records_employee_dates ON leave_records(employee_id, start_date, end_date);
