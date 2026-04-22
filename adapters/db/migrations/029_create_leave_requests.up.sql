CREATE TABLE leave_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE RESTRICT,
    leave_type_uid TEXT NOT NULL REFERENCES leave_types(uid) ON DELETE RESTRICT,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    days INTEGER NOT NULL CHECK (days > 0),
    notes TEXT,
    study_destination TEXT,
    assignment TEXT,
    assignment_country TEXT,
    spouse_work_country TEXT,
    submitted_at TEXT NOT NULL DEFAULT (datetime('now')),
    decided_at TEXT,
    approval_request_uid TEXT NOT NULL REFERENCES approval_requests(uid) ON DELETE RESTRICT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),

    CONSTRAINT valid_date_range CHECK (end_date >= start_date)
);

CREATE INDEX idx_leave_requests_employee ON leave_requests(employee_uid);
CREATE INDEX idx_leave_requests_leave_type ON leave_requests(leave_type_uid);
CREATE INDEX idx_leave_requests_dates ON leave_requests(start_date, end_date);
CREATE INDEX idx_leave_requests_approval ON leave_requests(approval_request_uid);
