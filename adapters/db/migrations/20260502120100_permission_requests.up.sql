-- Comprehensive permission requests implementation
-- This consolidates: create_permission_requests, seed_permission_approval_flow, 
-- seed_permission_perms, and resolve_permission_window_at_creation into one clean migration

-- 1. Create permission_requests table with final schema
CREATE TABLE permission_requests (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE RESTRICT,
    type TEXT NOT NULL CHECK (type IN ('morning','personal','official','health_insurance')),
    permission_date TEXT NOT NULL,
    start_time TEXT,
    end_time TEXT,
    reason TEXT,
    submitted_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    decided_at TEXT,
    approval_request_uid TEXT NOT NULL REFERENCES approval_requests(uid) ON DELETE RESTRICT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_permission_requests_employee_date ON permission_requests(employee_uid, permission_date);
CREATE INDEX idx_permission_requests_approval ON permission_requests(approval_request_uid);
CREATE INDEX idx_permission_requests_type_date ON permission_requests(type, permission_date);

-- Note: Seed data for approval flows, permissions, and role assignments
-- is in adapters/db/seed/seed_dev_consolidated.sql
