CREATE TABLE leave_balances (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    employee_id INTEGER NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    leave_type_id INTEGER NOT NULL REFERENCES leave_types(id) ON DELETE RESTRICT,
    year INTEGER NOT NULL,
    total_days INTEGER NOT NULL,
    used_days INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(employee_id, leave_type_id, year)
);

CREATE INDEX idx_leave_balances_employee_year ON leave_balances(employee_id, year);
