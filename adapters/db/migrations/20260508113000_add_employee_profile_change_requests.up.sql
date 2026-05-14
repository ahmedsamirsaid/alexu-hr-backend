-- Note: Seed data for permissions, roles, and approval flows
-- is in adapters/db/seed/seed_dev_consolidated.sql

CREATE TABLE employee_profile_change_requests (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE RESTRICT,
    approval_request_uid TEXT UNIQUE NOT NULL REFERENCES approval_requests(uid) ON DELETE RESTRICT,
    submitted_by_employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE RESTRICT,
    current_financial_grade TEXT,
    current_id_card_valid_until TEXT,
    current_marital_status TEXT,
    requested_financial_grade TEXT,
    requested_id_card_valid_until TEXT,
    requested_marital_status TEXT,
    comments TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_employee_profile_change_requests_employee_uid ON employee_profile_change_requests(employee_uid);
CREATE INDEX idx_employee_profile_change_requests_submitted_by ON employee_profile_change_requests(submitted_by_employee_uid);
CREATE INDEX idx_employee_profile_change_requests_created_at ON employee_profile_change_requests(created_at);

-- Note: PostgreSQL doesn't support triggers in the same way as SQLite
-- The single pending request constraint should be enforced at the application level
