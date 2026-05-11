INSERT OR IGNORE INTO permissions (uid, code, description) VALUES
    ('perm_employee_profile_changes_read', 'employee-profile-changes:read', 'View employee profile change requests'),
    ('perm_employee_profile_changes_write', 'employee-profile-changes:write', 'Submit employee profile change requests'),
    ('perm_employee_profile_changes_approve', 'employee-profile-changes:approve', 'Approve or reject employee profile change requests');

INSERT OR IGNORE INTO roles (uid, name, description, scope_type, is_system, created_at, updated_at)
VALUES (
    'role_information_center',
    'Information Center',
    'Information Center with final approval access for employee profile changes',
    'global',
    0,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code IN (
    'employees:read',
    'employees:write',
    'employee-profile-changes:read',
    'employee-profile-changes:approve'
)
WHERE r.uid = 'role_information_center';

INSERT OR IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code = 'employee-profile-changes:write'
WHERE r.uid IN ('role_university_human_resources', 'role_hr_manager');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code IN ('employee-profile-changes:read', 'employee-profile-changes:write')
WHERE r.uid = 'role_department_manager';

INSERT OR IGNORE INTO approval_flows (uid, code, name_en, name_ar, description, is_active, created_at, updated_at)
VALUES (
    'apf_employee_profile_change',
    'employee_profile_change',
    'Employee Profile Change Approval',
    'اعتماد تعديل بيانات الموظف',
    'Approval flow for HR-submitted employee profile changes',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid, created_at, updated_at)
VALUES (
    'afs_employee_profile_change_step1',
    'apf_employee_profile_change',
    1,
    'role_information_center',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

CREATE TABLE employee_profile_change_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX idx_employee_profile_change_requests_employee_uid ON employee_profile_change_requests(employee_uid);
CREATE INDEX idx_employee_profile_change_requests_submitted_by ON employee_profile_change_requests(submitted_by_employee_uid);
CREATE INDEX idx_employee_profile_change_requests_created_at ON employee_profile_change_requests(created_at);

CREATE TRIGGER trg_employee_profile_change_requests_single_pending
BEFORE INSERT ON employee_profile_change_requests
FOR EACH ROW
WHEN EXISTS (
    SELECT 1
    FROM employee_profile_change_requests epcr
    JOIN approval_requests ar ON ar.uid = epcr.approval_request_uid
    WHERE epcr.employee_uid = NEW.employee_uid
      AND ar.status = 'pending'
)
BEGIN
    SELECT RAISE(ABORT, 'pending employee profile change request already exists');
END;
