CREATE TABLE leave_records (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    leave_type_id INTEGER NOT NULL REFERENCES leave_types(id) ON DELETE RESTRICT,
    start_date TEXT NOT NULL,
    end_date TEXT NOT NULL,
    days INTEGER NOT NULL,
    recorded_at TEXT NOT NULL,
    recorded_by INTEGER REFERENCES employees(id),
    notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_leave_records_employee_id ON leave_records(employee_id);
CREATE INDEX idx_leave_records_dates ON leave_records(start_date, end_date);
CREATE INDEX idx_leave_records_employee_dates ON leave_records(employee_id, start_date, end_date);
