-- Comprehensive permission requests implementation
-- This consolidates: create_permission_requests, seed_permission_approval_flow, 
-- seed_permission_perms, and resolve_permission_window_at_creation into one clean migration

-- 1. Create permission_requests table with final schema
CREATE TABLE permission_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    employee_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE RESTRICT,
    type TEXT NOT NULL CHECK (type IN ('morning','personal','official','health_insurance')),
    permission_date TEXT NOT NULL,
    start_time TEXT,
    end_time TEXT,
    reason TEXT,
    submitted_at TEXT NOT NULL DEFAULT (datetime('now')),
    decided_at TEXT,
    approval_request_uid TEXT NOT NULL REFERENCES approval_requests(uid) ON DELETE RESTRICT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX idx_permission_requests_employee_date ON permission_requests(employee_uid, permission_date);
CREATE INDEX idx_permission_requests_approval ON permission_requests(approval_request_uid);
CREATE INDEX idx_permission_requests_type_date ON permission_requests(type, permission_date);

-- 2. Create permission approval flow
INSERT OR IGNORE INTO approval_flows (uid, code, name_en, name_ar, description, is_active)
VALUES (
    'apf_permission_default',
    'permission',
    'Permission Approval',
    'اعتماد طلب الإذن',
    'Default approval flow for permission requests',
    1
);

-- Step 1: department manager
INSERT OR IGNORE INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
SELECT 'afs_permission_step1_dm', 'apf_permission_default', 1, 'role_department_manager'
WHERE EXISTS (SELECT 1 FROM approval_flows WHERE uid = 'apf_permission_default')
  AND EXISTS (SELECT 1 FROM roles WHERE uid = 'role_department_manager');

-- Step 2: HR escalation (when employee has no department, or requester is the dept manager)
INSERT OR IGNORE INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
SELECT 'afs_permission_step2_uhr', 'apf_permission_default', 2, 'role_university_human_resources'
WHERE EXISTS (SELECT 1 FROM approval_flows WHERE uid = 'apf_permission_default')
  AND EXISTS (SELECT 1 FROM roles WHERE uid = 'role_university_human_resources');

-- 3. Create permissions and assign to roles
INSERT OR IGNORE INTO permissions (uid, code, description) VALUES
    ('perm_permission_request', 'permission:request', 'Submit permission/excuse requests'),
    ('perm_permission_approve', 'permission:approve', 'Approve/reject permission requests'),
    ('perm_permission_read',    'permission:read',    'View all permission requests across the org');

-- Employee role: can submit / cancel / view own
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.code = 'permission:request'
WHERE r.uid = 'role_employee';

-- Department Manager role: submit own + approve team
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.code IN ('permission:request', 'permission:approve')
WHERE r.uid = 'role_department_manager';

-- Dean: same as department manager (already on the leave approval chain)
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.code IN ('permission:request', 'permission:approve')
WHERE r.uid = 'role_dean';

-- University HR: full access (acts as escalation step)
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.code IN ('permission:request', 'permission:approve', 'permission:read')
WHERE r.uid = 'role_university_human_resources';

-- Admin role: all three
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r JOIN permissions p ON p.code IN ('permission:request', 'permission:approve', 'permission:read')
WHERE r.name = 'Admin';
