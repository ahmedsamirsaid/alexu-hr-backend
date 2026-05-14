-- ============================================================================
-- Development Seed Data for Banu Musa ERP
-- ============================================================================
-- This file contains ALL seed data consolidated from every migration that
-- contains INSERT / UPDATE / DELETE statements (migrations 009 → latest).
-- DO NOT run this in production!
--
-- To enable: Set BANU_MUSA_SEED_DEV=true in your environment
-- Sources: migrations copy / 009 through 20260502120100 (all DML files)
-- ============================================================================

-- ============================================================================
-- Weekend Configuration
-- (009_seed_data)
-- ============================================================================
INSERT INTO weekend_config (uid, day_of_week) VALUES
    ('wknd_00000000000000000000000000000001', 5),  -- Friday
    ('wknd_00000000000000000000000000000002', 6)   -- Saturday
ON CONFLICT (day_of_week) DO NOTHING;

-- ============================================================================
-- Permissions
-- (018_seed_permissions, 033, 035, 038, 040, 042, 059, 067, 071,
--  20260430005400, 20260502120100)
-- ============================================================================
INSERT INTO permissions (uid, code, description) VALUES
    ('perm_employees_read',           'employees:read',           'View employee list and details'),
    ('perm_employees_write',          'employees:write',          'Create/update employees'),
    ('perm_employees_import',         'employees:import',         'Bulk import employees'),
    ('perm_employees_export',         'employees:export',         'Export employees to Excel/PDF'),
    ('perm_leave_read',               'leave:read',               'View leave records'),
    ('perm_leave_record',             'leave:record',             'Record leave for employees'),
    ('perm_leave_request',            'leave:request',            'Submit leave requests'),
    ('perm_leave_approve',            'leave:approve',            'Approve/reject leave requests'),
    ('perm_leave_types_read',         'leave-types:read',         'View leave types'),
    ('perm_leave_types_write',        'leave-types:write',        'Enable/disable leave types'),
    ('perm_users_read',               'users:read',               'View user accounts'),
    ('perm_users_write',              'users:write',              'Create/update user accounts'),
    ('perm_roles_read',               'roles:read',               'View roles and permissions'),
    ('perm_roles_write',              'roles:write',              'Create/update roles, assign permissions'),
    ('perm_documents_read',           'documents:read',           'Read Documents'),
    ('perm_documents_write',          'documents:write',          'Write Documents'),
    ('perm_holidays_read',            'holidays:read',            'View holidays'),
    ('perm_holidays_write',           'holidays:write',           'Create/update holidays'),
    ('perm_audit_read',               'audit:read',               'View audit logs and system activity history'),
    ('perm_departments_read',         'departments:read',         'View departments'),
    ('perm_departments_write',        'departments:write',        'Create/update departments'),
    ('perm_attendance_read',          'attendance:read',          'View attendance records'),
    ('perm_attendance_write',         'attendance:write',         'Create/update attendance records'),
    ('perm_attendance_devices_read',  'attendance-devices:read',  'View attendance devices'),
    ('perm_attendance_devices_write', 'attendance-devices:write', 'Manage attendance devices'),
    ('perm_approval_flows_read',      'approval-flows:read',      'View approval flows'),
    ('perm_approval_flows_write',     'approval-flows:write',     'Create/update approval flows'),
    ('perm_approval_read',            'approval:read',            'View approval flows and configuration'),
    ('perm_approval_write',           'approval:write',           'Create/update approval flows and steps'),
    ('perm_shift_read',               'shift:read',               'View shifts'),
    ('perm_shift_write',              'shift:write',              'Create/update shifts'),
    ('perm_permission_request',       'permission:request',       'Submit permission/excuse requests'),
    ('perm_permission_approve',       'permission:approve',       'Approve/reject permission requests'),
    ('perm_permission_read',          'permission:read',          'View all permission requests across the org'),
    ('perm_employee_profile_changes_read',  'employee-profile-changes:read',  'View employee profile change requests'),
    ('perm_employee_profile_changes_write', 'employee-profile-changes:write', 'Submit employee profile change requests')
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Roles
-- (019_seed_admin_role, 021_seed_user_role, 032, 036, 068, 069, 073, 079)
-- ============================================================================
INSERT INTO roles (uid, name, description, is_system, scope_type) VALUES
    ('role_admin',                     'Admin',                       'System administrator with all permissions',                              TRUE,  'global')
ON CONFLICT (uid) DO NOTHING;

INSERT INTO roles (uid, name, description, is_system, scope_type) VALUES
    ('role_employee',                  'Employee',                    'Default role for all employees with basic self-service permissions',      TRUE,  'self')
ON CONFLICT (uid) DO NOTHING;

INSERT INTO roles (uid, name, description, is_system, scope_type) VALUES
    ('role_department_manager',        'Department Manager',          'Department Manager with approval authority for their department',         FALSE, 'department')
ON CONFLICT (uid) DO NOTHING;

INSERT INTO roles (uid, name, description, is_system, scope_type) VALUES
    ('role_user_00000000000000000000000', 'user',                    'Regular user role for employees',                                        FALSE, 'self')
ON CONFLICT (uid) DO NOTHING;

INSERT INTO roles (uid, name, description, is_system, scope_type) VALUES
    ('role_dean',                      'Dean',                        'Dean role with approval and leave visibility access',                    TRUE,  'global')
ON CONFLICT (uid) DO NOTHING;

INSERT INTO roles (uid, name, description, is_system, scope_type) VALUES
    ('role_university_human_resources', 'University Human Resources', 'University HR with full HR management permissions',                      FALSE, 'global')
ON CONFLICT (uid) DO NOTHING;

INSERT INTO roles (uid, name, description, is_system, scope_type) VALUES
    ('role_it_manager',                'IT Manager',                  'IT manager with authority to manually correct attendance logs',           FALSE, 'global')
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Role Permissions
-- (019, 021, 032, 033, 035, 036, 038, 040, 042, 057, 058, 059, 067, 069,
--  070, 071, 073, 079, 20260426120000, 20260426144700, 20260426153000,
--  20260502120100)
-- ============================================================================

-- Admin: all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE uid = 'role_admin'),
    id
FROM permissions
ON CONFLICT DO NOTHING;

-- Employee role permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE uid = 'role_employee'),
    id
FROM permissions
WHERE code IN (
    'leave:request', 'leave:read', 'holidays:read', 'employees:read',
    'attendance:read', 'permission:request'
)
ON CONFLICT DO NOTHING;

-- Department Manager permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE uid = 'role_department_manager'),
    id
FROM permissions
WHERE code IN (
    'employees:read', 'departments:read',
    'leave:read', 'leave:approve', 'leave:request',
    'attendance:read', 'attendance-devices:read',
    'holidays:read',
    'permission:request', 'permission:approve'
)
ON CONFLICT DO NOTHING;

-- User role permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE uid = 'role_user_00000000000000000000000'),
    id
FROM permissions
WHERE code IN ('employees:read', 'leave:read')
ON CONFLICT DO NOTHING;

-- Dean permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE uid = 'role_dean'),
    id
FROM permissions
WHERE code IN (
    'approval:read', 'approval:write',
    'leave:approve', 'leave:read',
    'documents:read', 'departments:read',
    'permission:request', 'permission:approve'
)
ON CONFLICT DO NOTHING;

-- University Human Resources permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE uid = 'role_university_human_resources'),
    id
FROM permissions
WHERE code IN (
    'approval:read', 'approval:write',
    'leave:approve', 'leave:read',
    'documents:read', 'departments:read',
    'leave-types:read', 'leave-types:write',
    'attendance-devices:read', 'attendance-devices:write',
    'attendance:read', 'attendance:write',
    'holidays:write', 'holidays:read',
    'permission:request', 'permission:approve', 'permission:read'
)
ON CONFLICT DO NOTHING;

-- IT Manager permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE uid = 'role_it_manager'),
    id
FROM permissions
WHERE code IN (
    'attendance:read', 'attendance:write',
    'attendance-devices:read',
    'employees:read', 'departments:read'
)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Users
-- (020_seed_admin_user, 055_seed_all_cases, 079_seed_dean_and_university_hr)
-- ============================================================================
INSERT INTO users (uid, phone, is_active, preferred_language, created_at, updated_at) VALUES
    ('usr_admin_seed_00000000000000000000',   '+201000000000', TRUE, 'en', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (phone) DO NOTHING;

INSERT INTO users (uid, phone, is_active, preferred_language, created_at, updated_at) VALUES
    ('usr_manager_seed_000000000000000000',   '+201000000001', TRUE, 'en', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (phone) DO NOTHING;

-- Case users (from 055_seed_all_cases)
INSERT INTO users (uid, phone, employee_uid, is_active, created_at, updated_at) VALUES
    ('usr_case_active_0000000000000000000001',  '+201199200001', 'emp_00000000000000000000000000000001', TRUE,  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('usr_case_inactive_000000000000000000001', '+201199200002', 'emp_case_inactive_000000000000000000001', FALSE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('usr_case_mgr_it_0000000000000000000001',  '+201199200003', 'emp_00000000000000000000000000000003', TRUE,  CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (phone) DO NOTHING;

-- Dean / UHR fallback users (from 079)
INSERT INTO users (uid, phone, is_active, created_at, updated_at)
SELECT
    'usr_seed_dean_20260423',
    '+201000000072',
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
WHERE (
    SELECT COUNT(*)
    FROM users u
    WHERE u.uid NOT IN ('usr_seed_dean_20260423', 'usr_seed_university_hr_20260423')
      AND NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id)
) = 0
ON CONFLICT (phone) DO NOTHING;

INSERT INTO users (uid, phone, is_active, created_at, updated_at)
SELECT
    'usr_seed_university_hr_20260423',
    '+201000000073',
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
WHERE (
    SELECT COUNT(*)
    FROM users u
    WHERE u.uid NOT IN ('usr_seed_dean_20260423', 'usr_seed_university_hr_20260423')
      AND NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id)
) <= 1
ON CONFLICT (phone) DO NOTHING;

-- ============================================================================
-- Departments
-- (047_seed_departments)
-- ============================================================================
INSERT INTO departments (uid, code, name_en, name_ar, is_active) VALUES
    ('dept_hr',   'HR',  'Human Resources',       'الموارد البشرية',       TRUE),
    ('dept_it',   'IT',  'Information Technology', 'تكنولوجيا المعلومات', TRUE),
    ('dept_fin',  'FIN', 'Finance',               'المالية',               TRUE),
    ('dept_ops',  'OPS', 'Operations',            'العمليات',              TRUE),
    ('dept_sec',  'SEC', 'Security',              'الأمن',                 TRUE),
    ('dept_admin','ADM', 'Administration',        'الإدارة',               TRUE),
    ('dept_acad', 'ACA', 'Academic Affairs',      'الشؤون الأكاديمية',    TRUE),
    ('dept_lib',  'LIB', 'Library',               'المكتبة',               TRUE)
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- User Roles
-- (020, 021, 032, 055, 057, 058, 079, 20260426153000)
-- ============================================================================
-- Admin user → Admin role
INSERT INTO user_roles (user_id, role_id, created_at)
SELECT
    (SELECT id FROM users WHERE uid = 'usr_admin_seed_00000000000000000000'),
    (SELECT id FROM roles WHERE uid = 'role_admin'),
    CURRENT_TIMESTAMP
ON CONFLICT DO NOTHING;

-- Manager user → Department Manager role (scoped to Security)
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT
    (SELECT id FROM users WHERE uid = 'usr_manager_seed_000000000000000000'),
    (SELECT id FROM roles WHERE uid = 'role_department_manager'),
    'dept_sec',
    CURRENT_TIMESTAMP
ON CONFLICT DO NOTHING;

-- Case: active user → Employee role
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT
    u.id, r.id, NULL, CURRENT_TIMESTAMP
FROM users u
JOIN roles r ON r.uid = 'role_employee'
WHERE u.uid = 'usr_case_active_0000000000000000000001'
ON CONFLICT DO NOTHING;

-- Case: IT manager user → Department Manager role (scoped to IT)
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT
    u.id, r.id, 'dept_it', CURRENT_TIMESTAMP
FROM users u
JOIN roles r ON r.uid = 'role_department_manager'
WHERE u.uid = 'usr_case_mgr_it_0000000000000000000001'
ON CONFLICT DO NOTHING;

-- Dean user assignment (079)
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT chosen.user_id, r.id, NULL, CURRENT_TIMESTAMP
FROM roles r
CROSS JOIN (
    SELECT COALESCE(
        (
            SELECT u.id FROM users u
            WHERE u.uid NOT IN ('usr_seed_dean_20260423','usr_seed_university_hr_20260423')
              AND NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id)
            ORDER BY u.created_at, u.id LIMIT 1
        ),
        (SELECT u.id FROM users u WHERE u.uid = 'usr_seed_dean_20260423')
    ) AS user_id
) chosen
WHERE r.uid = 'role_dean' AND chosen.user_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- University HR user assignment (079)
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT chosen.user_id, r.id, NULL, CURRENT_TIMESTAMP
FROM roles r
CROSS JOIN (
    SELECT COALESCE(
        (
            SELECT u.id FROM users u
            WHERE u.uid NOT IN ('usr_seed_dean_20260423','usr_seed_university_hr_20260423')
              AND NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id)
            ORDER BY u.created_at, u.id LIMIT 1 OFFSET 1
        ),
        (SELECT u.id FROM users u WHERE u.uid = 'usr_seed_university_hr_20260423')
    ) AS user_id
) chosen
WHERE r.uid = 'role_university_human_resources' AND chosen.user_id IS NOT NULL
ON CONFLICT DO NOTHING;

-- UHR user also gets Employee role (079 / 20260426153000)
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT uhr_users.user_id, employee_role.id, NULL, CURRENT_TIMESTAMP
FROM (
    SELECT DISTINCT ur.user_id
    FROM user_roles ur
    JOIN roles r ON r.id = ur.role_id
    WHERE r.uid = 'role_university_human_resources'
) uhr_users
JOIN roles employee_role ON employee_role.uid = 'role_employee'
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Leave Types
-- (009, 010, 033, 072/074, 20260428001720, 20260501110000)
-- ============================================================================
INSERT INTO leave_types (uid, code, name_en, name_ar, default_balance, max_consecutive, recording_deadline_days, advance_notice_days, is_active, is_paid) VALUES
    ('ltype_00000000000000000000000000000001', 'CASUAL',               'Casual Leave',              'إجازة عارضة',               7,   2,   2,    NULL, TRUE, TRUE),
    ('ltype_00000000000000000000000000000002', 'ANNUAL',               'Annual Leave',              'إجازة سنوية',              30,  30,  30,   7,    TRUE, TRUE),
    ('ltype_00000000000000000000000000000003', 'SICK',                 'Sick Leave',                'إجازة مرضية',             180, 365,  30,   NULL, TRUE, TRUE),
    ('ltype_00000000000000000000000000000004', 'SPECIAL_PAID',         'Special Paid Leave',        'إجازة خاصة بأجر',        365, 365,  30,   NULL, TRUE, TRUE),
    ('ltype_00000000000000000000000000000005', 'SPECIAL_UNPAID',       'Special Unpaid Leave',      'إجازة خاصة بدون أجر',    365, 365,  30,   NULL, TRUE, FALSE),
    ('ltype_00000000000000000000000000000011', 'REGULAR',              'Regular Leave',             'إجازة اعتيادى',           21,  90,   2,   7,    TRUE, TRUE),
    ('ltype_00000000000000000000000000000012', 'MATERNITY',            'Maternity Leave',           'إجازة وضع',              120, 365,  30,   NULL, TRUE, TRUE),
    ('ltype_00000000000000000000000000000009', 'SPECIAL_PAID_EXT',     'Special Paid Leave',        'إجازة خاصة بمرتب',       365, 365,  30,   NULL, TRUE, TRUE),
    ('ltype_00000000000000000000000000000010', 'SPECIAL_UNPAID_EXT',   'Special Unpaid Leave',      'إجازة خاصة بدون مرتب',   365, 365,  30,   NULL, TRUE, FALSE)
ON CONFLICT (code) DO NOTHING;

-- Legacy leave types from 033 (may not be in final schema but present in migrations copy)
INSERT INTO leave_types (uid, code, name_en, name_ar, default_balance, max_consecutive, recording_deadline_days, advance_notice_days, is_active, is_paid) VALUES
    ('ltype_00000000000000000000000000000006', 'CHILD_CARE',              'Child Care Leave',         'رعاية طفل',       730, 730,  30, 14, TRUE, TRUE),
    ('ltype_00000000000000000000000000000007', 'SPOUSE_ACCOMPANIMENT',    'Spouse Accompaniment Leave','مرافقة زوج',     365, 365,  30, 14, TRUE, TRUE)
ON CONFLICT (code) DO NOTHING;

-- Apply is_paid = FALSE for unpaid types (20260501110000)
UPDATE leave_types SET is_paid = FALSE WHERE code IN ('SPECIAL_UNPAID', 'SPECIAL_UNPAID_EXT');

-- Apply recording_deadline_days corrections (20260428001720)
UPDATE leave_types SET recording_deadline_days = 30;
UPDATE leave_types SET recording_deadline_days = 2  WHERE code IN ('CASUAL', 'REGULAR');
UPDATE leave_types SET max_consecutive = 90          WHERE code = 'REGULAR';
UPDATE leave_types SET max_consecutive = 730         WHERE code = 'CHILD_CARE';
UPDATE leave_types SET default_balance = 120         WHERE code = 'MATERNITY';
UPDATE leave_types SET default_balance = 365         WHERE code IN ('SPECIAL_PAID', 'SPECIAL_UNPAID', 'SPECIAL_PAID_EXT', 'SPECIAL_UNPAID_EXT');
UPDATE leave_types SET max_consecutive = 365
WHERE code NOT IN ('REGULAR', 'CASUAL', 'CHILD_CARE');

-- ============================================================================
-- Approval Flows
-- (032_seed_approval_flow, 20260502120100)
-- ============================================================================
INSERT INTO approval_flows (uid, code, name_en, name_ar, description, is_active) VALUES
    ('apf_leave_default',      'leave',      'Leave Request Approval', 'اعتماد طلب الإجازة',  'Default approval flow for leave requests requiring approval', TRUE),
    ('apf_permission_default', 'permission', 'Permission Approval',    'اعتماد طلب الإذن',    'Default approval flow for permission requests',              TRUE),
    ('apf_case_two_step',      'case_two_step', 'Case Two-Step Flow',  'تدفق اعتمادات تجريبي من مرحلتين', 'Seed flow to cover current_step progression and pending-at-step-2 cases', TRUE)
ON CONFLICT (code) DO NOTHING;

-- ============================================================================
-- Approval Flow Steps
-- (032, 20260428003937, 20260502120100)
-- ============================================================================
-- Leave flow: Step 1 – Department Manager
INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid) VALUES
    ('afs_leave_step1', 'apf_leave_default', 1, 'role_department_manager')
ON CONFLICT (approval_flow_uid, step_order) DO NOTHING;

-- Leave flow: Step 2 – Dean
INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
SELECT 'afs_leave_step2_dean', 'apf_leave_default', 2, 'role_dean'
WHERE EXISTS (SELECT 1 FROM roles WHERE uid = 'role_dean')
  AND NOT EXISTS (SELECT 1 FROM approval_flow_steps WHERE approval_flow_uid = 'apf_leave_default' AND step_order = 2)
ON CONFLICT (approval_flow_uid, step_order) DO NOTHING;

-- Leave flow: Step 3 – University HR
INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
SELECT 'afs_leave_step3_university_hr', 'apf_leave_default', 3, 'role_university_human_resources'
WHERE EXISTS (SELECT 1 FROM roles WHERE uid = 'role_university_human_resources')
  AND NOT EXISTS (SELECT 1 FROM approval_flow_steps WHERE approval_flow_uid = 'apf_leave_default' AND step_order = 3)
ON CONFLICT (approval_flow_uid, step_order) DO NOTHING;

-- Permission flow: Step 1 – Department Manager
INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
SELECT 'afs_permission_step1_dm', 'apf_permission_default', 1, 'role_department_manager'
WHERE EXISTS (SELECT 1 FROM approval_flows WHERE uid = 'apf_permission_default')
  AND EXISTS (SELECT 1 FROM roles WHERE uid = 'role_department_manager')
ON CONFLICT (approval_flow_uid, step_order) DO NOTHING;

-- Permission flow: Step 2 – University HR
INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
SELECT 'afs_permission_step2_uhr', 'apf_permission_default', 2, 'role_university_human_resources'
WHERE EXISTS (SELECT 1 FROM approval_flows WHERE uid = 'apf_permission_default')
  AND EXISTS (SELECT 1 FROM roles WHERE uid = 'role_university_human_resources')
ON CONFLICT (approval_flow_uid, step_order) DO NOTHING;

-- Case two-step flow steps (055)
INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid) VALUES
    ('afs_case_two_step_1', 'apf_case_two_step', 1, 'role_department_manager'),
    ('afs_case_two_step_2', 'apf_case_two_step', 2, 'role_department_manager')
ON CONFLICT (approval_flow_uid, step_order) DO NOTHING;

-- ============================================================================
-- Link Leave Types to Approval Flow
-- (033_seed_permissions_and_leave_types)
-- ============================================================================
UPDATE leave_types
SET approval_flow_uid = 'apf_leave_default'
WHERE code IN ('ANNUAL', 'SICK', 'SPECIAL_PAID', 'SPECIAL_UNPAID',
               'SPECIAL_PAID_EXT', 'SPECIAL_UNPAID_EXT', 'REGULAR',
               'MATERNITY', 'CHILD_CARE', 'SPOUSE_ACCOMPANIMENT')
  AND approval_flow_uid IS NULL;

-- ============================================================================
-- Departments
-- (047_seed_departments)
-- ============================================================================
INSERT INTO departments (uid, code, name_en, name_ar, is_active) VALUES
    ('dept_hr',   'HR',  'Human Resources',       'الموارد البشرية',       TRUE),
    ('dept_it',   'IT',  'Information Technology', 'تكنولوجيا المعلومات', TRUE),
    ('dept_fin',  'FIN', 'Finance',               'المالية',               TRUE),
    ('dept_ops',  'OPS', 'Operations',            'العمليات',              TRUE),
    ('dept_sec',  'SEC', 'Security',              'الأمن',                 TRUE),
    ('dept_admin','ADM', 'Administration',        'الإدارة',               TRUE),
    ('dept_acad', 'ACA', 'Academic Affairs',      'الشؤون الأكاديمية',    TRUE),
    ('dept_lib',  'LIB', 'Library',               'المكتبة',               TRUE)
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Employees (sample set)
-- (010_seed_employees_and_leaves, 054_backfill_seed_attendance_exceptions,
--  055_seed_all_cases)
-- ============================================================================
INSERT INTO employees (uid, name, mobile, government_id, university_id, email, hire_date, status, type, sub_type) VALUES
    ('emp_00000000000000000000000000000001', 'أحمد محمد حسن',        '+201001234567', '28501151234567', 'UNI001', 'ahmed.hassan@university.edu.eg',    '2020-01-15', 'active',     'permanent', 'normal'),
    ('emp_00000000000000000000000000000002', 'فاطمة علي إبراهيم',    '+201012345678', '29002281234568', 'UNI002', 'fatma.ibrahim@university.edu.eg',   '2019-06-01', 'active',     'permanent', 'special_needs'),
    ('emp_00000000000000000000000000000003', 'محمد محمود سعيد',      '+201023456789', '28803151234569', 'UNI003', 'mohamed.said@university.edu.eg',    '2021-09-01', 'active',     'temporary', 'contract_employees'),
    ('emp_00000000000000000000000000000004', 'نادية خالد عمر',       '+201034567890', '29105201234570', 'UNI004', 'nadia.omar@university.edu.eg',      '2018-03-15', 'active',     'temporary', 'comprehensive_bonus'),
    ('emp_00000000000000000000000000000005', 'يوسف إبراهيم فاروق',   '+201045678901', '28707101234571', 'UNI005', 'youssef.farouk@university.edu.eg',  '2022-02-01', 'active',     'temporary', 'separation_termination_for_budget'),
    ('emp_00000000000000000000000000000006', 'مريم أحمد نور',        '+201056789012', '29209251234572', 'UNI006', 'mariam.nour@university.edu.eg',     '2020-11-15', 'active',     'permanent', 'special_needs'),
    ('emp_00000000000000000000000000000007', 'حسن علي مصطفى',        '+201067890123', '28604181234573', 'UNI007', 'hassan.mostafa@university.edu.eg',  '2017-08-01', 'active',     'permanent', 'normal'),
    ('emp_00000000000000000000000000000008', 'سارة محمد عزت',        '+201078901234', '29401051234574', 'UNI008', 'sara.ezzat@university.edu.eg',      '2023-01-15', 'active',     'temporary', 'contract_employees'),
    ('emp_00000000000000000000000000000009', 'خالد عبدالرحمن',       '+201089012345', '28211221234575', 'UNI009', 'khaled.rahman@university.edu.eg',   '2016-04-01', 'active',     'temporary', 'comprehensive_bonus'),
    ('emp_00000000000000000000000000000010', 'ليلى سمير حلمي',       '+201090123456', '29308121234576', 'UNI010', 'layla.helmy@university.edu.eg',     '2021-07-01', 'active',     'permanent', 'special_needs'),
    ('emp_00000000000000000000000000000011', 'عمر طارق زكي',         '+201101234567', '28909301234577', 'UNI011', 'omar.zaki@university.edu.eg',       '2019-12-01', 'active',     'temporary', 'separation_termination_for_budget'),
    ('emp_00000000000000000000000000000012', 'هبة فتحي صالح',        '+201112345678', '29106151234578', 'UNI012', 'heba.saleh@university.edu.eg',      '2020-05-15', 'active',     'permanent', 'normal')
ON CONFLICT (uid) DO NOTHING;

-- Case employees (055)
INSERT INTO employees (uid, name, mobile, government_id, university_id, email, hire_date, status, department_uid, type, sub_type) VALUES
    ('emp_case_inactive_000000000000000000001',  'Case Employee Inactive',    '+201199100001', '29901010000001', 'UNICASE001', 'case.inactive@university.edu.eg',    '2024-01-15', 'inactive',   'dept_admin', 'temporary', 'comprehensive_bonus'),
    ('emp_case_terminated_0000000000000000001',  'Case Employee Terminated',  '+201199100002', '29901010000002', 'UNICASE002', 'case.terminated@university.edu.eg', '2023-06-01', 'terminated', 'dept_ops',   'temporary', 'separation_termination_for_budget')
ON CONFLICT (uid) DO NOTHING;

-- Assign departments to original 12 employees (050_assign_seed_employee_departments)
UPDATE employees
SET department_uid = CASE uid
    WHEN 'emp_00000000000000000000000000000001' THEN 'dept_hr'
    WHEN 'emp_00000000000000000000000000000002' THEN 'dept_hr'
    WHEN 'emp_00000000000000000000000000000003' THEN 'dept_it'
    WHEN 'emp_00000000000000000000000000000004' THEN 'dept_fin'
    WHEN 'emp_00000000000000000000000000000005' THEN 'dept_acad'
    WHEN 'emp_00000000000000000000000000000006' THEN 'dept_admin'
    WHEN 'emp_00000000000000000000000000000007' THEN 'dept_sec'
    WHEN 'emp_00000000000000000000000000000008' THEN 'dept_lib'
    WHEN 'emp_00000000000000000000000000000009' THEN 'dept_ops'
    WHEN 'emp_00000000000000000000000000000010' THEN 'dept_admin'
    WHEN 'emp_00000000000000000000000000000011' THEN 'dept_it'
    WHEN 'emp_00000000000000000000000000000012' THEN 'dept_fin'
END
WHERE uid IN (
    'emp_00000000000000000000000000000001', 'emp_00000000000000000000000000000002',
    'emp_00000000000000000000000000000003', 'emp_00000000000000000000000000000004',
    'emp_00000000000000000000000000000005', 'emp_00000000000000000000000000000006',
    'emp_00000000000000000000000000000007', 'emp_00000000000000000000000000000008',
    'emp_00000000000000000000000000000009', 'emp_00000000000000000000000000000010',
    'emp_00000000000000000000000000000011', 'emp_00000000000000000000000000000012'
);

-- ============================================================================
-- Departmental Employees (5 per dept × 8 depts = 40 employees)
-- (056_seed_department_employees_attendance_leaves)
-- ============================================================================
INSERT INTO employees (uid, name, mobile, government_id, university_id, email, hire_date, status, department_uid, type, sub_type, created_at, updated_at)
VALUES
    -- dept_hr
    ('emp56_HR_01','HR Department Manager',    '+201188011000','29900000000011','U56HR01','hr.seed01@university.edu.eg',   '2022-01-06','active','dept_hr',  'permanent','normal',     CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_HR_02','HR Staff 02',              '+201188011200','29900000000012','U56HR02','hr.seed02@university.edu.eg',   '2022-01-07','active','dept_hr',  'permanent','special_needs',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_HR_03','HR Staff 03',              '+201188011300','29900000000013','U56HR03','hr.seed03@university.edu.eg',   '2022-01-08','active','dept_hr',  'temporary','contract_employees',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_HR_04','HR Staff 04',              '+201188011400','29900000000014','U56HR04','hr.seed04@university.edu.eg',   '2022-01-09','active','dept_hr',  'temporary','comprehensive_bonus',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_HR_05','HR Staff 05',              '+201188011500','29900000000015','U56HR05','hr.seed05@university.edu.eg',   '2022-01-10','active','dept_hr',  'temporary','separation_termination_for_budget',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    -- dept_it
    ('emp56_IT_01','IT Department Manager',    '+201188021000','29900000000021','U56IT01','it.seed01@university.edu.eg',   '2022-01-11','active','dept_it',  'permanent','normal',     CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_IT_02','IT Staff 02',              '+201188021200','29900000000022','U56IT02','it.seed02@university.edu.eg',   '2022-01-12','active','dept_it',  'permanent','special_needs',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_IT_03','IT Staff 03',              '+201188021300','29900000000023','U56IT03','it.seed03@university.edu.eg',   '2022-01-13','active','dept_it',  'temporary','contract_employees',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_IT_04','IT Staff 04',              '+201188021400','29900000000024','U56IT04','it.seed04@university.edu.eg',   '2022-01-14','active','dept_it',  'temporary','comprehensive_bonus',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_IT_05','IT Staff 05',              '+201188021500','29900000000025','U56IT05','it.seed05@university.edu.eg',   '2022-01-15','active','dept_it',  'temporary','separation_termination_for_budget',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    -- dept_fin
    ('emp56_FIN_01','FIN Department Manager',  '+201188031000','29900000000031','U56FIN01','fin.seed01@university.edu.eg', '2022-01-16','active','dept_fin', 'permanent','normal',     CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_FIN_02','FIN Staff 02',            '+201188031200','29900000000032','U56FIN02','fin.seed02@university.edu.eg', '2022-01-17','active','dept_fin', 'permanent','special_needs',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_FIN_03','FIN Staff 03',            '+201188031300','29900000000033','U56FIN03','fin.seed03@university.edu.eg', '2022-01-18','active','dept_fin', 'temporary','contract_employees',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_FIN_04','FIN Staff 04',            '+201188031400','29900000000034','U56FIN04','fin.seed04@university.edu.eg', '2022-01-19','active','dept_fin', 'temporary','comprehensive_bonus',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_FIN_05','FIN Staff 05',            '+201188031500','29900000000035','U56FIN05','fin.seed05@university.edu.eg', '2022-01-20','active','dept_fin', 'temporary','separation_termination_for_budget',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    -- dept_ops
    ('emp56_OPS_01','OPS Department Manager',  '+201188041000','29900000000041','U56OPS01','ops.seed01@university.edu.eg', '2022-01-21','active','dept_ops', 'permanent','normal',     CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_OPS_02','OPS Staff 02',            '+201188041200','29900000000042','U56OPS02','ops.seed02@university.edu.eg', '2022-01-22','active','dept_ops', 'permanent','special_needs',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_OPS_03','OPS Staff 03',            '+201188041300','29900000000043','U56OPS03','ops.seed03@university.edu.eg', '2022-01-23','active','dept_ops', 'temporary','contract_employees',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_OPS_04','OPS Staff 04',            '+201188041400','29900000000044','U56OPS04','ops.seed04@university.edu.eg', '2022-01-24','active','dept_ops', 'temporary','comprehensive_bonus',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_OPS_05','OPS Staff 05',            '+201188041500','29900000000045','U56OPS05','ops.seed05@university.edu.eg', '2022-01-25','active','dept_ops', 'temporary','separation_termination_for_budget',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    -- dept_sec
    ('emp56_SEC_01','SEC Department Manager',  '+201188051000','29900000000051','U56SEC01','sec.seed01@university.edu.eg', '2022-01-26','active','dept_sec', 'permanent','normal',     CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_SEC_02','SEC Staff 02',            '+201188051200','29900000000052','U56SEC02','sec.seed02@university.edu.eg', '2022-01-27','active','dept_sec', 'permanent','special_needs',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_SEC_03','SEC Staff 03',            '+201188051300','29900000000053','U56SEC03','sec.seed03@university.edu.eg', '2022-01-28','active','dept_sec', 'temporary','contract_employees',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_SEC_04','SEC Staff 04',            '+201188051400','29900000000054','U56SEC04','sec.seed04@university.edu.eg', '2022-01-29','active','dept_sec', 'temporary','comprehensive_bonus',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_SEC_05','SEC Staff 05',            '+201188051500','29900000000055','U56SEC05','sec.seed05@university.edu.eg', '2022-01-30','active','dept_sec', 'temporary','separation_termination_for_budget',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    -- dept_admin
    ('emp56_ADM_01','ADM Department Manager',  '+201188061000','29900000000061','U56ADM01','adm.seed01@university.edu.eg', '2022-01-31','active','dept_admin','permanent','normal',     CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_ADM_02','ADM Staff 02',            '+201188061200','29900000000062','U56ADM02','adm.seed02@university.edu.eg', '2022-02-01','active','dept_admin','permanent','special_needs',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_ADM_03','ADM Staff 03',            '+201188061300','29900000000063','U56ADM03','adm.seed03@university.edu.eg', '2022-02-02','active','dept_admin','temporary','contract_employees',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_ADM_04','ADM Staff 04',            '+201188061400','29900000000064','U56ADM04','adm.seed04@university.edu.eg', '2022-02-03','active','dept_admin','temporary','comprehensive_bonus',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_ADM_05','ADM Staff 05',            '+201188061500','29900000000065','U56ADM05','adm.seed05@university.edu.eg', '2022-02-04','active','dept_admin','temporary','separation_termination_for_budget',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    -- dept_acad
    ('emp56_ACA_01','ACA Department Manager',  '+201188071000','29900000000071','U56ACA01','aca.seed01@university.edu.eg', '2022-02-05','active','dept_acad','permanent','normal',     CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_ACA_02','ACA Staff 02',            '+201188071200','29900000000072','U56ACA02','aca.seed02@university.edu.eg', '2022-02-06','active','dept_acad','permanent','special_needs',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_ACA_03','ACA Staff 03',            '+201188071300','29900000000073','U56ACA03','aca.seed03@university.edu.eg', '2022-02-07','active','dept_acad','temporary','contract_employees',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_ACA_04','ACA Staff 04',            '+201188071400','29900000000074','U56ACA04','aca.seed04@university.edu.eg', '2022-02-08','active','dept_acad','temporary','comprehensive_bonus',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_ACA_05','ACA Staff 05',            '+201188071500','29900000000075','U56ACA05','aca.seed05@university.edu.eg', '2022-02-09','active','dept_acad','temporary','separation_termination_for_budget',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    -- dept_lib
    ('emp56_LIB_01','LIB Department Manager',  '+201188081000','29900000000081','U56LIB01','lib.seed01@university.edu.eg', '2022-02-10','active','dept_lib', 'permanent','normal',     CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_LIB_02','LIB Staff 02',            '+201188081200','29900000000082','U56LIB02','lib.seed02@university.edu.eg', '2022-02-11','active','dept_lib', 'permanent','special_needs',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_LIB_03','LIB Staff 03',            '+201188081300','29900000000083','U56LIB03','lib.seed03@university.edu.eg', '2022-02-12','active','dept_lib', 'temporary','contract_employees',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_LIB_04','LIB Staff 04',            '+201188081400','29900000000084','U56LIB04','lib.seed04@university.edu.eg', '2022-02-13','active','dept_lib', 'temporary','comprehensive_bonus',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('emp56_LIB_05','LIB Staff 05',            '+201188081500','29900000000085','U56LIB05','lib.seed05@university.edu.eg', '2022-02-14','active','dept_lib', 'temporary','separation_termination_for_budget',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Users for dept employees (056)
-- ============================================================================
INSERT INTO users (uid, phone, employee_uid, is_active, created_at, updated_at)
SELECT
    'usr56_' || e.uid,
    e.mobile,
    e.uid,
    TRUE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM employees e
WHERE e.uid LIKE 'emp56_%'
ON CONFLICT (phone) DO NOTHING;

-- User roles for dept employees – everyone gets Employee role (056)
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP
FROM users u
JOIN roles r ON r.uid = 'role_employee'
WHERE u.uid LIKE 'usr56_emp56_%'
ON CONFLICT DO NOTHING;

-- Managers (_01) also get Department Manager role scoped to their dept (056)
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, e.department_uid, CURRENT_TIMESTAMP
FROM users u
JOIN employees e ON e.uid = u.employee_uid
JOIN roles r ON r.uid = 'role_department_manager'
WHERE u.uid LIKE 'usr56_emp56_%_01'
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Leave Balances – original 12 employees (010)
-- ============================================================================
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_casual_2025', e.id, lt.id, 2025, 7,  3 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_annual_2025', e.id, lt.id, 2025, 21, 5 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_sick_2025',   e.id, lt.id, 2025, 30, 2 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_casual_2026', e.id, lt.id, 2026, 7,  1 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_annual_2026', e.id, lt.id, 2026, 21, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_casual_2025', e.id, lt.id, 2025, 7,  5 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_annual_2025', e.id, lt.id, 2025, 21,10 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_sick_2025',   e.id, lt.id, 2025, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_casual_2026', e.id, lt.id, 2026, 7,  2 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_annual_2026', e.id, lt.id, 2026, 21, 3 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_sick_2026',   e.id, lt.id, 2026, 30, 1 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_casual_2025', e.id, lt.id, 2025, 7,  7 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_annual_2025', e.id, lt.id, 2025, 21,15 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_sick_2025',   e.id, lt.id, 2025, 30, 5 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_casual_2026', e.id, lt.id, 2026, 7,  0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_annual_2026', e.id, lt.id, 2026, 21, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_casual_2025', e.id, lt.id, 2025, 7,  2 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000004' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_annual_2025', e.id, lt.id, 2025, 21, 8 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000004' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_sick_2025',   e.id, lt.id, 2025, 30, 3 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000004' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_casual_2026', e.id, lt.id, 2026, 7,  0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000004' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_annual_2026', e.id, lt.id, 2026, 21, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000004' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000004' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_casual_2025', e.id, lt.id, 2025, 7,  4 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000005' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_annual_2025', e.id, lt.id, 2025, 21,12 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000005' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_sick_2025',   e.id, lt.id, 2025, 30, 1 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000005' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_casual_2026', e.id, lt.id, 2026, 7,  1 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000005' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_annual_2026', e.id, lt.id, 2026, 21, 2 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000005' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000005' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_casual_2025', e.id, lt.id, 2025, 7,  1 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000006' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_annual_2025', e.id, lt.id, 2025, 21, 6 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000006' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_sick_2025',   e.id, lt.id, 2025, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000006' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_casual_2026', e.id, lt.id, 2026, 7,  0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000006' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_annual_2026', e.id, lt.id, 2026, 21, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000006' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000006' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_casual_2025', e.id, lt.id, 2025, 7,  6 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_annual_2025', e.id, lt.id, 2025, 21,18 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_sick_2025',   e.id, lt.id, 2025, 30,10 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_casual_2026', e.id, lt.id, 2026, 7,  2 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_annual_2026', e.id, lt.id, 2026, 21, 5 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_casual_2025', e.id, lt.id, 2025, 7,  0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000008' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_annual_2025', e.id, lt.id, 2025, 21, 3 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000008' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_sick_2025',   e.id, lt.id, 2025, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000008' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_casual_2026', e.id, lt.id, 2026, 7,  0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000008' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_annual_2026', e.id, lt.id, 2026, 21, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000008' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000008' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_casual_2025', e.id, lt.id, 2025, 7,  5 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_annual_2025', e.id, lt.id, 2025, 21,21 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_sick_2025',   e.id, lt.id, 2025, 30, 7 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_casual_2026', e.id, lt.id, 2026, 7,  1 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_annual_2026', e.id, lt.id, 2026, 21, 3 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_casual_2025', e.id, lt.id, 2025, 7,  3 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000010' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_annual_2025', e.id, lt.id, 2025, 21, 7 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000010' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_sick_2025',   e.id, lt.id, 2025, 30, 2 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000010' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_casual_2026', e.id, lt.id, 2026, 7,  0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000010' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_annual_2026', e.id, lt.id, 2026, 21, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000010' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000010' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp11_casual_2025', e.id, lt.id, 2025, 7,  2 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000011' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp11_annual_2025', e.id, lt.id, 2025, 21, 4 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000011' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp11_sick_2025',   e.id, lt.id, 2025, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000011' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_casual_2025', e.id, lt.id, 2025, 7,  4 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000012' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_annual_2025', e.id, lt.id, 2025, 21, 9 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000012' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_sick_2025',   e.id, lt.id, 2025, 30, 4 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000012' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_casual_2026', e.id, lt.id, 2026, 7,  1 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000012' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_annual_2026', e.id, lt.id, 2026, 21, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000012' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_sick_2026',   e.id, lt.id, 2026, 30, 0 FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000012' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

-- Leave balances for case emp (055)
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbt_case_annual_2026_emp01', e.id, lt.id, 2026, 21, 0
FROM employees e JOIN leave_types lt ON lt.code='ANNUAL'
WHERE e.uid='emp_00000000000000000000000000000001'
ON CONFLICT DO NOTHING;

-- Leave balances for 56-employees for 2026 (056)
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days, created_at, updated_at)
SELECT
    'lbal56_' || e.uid || '_' || lt.code || '_2026',
    e.id, lt.id, 2026, lt.default_balance, 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
JOIN leave_types lt ON lt.code IN ('CASUAL','ANNUAL','SICK')
WHERE e.uid LIKE 'emp56_%'
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Leave Records – original 12 employees (010)
-- ============================================================================
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000001', e.id, lt.id, '2025-02-15'::DATE,'2025-02-16'::DATE, 2,'2025-02-17 09:00:00'::TIMESTAMP,e.id,'عارضة لظروف عائلية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000002', e.id, lt.id, '2025-03-10'::DATE,'2025-03-10'::DATE, 1,'2025-03-11 10:30:00'::TIMESTAMP,e.id,'عارضة'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000003', e.id, lt.id, '2025-04-20'::DATE,'2025-04-24'::DATE, 5,'2025-04-10 14:00:00'::TIMESTAMP,e.id,'إجازة سنوية - سفر'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000004', e.id, lt.id, '2025-06-01'::DATE,'2025-06-02'::DATE, 2,'2025-06-03 08:30:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000005', e.id, lt.id, '2026-01-05'::DATE,'2026-01-05'::DATE, 1,'2026-01-06 09:15:00'::TIMESTAMP,e.id,'عارضة'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000006', e.id, lt.id, '2025-01-20'::DATE,'2025-01-21'::DATE, 2,'2025-01-22 10:00:00'::TIMESTAMP,e.id,'عارضة طارئة'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000007', e.id, lt.id, '2025-03-01'::DATE,'2025-03-02'::DATE, 2,'2025-03-03 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000008', e.id, lt.id, '2025-05-15'::DATE,'2025-05-15'::DATE, 1,'2025-05-16 08:45:00'::TIMESTAMP,e.id,'عارضة'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000009', e.id, lt.id, '2025-07-01','2025-07-10',10,'2025-06-15 11:00:00',e.id,'إجازة سنوية - صيف'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000010', e.id, lt.id, '2026-01-02'::DATE,'2026-01-03'::DATE, 2,'2026-01-04 09:30:00'::TIMESTAMP,e.id,'عارضة'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000011', e.id, lt.id, '2026-01-12'::DATE,'2026-01-14'::DATE, 3,'2026-01-05 14:00:00'::TIMESTAMP,e.id,'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000012', e.id, lt.id, '2026-01-20'::DATE,'2026-01-20'::DATE, 1,'2026-01-21 08:00:00'::TIMESTAMP,e.id,'مرضية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000002' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000013', e.id, lt.id, '2025-01-10'::DATE,'2025-01-11'::DATE, 2,'2025-01-12 10:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000014', e.id, lt.id, '2025-02-05'::DATE,'2025-02-06'::DATE, 2,'2025-02-07 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000015', e.id, lt.id, '2025-03-20'::DATE,'2025-03-21'::DATE, 2,'2025-03-22 08:30:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000016', e.id, lt.id, '2025-04-15'::DATE,'2025-04-15'::DATE, 1,'2025-04-16 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000017', e.id, lt.id, '2025-05-01','2025-05-15',15,'2025-04-15 14:00:00',e.id,'إجازة سنوية طويلة'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000018', e.id, lt.id, '2025-08-10'::DATE,'2025-08-14'::DATE, 5,'2025-08-15 09:00:00'::TIMESTAMP,e.id,'مرضية - شهادة طبية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000003' AND lt.code='SICK'   ON CONFLICT DO NOTHING;

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000019', e.id, lt.id, '2025-01-05'::DATE,'2025-01-06'::DATE, 2,'2025-01-07 10:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000020', e.id, lt.id, '2025-02-10'::DATE,'2025-02-11'::DATE, 2,'2025-02-12 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000021', e.id, lt.id, '2025-04-01'::DATE,'2025-04-02'::DATE, 2,'2025-04-03 08:30:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000022', e.id, lt.id, '2025-03-01'::DATE,'2025-03-07'::DATE, 7,'2025-02-20 14:00:00'::TIMESTAMP,e.id,'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000023', e.id, lt.id, '2025-06-15','2025-06-25',11,'2025-06-01 10:00:00',e.id,'إجازة سنوية - صيف'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000024', e.id, lt.id, '2025-05-05','2025-05-14',10,'2025-05-15 09:00:00',e.id,'مرضية - عملية جراحية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000025', e.id, lt.id, '2026-01-02'::DATE,'2026-01-03'::DATE, 2,'2026-01-04 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000026', e.id, lt.id, '2026-01-19'::DATE,'2026-01-23'::DATE, 5,'2026-01-10 11:00:00'::TIMESTAMP,e.id,'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000007' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000027', e.id, lt.id, '2025-02-01'::DATE,'2025-02-07'::DATE, 7,'2025-01-20 14:00:00'::TIMESTAMP,e.id,'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000028', e.id, lt.id, '2025-04-10'::DATE,'2025-04-16'::DATE, 7,'2025-04-01 10:00:00'::TIMESTAMP,e.id,'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000029', e.id, lt.id, '2025-08-01'::DATE,'2025-08-07'::DATE, 7,'2025-07-15 11:00:00'::TIMESTAMP,e.id,'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000030', e.id, lt.id, '2025-03-15'::DATE,'2025-03-16'::DATE, 2,'2025-03-17 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000031', e.id, lt.id, '2025-06-01'::DATE,'2025-06-02'::DATE, 2,'2025-06-03 08:30:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000032', e.id, lt.id, '2025-09-10'::DATE,'2025-09-10'::DATE, 1,'2025-09-11 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000033', e.id, lt.id, '2025-07-20'::DATE,'2025-07-26'::DATE, 7,'2025-07-27 10:00:00'::TIMESTAMP,e.id,'مرضية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000034', e.id, lt.id, '2026-01-05'::DATE,'2026-01-05'::DATE, 1,'2026-01-06 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000035', e.id, lt.id, '2026-01-12'::DATE,'2026-01-14'::DATE, 3,'2026-01-05 14:00:00'::TIMESTAMP,e.id,'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000009' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;

-- Historical records for emp01 (2024)
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000036', e.id, lt.id, '2024-01-15'::DATE,'2024-01-16'::DATE, 2,'2024-01-17 09:00:00'::TIMESTAMP,e.id,'عارضة 2024'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000037', e.id, lt.id, '2024-03-10'::DATE,'2024-03-14'::DATE, 5,'2024-03-01 10:00:00'::TIMESTAMP,e.id,'إجازة سنوية 2024'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000038', e.id, lt.id, '2024-05-20'::DATE,'2024-05-21'::DATE, 2,'2024-05-22 08:30:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000039', e.id, lt.id, '2024-07-01','2024-07-10',10,'2024-06-15 11:00:00',e.id,'إجازة صيفية 2024'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000040', e.id, lt.id, '2024-09-05'::DATE,'2024-09-05'::DATE, 1,'2024-09-06 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000041', e.id, lt.id, '2024-10-15'::DATE,'2024-10-17'::DATE, 3,'2024-10-18 10:00:00'::TIMESTAMP,e.id,'مرضية'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='SICK'   ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000042', e.id, lt.id, '2024-11-20'::DATE,'2024-11-21'::DATE, 2,'2024-11-22 09:00:00'::TIMESTAMP,e.id, NULL
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='CASUAL' ON CONFLICT DO NOTHING;
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000043', e.id, lt.id, '2024-12-22'::DATE,'2024-12-26'::DATE, 5,'2024-12-10 14:00:00'::TIMESTAMP,e.id,'إجازة نهاية السنة'
FROM employees e, leave_types lt WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' ON CONFLICT DO NOTHING;

-- ============================================================================
-- Attendance Devices
-- (044_seed_attendance_devices, 055, 056)
-- ============================================================================
INSERT INTO attendance_devices (uid, ip, port, name, location, serial_number, status) VALUES
    ('adev_main_gate',  '192.168.1.201', 4370, 'Main Gate',    'Main Entrance',         'ZK-SERIAL-001', 'offline'),
    ('adev_hr_floor',   '192.168.1.202', 4370, 'HR Floor',     'Building A - Floor 2',  'ZK-SERIAL-002', 'offline'),
    ('adev_seed_001',   '192.168.1.210', 4370, 'Main Gate A',  'Main Entrance',         'ZK-SEED-001',   'offline'),
    ('adev_seed_002',   '192.168.1.211', 4370, 'Main Gate B',  'Main Entrance',         'ZK-SEED-002',   'offline'),
    ('adev_seed_003',   '192.168.1.212', 4370, 'Admin Block 1','Admin Building',        'ZK-SEED-003',   'offline'),
    ('adev_seed_004',   '192.168.1.213', 4370, 'Admin Block 2','Admin Building',        'ZK-SEED-004',   'offline'),
    ('adev_seed_005',   '192.168.1.214', 4370, 'Finance Office','Building B - Floor 1', 'ZK-SEED-005',   'offline'),
    ('adev_seed_006',   '192.168.1.215', 4370, 'Library Gate', 'Library',               'ZK-SEED-006',   'offline'),
    ('adev_seed_007',   '192.168.1.216', 4370, 'Warehouse North','Warehouse North',     'ZK-SEED-007',   'offline'),
    ('adev_seed_008',   '192.168.1.217', 4370, 'Warehouse South','Warehouse South',     'ZK-SEED-008',   'offline'),
    ('adev_seed_009',   '192.168.1.218', 4370, 'Lab Entrance 1','Lab Complex',          'ZK-SEED-009',   'offline'),
    ('adev_seed_010',   '192.168.1.219', 4370, 'Lab Entrance 2','Lab Complex',          'ZK-SEED-010',   'offline'),
    ('adev_seed_011',   '192.168.1.220', 4370, 'Classrooms East','Academic Wing',       'ZK-SEED-011',   'offline'),
    ('adev_seed_012',   '192.168.1.221', 4370, 'Classrooms West','Academic Wing',       'ZK-SEED-012',   'offline'),
    ('adev_seed_013',   '192.168.1.222', 4370, 'IT Department', 'Building C - Floor 3', 'ZK-SEED-013',   'offline'),
    ('adev_seed_014',   '192.168.1.223', 4370, 'HR Annex',     'Building A - Floor 4',  'ZK-SEED-014',   'offline'),
    ('adev_seed_015',   '192.168.1.224', 4370, 'Reception 1',  'Reception',             'ZK-SEED-015',   'offline'),
    ('adev_seed_016',   '192.168.1.225', 4370, 'Reception 2',  'Reception',             'ZK-SEED-016',   'offline'),
    ('adev_seed_017',   '192.168.1.226', 4370, 'Auditorium Backstage','Auditorium',     'ZK-SEED-017',   'offline'),
    ('adev_seed_018',   '192.168.1.227', 4370, 'Parking Gate 1','Parking',              'ZK-SEED-018',   'offline'),
    ('adev_seed_019',   '192.168.1.228', 4370, 'Parking Gate 2','Parking',              'ZK-SEED-019',   'offline'),
    ('adev_seed_020',   '192.168.1.229', 4370, 'Security Cabin','Perimeter',            'ZK-SEED-020',   'offline'),
    ('adev_seed_021',   '192.168.1.230', 4370, 'Dormitory North','Dormitory',           'ZK-SEED-021',   'offline'),
    ('adev_seed_022',   '192.168.1.231', 4370, 'Dormitory South','Dormitory',           'ZK-SEED-022',   'offline'),
    ('adev_seed_023',   '192.168.1.232', 4370, 'Cafeteria Entry','Cafeteria',           'ZK-SEED-023',   'offline'),
    ('adev_seed_024',   '192.168.1.233', 4370, 'Sports Hall',  'Sports Complex',        'ZK-SEED-024',   'offline'),
    ('adev_seed_025',   '192.168.1.234', 4370, 'Clinic Entrance','Campus Clinic',       'ZK-SEED-025',   'offline')
ON CONFLICT (uid) DO NOTHING;

-- Case device variants (055)
INSERT INTO attendance_devices (uid, ip, port, name, location, serial_number, status) VALUES
    ('adev_case_online_000000000000000000001',  '192.168.1.240', 4370, 'Case Online Device',      'Case Lab', 'ZK-CASE-ONLINE-001',  'online'),
    ('adev_case_offline_00000000000000000001',  '192.168.1.241', 4370, 'Case Offline Device',     'Case Lab', 'ZK-CASE-OFFLINE-001', 'offline'),
    ('adev_case_deact_0000000000000000000001',  '192.168.1.242', 4370, 'Case Deactivated Device', 'Case Lab', 'ZK-CASE-DEACT-001',   'deactivated')
ON CONFLICT (uid) DO NOTHING;

-- Departmental devices (056)
INSERT INTO attendance_devices (uid, ip, port, name, location, serial_number, status) VALUES
    ('adev56_hr',  '192.168.2.11', 4370, 'HR Device',  'HR Floor',         'ZK-56-HR-001',  'online'),
    ('adev56_it',  '192.168.2.12', 4370, 'IT Device',  'IT Floor',         'ZK-56-IT-001',  'online'),
    ('adev56_fin', '192.168.2.13', 4370, 'FIN Device', 'Finance Floor',    'ZK-56-FIN-001', 'online'),
    ('adev56_ops', '192.168.2.14', 4370, 'OPS Device', 'Operations Floor', 'ZK-56-OPS-001', 'offline'),
    ('adev56_sec', '192.168.2.15', 4370, 'SEC Device', 'Security Gate',    'ZK-56-SEC-001', 'online'),
    ('adev56_adm', '192.168.2.16', 4370, 'ADM Device', 'Admin Floor',      'ZK-56-ADM-001', 'offline'),
    ('adev56_aca', '192.168.2.17', 4370, 'ACA Device', 'Academic Floor',   'ZK-56-ACA-001', 'online'),
    ('adev56_lib', '192.168.2.18', 4370, 'LIB Device', 'Library Gate',     'ZK-56-LIB-001', 'online')
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Attendance Records – fixed-date sample (046/049)
-- ============================================================================
INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload) VALUES
    ('atr_seed_20260413_emp01_in',  'emp_00000000000000000000000000000001','adev_main_gate','UNI001','2026-04-13 08:52:00','check_in', '{\"source\":\"seed\",\"note\":\"on_time\"}'),
    ('atr_seed_20260413_emp01_out', 'emp_00000000000000000000000000000001','adev_main_gate','UNI001','2026-04-13 16:58:00','check_out','{\"source\":\"seed\",\"note\":\"full_day\"}'),
    ('atr_seed_20260413_emp02_in',  'emp_00000000000000000000000000000002','adev_hr_floor', 'UNI002','2026-04-13 09:18:00','check_in', '{\"source\":\"seed\",\"note\":\"late_arrival\"}'),
    ('atr_seed_20260413_emp02_out', 'emp_00000000000000000000000000000002','adev_hr_floor', 'UNI002','2026-04-13 17:06:00','check_out','{\"source\":\"seed\",\"note\":\"full_day\"}'),
    ('atr_seed_20260413_emp03_in',  'emp_00000000000000000000000000000003','adev_seed_013', 'UNI003','2026-04-13 08:41:00','check_in', '{\"source\":\"seed\",\"note\":\"early_arrival\"}'),
    ('atr_seed_20260413_emp03_out', 'emp_00000000000000000000000000000003','adev_seed_013', 'UNI003','2026-04-13 16:47:00','check_out','{\"source\":\"seed\",\"note\":\"left_early\"}'),
    ('atr_seed_20260413_emp04_in',  'emp_00000000000000000000000000000004','adev_seed_005', 'UNI004','2026-04-13 08:57:00','check_in', '{\"source\":\"seed\",\"note\":\"finance_office\"}'),
    ('atr_seed_20260413_emp04_out', 'emp_00000000000000000000000000000004','adev_seed_005', 'UNI004','2026-04-13 17:11:00','check_out','{\"source\":\"seed\",\"note\":\"overtime\"}'),
    ('atr_seed_20260413_emp05_in',  'emp_00000000000000000000000000000005','adev_seed_011', 'UNI005','2026-04-13 09:03:00','check_in', '{\"source\":\"seed\",\"note\":\"academic_wing\"}'),
    ('atr_seed_20260413_emp05_out', 'emp_00000000000000000000000000000005','adev_seed_011', 'UNI005','2026-04-13 16:55:00','check_out','{\"source\":\"seed\",\"note\":\"regular_day\"}'),
    ('atr_seed_20260414_emp01_in',  'emp_00000000000000000000000000000001','adev_main_gate','UNI001','2026-04-14 08:49:00','check_in', '{\"source\":\"seed\",\"note\":\"regular_day\"}'),
    ('atr_seed_20260414_emp01_out', 'emp_00000000000000000000000000000001','adev_main_gate','UNI001','2026-04-14 17:02:00','check_out','{\"source\":\"seed\",\"note\":\"regular_day\"}'),
    ('atr_seed_20260414_emp02_in',  'emp_00000000000000000000000000000002','adev_hr_floor', 'UNI002','2026-04-14 08:59:00','check_in', '{\"source\":\"seed\",\"note\":\"regular_day\"}'),
    ('atr_seed_20260414_emp02_out', 'emp_00000000000000000000000000000002','adev_hr_floor', 'UNI002','2026-04-14 16:49:00','check_out','{\"source\":\"seed\",\"note\":\"early_checkout\"}'),
    ('atr_seed_20260414_emp03_in',  'emp_00000000000000000000000000000003','adev_seed_013', 'UNI003','2026-04-14 08:44:00','check_in', '{\"source\":\"seed\",\"note\":\"it_department\"}'),
    ('atr_seed_20260414_emp03_out', 'emp_00000000000000000000000000000003','adev_seed_013', 'UNI003','2026-04-14 17:08:00','check_out','{\"source\":\"seed\",\"note\":\"regular_day\"}'),
    ('atr_seed_20260414_emp06_in',  'emp_00000000000000000000000000000006','adev_seed_014', 'UNI006','2026-04-14 09:11:00','check_in', '{\"source\":\"seed\",\"note\":\"hr_annex\"}'),
    ('atr_seed_20260414_emp06_out', 'emp_00000000000000000000000000000006','adev_seed_014', 'UNI006','2026-04-14 17:14:00','check_out','{\"source\":\"seed\",\"note\":\"late_checkout\"}'),
    ('atr_seed_20260414_emp07_in',  'emp_00000000000000000000000000000007','adev_seed_020', 'UNI007','2026-04-14 08:35:00','check_in', '{\"source\":\"seed\",\"note\":\"security_cabin\"}'),
    ('atr_seed_20260414_emp07_out', 'emp_00000000000000000000000000000007','adev_seed_020', 'UNI007','2026-04-14 16:40:00','check_out','{\"source\":\"seed\",\"note\":\"field_shift\"}'),
    ('atr_seed_20260415_emp01_in',  'emp_00000000000000000000000000000001','adev_main_gate','UNI001','2026-04-15 08:47:00','check_in', '{\"source\":\"seed\",\"note\":\"current_day_regular\"}'),
    ('atr_seed_20260415_emp01_out', 'emp_00000000000000000000000000000001','adev_main_gate','UNI001','2026-04-15 17:04:00','check_out','{\"source\":\"seed\",\"note\":\"current_day_regular\"}'),
    ('atr_seed_20260415_emp02_in',  'emp_00000000000000000000000000000002','adev_hr_floor', 'UNI002','2026-04-15 09:24:00','check_in', '{\"source\":\"seed\",\"note\":\"late_current_day\"}'),
    ('atr_seed_20260415_emp02_out', 'emp_00000000000000000000000000000002','adev_hr_floor', 'UNI002','2026-04-15 16:56:00','check_out','{\"source\":\"seed\",\"note\":\"slightly_early_checkout\"}'),
    ('atr_seed_20260415_emp03_in',  'emp_00000000000000000000000000000003','adev_seed_013', 'UNI003','2026-04-15 08:39:00','check_in', '{\"source\":\"seed\",\"note\":\"it_department_current_day\"}'),
    ('atr_seed_20260415_emp03_break_start','emp_00000000000000000000000000000003','adev_seed_013','UNI003','2026-04-15 12:31:00','break_start','{\"source\":\"seed\",\"note\":\"lunch_break\"}'),
    ('atr_seed_20260415_emp03_break_end',  'emp_00000000000000000000000000000003','adev_seed_013','UNI003','2026-04-15 13:02:00','break_end',  '{\"source\":\"seed\",\"note\":\"lunch_break\"}'),
    ('atr_seed_20260415_emp03_out', 'emp_00000000000000000000000000000003','adev_seed_013', 'UNI003','2026-04-15 17:13:00','check_out','{\"source\":\"seed\",\"note\":\"regular_day\"}'),
    ('atr_seed_20260415_emp04_in',  'emp_00000000000000000000000000000004','adev_seed_005', 'UNI004','2026-04-15 08:58:00','check_in', '{\"source\":\"seed\",\"note\":\"missing_checkout_case\"}'),
    ('atr_seed_20260415_emp05_unknown','emp_00000000000000000000000000000005','adev_seed_011','UNI005','2026-04-15 09:01:00','unknown','{\"source\":\"seed\",\"note\":\"unknown_punch_type\"}'),
    ('atr_seed_20260415_emp06_in',  'emp_00000000000000000000000000000006','adev_seed_014', 'UNI006','2026-04-15 09:08:00','check_in', '{\"source\":\"seed\",\"note\":\"hr_annex_current_day\"}'),
    ('atr_seed_20260415_emp06_out', 'emp_00000000000000000000000000000006','adev_seed_014', 'UNI006','2026-04-15 17:22:00','check_out','{\"source\":\"seed\",\"note\":\"overtime\"}'),
    ('atr_seed_20260415_emp07_in',  'emp_00000000000000000000000000000007','adev_seed_020', 'UNI007','2026-04-15 08:28:00','check_in', '{\"source\":\"seed\",\"note\":\"security_shift_start\"}'),
    ('atr_seed_20260415_emp07_out', 'emp_00000000000000000000000000000007','adev_seed_020', 'UNI007','2026-04-15 16:21:00','check_out','{\"source\":\"seed\",\"note\":\"early_departure_case\"}')
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Attendance Records – dept baseline (Day A) & case (Day B) (056)
-- ============================================================================
-- Day A: 2026-04-21 baseline check-in / check-out for all emp56_*
INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_a_in_'  || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad'  THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-21 08:50:00',
    'check_in',
    '{\"source\":\"seed56\",\"case\":\"baseline\"}',
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
WHERE e.uid LIKE 'emp56_%'
ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_a_out_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad'  THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-21 17:00:00',
    'check_out',
    '{\"source\":\"seed56\",\"case\":\"baseline\"}',
    CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM employees e
WHERE e.uid LIKE 'emp56_%'
ON CONFLICT (uid) DO NOTHING;

-- Day B: 2026-04-22 – manager normal (_01), late (_02), missed-out (_03),
--                      missed-in (_04), unknown (_05 IT only)
INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT 'atr56_b_mgr_in_'     || e.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id,'2026-04-22 08:45:00','check_in','{\"source\":\"seed56\",\"case\":\"manager_normal\"}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_01' ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT 'atr56_b_mgr_break_s_'|| e.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id,'2026-04-22 12:30:00','break_start','{\"source\":\"seed56\",\"case\":\"manager_break\"}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_01' ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT 'atr56_b_mgr_break_e_'|| e.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id,'2026-04-22 13:00:00','break_end','{\"source\":\"seed56\",\"case\":\"manager_break\"}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_01' ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT 'atr56_b_mgr_out_'    || e.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id,'2026-04-22 17:05:00','check_out','{\"source\":\"seed56\",\"case\":\"manager_normal\"}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_01' ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT 'atr56_b_late_in_'    || e.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id,'2026-04-22 09:25:00','check_in','{\"source\":\"seed56\",\"case\":\"late_arrival\"}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_02' ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT 'atr56_b_late_out_'   || e.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id,'2026-04-22 17:02:00','check_out','{\"source\":\"seed56\",\"case\":\"late_arrival\"}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_02' ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT 'atr56_b_miss_out_in_'|| e.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id,'2026-04-22 08:58:00','check_in','{\"source\":\"seed56\",\"case\":\"missed_punch_out\"}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_03' ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT 'atr56_b_miss_in_out_'|| e.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id,'2026-04-22 16:55:00','check_out','{\"source\":\"seed56\",\"case\":\"missed_punch_in\"}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_04' ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT 'atr56_b_unknown_'    || e.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id,'2026-04-22 10:00:00','unknown','{\"source\":\"seed56\",\"case\":\"unknown_punch_type\"}',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid = 'emp56_IT_05' ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Fixed-date attendance 2026-04-19 (065)
-- ============================================================================
INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT s.uid, e.uid,
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END,
    e.university_id, s.punched_at::TIMESTAMP, s.punch_type, '{\"source\":\"seed65\",\"date\":\"2026-04-19\"}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM (VALUES
    ('atr65_20260419_hr01_in',  'emp56_HR_01',  '2026-04-19 08:44:00','check_in'),
    ('atr65_20260419_it01_in',  'emp56_IT_01',  '2026-04-19 08:52:00','check_in'),
    ('atr65_20260419_fin01_in', 'emp56_FIN_01', '2026-04-19 08:59:00','check_in'),
    ('atr65_20260419_ops01_in', 'emp56_OPS_01', '2026-04-19 09:05:00','check_in'),
    ('atr65_20260419_sec01_in', 'emp56_SEC_01', '2026-04-19 08:35:00','check_in'),
    ('atr65_20260419_adm01_in', 'emp56_ADM_01', '2026-04-19 08:48:00','check_in'),
    ('atr65_20260419_aca01_in', 'emp56_ACA_01', '2026-04-19 08:57:00','check_in'),
    ('atr65_20260419_lib01_in', 'emp56_LIB_01', '2026-04-19 09:02:00','check_in'),
    ('atr65_20260419_hr01_out',  'emp56_HR_01',  '2026-04-19 17:01:00','check_out'),
    ('atr65_20260419_it01_out',  'emp56_IT_01',  '2026-04-19 17:07:00','check_out'),
    ('atr65_20260419_fin01_out', 'emp56_FIN_01', '2026-04-19 16:55:00','check_out'),
    ('atr65_20260419_sec01_out', 'emp56_SEC_01', '2026-04-19 16:46:00','check_out'),
    ('atr65_20260419_aca01_out', 'emp56_ACA_01', '2026-04-19 17:12:00','check_out')
) AS s(uid, emp_uid, punched_at, punch_type)
JOIN employees e ON e.uid = s.emp_uid
JOIN attendance_devices d ON d.uid = (
    CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm' WHEN e.department_uid='dept_acad' THEN 'adev56_aca' ELSE 'adev56_'||lower(substr(e.department_uid,6)) END
)
ON CONFLICT (uid) DO NOTHING;

-- Long-range attendance 180 days (064) – generates ~5000+ rows for emp56_*
INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
WITH RECURSIVE
days(day_date, day_index) AS (
    VALUES ((CURRENT_DATE - 179)::DATE, 0)
    UNION ALL
    SELECT (day_date + 1)::DATE, day_index + 1
    FROM days WHERE day_index < 179
),
workdays AS (
    SELECT day_date, day_index FROM days
    WHERE EXTRACT(DOW FROM day_date) NOT IN (5, 6)
),
seed_employees AS (
    SELECT e.uid AS employee_uid, e.department_uid, e.university_id,
        CAST(RIGHT(e.uid, 2) AS INTEGER) AS employee_slot,
        CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm'
             WHEN e.department_uid='dept_acad'  THEN 'adev56_aca'
             ELSE 'adev56_'||lower(substr(e.department_uid,6)) END AS device_uid
    FROM employees e WHERE e.uid LIKE 'emp56_%' AND e.status='active'
),
base_rows AS (
    SELECT se.employee_uid, se.department_uid, se.university_id, se.device_uid, se.employee_slot,
        wd.day_date, wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 13 = 0) AS is_missed_punch_in,
        ((wd.day_index + se.employee_slot) % 9  = 0) AS is_late_arrival
    FROM seed_employees se CROSS JOIN workdays wd
)
SELECT
    'atr60_in_'  || br.employee_uid || '_' || REPLACE(br.day_date::text, '-', ''),
    br.employee_uid, br.device_uid, br.university_id,
    (br.day_date::text || ' ' || CASE WHEN br.is_late_arrival THEN '09:24:00' WHEN (br.day_index+br.employee_slot)%5=0 THEN '09:03:00' ELSE '08:47:00' END)::TIMESTAMP,
    'check_in', '{\"source\":\"seed60\",\"case\":\"long_range\"}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM base_rows br
WHERE NOT br.is_absent AND NOT br.is_missed_punch_in
ON CONFLICT (uid) DO NOTHING;

INSERT INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
WITH RECURSIVE
days(day_date, day_index) AS (
    VALUES ((CURRENT_DATE - 179)::DATE, 0)
    UNION ALL
    SELECT (day_date + 1)::DATE, day_index + 1
    FROM days WHERE day_index < 179
),
workdays AS (
    SELECT day_date, day_index FROM days
    WHERE EXTRACT(DOW FROM day_date) NOT IN (5, 6)
),
seed_employees AS (
    SELECT e.uid AS employee_uid, e.department_uid, e.university_id,
        CAST(RIGHT(e.uid, 2) AS INTEGER) AS employee_slot,
        CASE WHEN e.department_uid='dept_admin' THEN 'adev56_adm'
             WHEN e.department_uid='dept_acad'  THEN 'adev56_aca'
             ELSE 'adev56_'||lower(substr(e.department_uid,6)) END AS device_uid
    FROM employees e WHERE e.uid LIKE 'emp56_%' AND e.status='active'
),
base_rows AS (
    SELECT se.employee_uid, se.department_uid, se.university_id, se.device_uid, se.employee_slot,
        wd.day_date, wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 8  = 0) AS is_missed_punch_out,
        ((wd.day_index + se.employee_slot) % 17 = 0) AS is_early_departure
    FROM seed_employees se CROSS JOIN workdays wd
)
SELECT
    'atr60_out_' || br.employee_uid || '_' || REPLACE(br.day_date::text, '-', ''),
    br.employee_uid, br.device_uid, br.university_id,
    (br.day_date::text || ' ' || CASE WHEN br.is_early_departure THEN '15:38:00' WHEN (br.day_index+br.employee_slot)%6=0 THEN '17:19:00' ELSE '16:58:00' END)::TIMESTAMP,
    'check_out', '{\"source\":\"seed60\",\"case\":\"long_range\"}', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM base_rows br
WHERE NOT br.is_absent AND NOT br.is_missed_punch_out
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Attendance Exceptions (053, 054, 055, 056, 064)
-- ============================================================================
-- Fixed-date exceptions from 053/054 (removed duplicates that conflict with 064 generated data)

-- Case exceptions from 055
INSERT INTO attendance_exceptions (uid, employee_uid, attendance_date, exception_type, check_in, check_out, grace_minutes, minutes_delta, created_at, updated_at) VALUES
    ('aex_case_20260420_missed_in',  'emp_00000000000000000000000000000001','2026-04-20','missed_punch_in', NULL,                  '2026-04-20 17:03:00',15,NULL,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('aex_case_20260420_missed_out', 'emp_00000000000000000000000000000002','2026-04-20','missed_punch_out','2026-04-20 09:01:00', NULL,                  15,NULL,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('aex_case_20260420_late',       'emp_00000000000000000000000000000003','2026-04-20','late_arrival',    '2026-04-20 09:22:00', '2026-04-20 17:06:00',15,7,  CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('aex_case_20260420_early',      'emp_00000000000000000000000000000004','2026-04-20','early_departure',  '2026-04-20 08:55:00', '2026-04-20 16:34:00',15,11, CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('aex_case_20260420_absence',    'emp_00000000000000000000000000000005','2026-04-20','absence',          NULL,                   NULL,                  NULL,NULL,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
ON CONFLICT (employee_uid, attendance_date, exception_type) DO UPDATE SET
    uid = EXCLUDED.uid,
    check_in = EXCLUDED.check_in,
    check_out = EXCLUDED.check_out,
    grace_minutes = EXCLUDED.grace_minutes,
    minutes_delta = EXCLUDED.minutes_delta,
    updated_at = CURRENT_TIMESTAMP;

-- Dept exceptions Day B 2026-04-22 (056)
INSERT INTO attendance_exceptions (uid, employee_uid, attendance_date, exception_type, check_in, check_out, grace_minutes, minutes_delta, created_at, updated_at)
SELECT 'aex56_late_'    ||e.uid,e.uid,'2026-04-22','late_arrival',    '2026-04-22 09:25:00','2026-04-22 17:02:00',15,10,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_02'
ON CONFLICT (employee_uid, attendance_date, exception_type) DO UPDATE SET
    uid = EXCLUDED.uid,
    check_in = EXCLUDED.check_in,
    check_out = EXCLUDED.check_out,
    grace_minutes = EXCLUDED.grace_minutes,
    minutes_delta = EXCLUDED.minutes_delta,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO attendance_exceptions (uid, employee_uid, attendance_date, exception_type, check_in, check_out, grace_minutes, minutes_delta, created_at, updated_at)
SELECT 'aex56_miss_out_'||e.uid,e.uid,'2026-04-22','missed_punch_out','2026-04-22 08:58:00',NULL,                  15,NULL,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_03'
ON CONFLICT (employee_uid, attendance_date, exception_type) DO UPDATE SET
    uid = EXCLUDED.uid,
    check_in = EXCLUDED.check_in,
    check_out = EXCLUDED.check_out,
    grace_minutes = EXCLUDED.grace_minutes,
    minutes_delta = EXCLUDED.minutes_delta,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO attendance_exceptions (uid, employee_uid, attendance_date, exception_type, check_in, check_out, grace_minutes, minutes_delta, created_at, updated_at)
SELECT 'aex56_miss_in_' ||e.uid,e.uid,'2026-04-22','missed_punch_in', NULL,'2026-04-22 16:55:00',15,NULL,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_04'
ON CONFLICT (employee_uid, attendance_date, exception_type) DO UPDATE SET
    uid = EXCLUDED.uid,
    check_in = EXCLUDED.check_in,
    check_out = EXCLUDED.check_out,
    grace_minutes = EXCLUDED.grace_minutes,
    minutes_delta = EXCLUDED.minutes_delta,
    updated_at = CURRENT_TIMESTAMP;

INSERT INTO attendance_exceptions (uid, employee_uid, attendance_date, exception_type, check_in, check_out, grace_minutes, minutes_delta, created_at, updated_at)
SELECT 'aex56_absence_' ||e.uid,e.uid,'2026-04-22','absence',          NULL,NULL,NULL,NULL,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM employees e WHERE e.uid LIKE 'emp56_%_05'
ON CONFLICT (employee_uid, attendance_date, exception_type) DO UPDATE SET
    uid = EXCLUDED.uid,
    updated_at = CURRENT_TIMESTAMP;

-- Long-range exceptions 180 days (064)
INSERT INTO attendance_exceptions (uid, employee_uid, attendance_date, exception_type, check_in, check_out, grace_minutes, minutes_delta, created_at, updated_at)
WITH RECURSIVE
days(day_date, day_index) AS (
    VALUES ((CURRENT_DATE - 179)::DATE, 0)
    UNION ALL SELECT (day_date + 1)::DATE, day_index + 1 FROM days WHERE day_index < 179
),
workdays AS (SELECT day_date, day_index FROM days WHERE EXTRACT(DOW FROM day_date) NOT IN (5,6)),
seed_employees AS (
    SELECT e.uid AS employee_uid, CAST(RIGHT(e.uid,2) AS INTEGER) AS employee_slot
    FROM employees e WHERE e.uid LIKE 'emp56_%' AND e.status='active'
),
base_rows AS (
    SELECT se.employee_uid, se.employee_slot, wd.day_date, wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 8  = 0) AS is_missed_punch_out,
        ((wd.day_index + se.employee_slot) % 13 = 0) AS is_missed_punch_in,
        ((wd.day_index + se.employee_slot) % 9  = 0) AS is_late_arrival,
        ((wd.day_index + se.employee_slot) % 17 = 0) AS is_early_departure
    FROM seed_employees se CROSS JOIN workdays wd
)
SELECT 'aex60_abs_'  || br.employee_uid || '_' || REPLACE(br.day_date::text,'-',''),
    br.employee_uid, br.day_date, 'absence', NULL, NULL, NULL, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM base_rows br WHERE br.is_absent
UNION ALL
SELECT 'aex60_late_' || br.employee_uid || '_' || REPLACE(br.day_date::text,'-',''),
    br.employee_uid, br.day_date, 'late_arrival',
    (br.day_date::text||' 09:24:00')::TIMESTAMP,(br.day_date::text||' 16:58:00')::TIMESTAMP,15,9,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM base_rows br WHERE NOT br.is_absent AND br.is_late_arrival
UNION ALL
SELECT 'aex60_mout_' || br.employee_uid || '_' || REPLACE(br.day_date::text,'-',''),
    br.employee_uid, br.day_date, 'missed_punch_out',
    (br.day_date::text||' 08:47:00')::TIMESTAMP,NULL,15,NULL,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM base_rows br WHERE NOT br.is_absent AND br.is_missed_punch_out
UNION ALL
SELECT 'aex60_min_'  || br.employee_uid || '_' || REPLACE(br.day_date::text,'-',''),
    br.employee_uid, br.day_date, 'missed_punch_in',
    NULL,(br.day_date::text||' 16:58:00')::TIMESTAMP,15,NULL,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM base_rows br WHERE NOT br.is_absent AND NOT br.is_missed_punch_out AND br.is_missed_punch_in
UNION ALL
SELECT 'aex60_early_'|| br.employee_uid || '_' || REPLACE(br.day_date::text,'-',''),
    br.employee_uid, br.day_date, 'early_departure',
    (br.day_date::text||' 08:47:00')::TIMESTAMP,(br.day_date::text||' 15:38:00')::TIMESTAMP,0,82,CURRENT_TIMESTAMP,CURRENT_TIMESTAMP
FROM base_rows br WHERE NOT br.is_absent AND NOT br.is_missed_punch_out AND NOT br.is_missed_punch_in AND br.is_early_departure
ON CONFLICT (employee_uid, attendance_date, exception_type) DO NOTHING;

-- ============================================================================
-- OTP Code Variants (055)
-- ============================================================================
INSERT INTO otp_codes (phone, code, expires_at, used, created_at) VALUES
    ('+201199300001','123456', NOW() + INTERVAL '20 minutes', FALSE, NOW()),
    ('+201199300002','654321', NOW() + INTERVAL '20 minutes', TRUE,  NOW()),
    ('+201199300003','111111', NOW() - INTERVAL '20 minutes', FALSE, NOW())
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Refresh Token Variants (055)
-- ============================================================================
INSERT INTO refresh_tokens (token_hash, user_id, expires_at, revoked, created_at)
SELECT 'seedhash_case_refresh_active_001',  u.id, NOW()+INTERVAL '30 days', FALSE, NOW()
FROM users u WHERE u.uid='usr_case_active_0000000000000000000001'
ON CONFLICT DO NOTHING;

INSERT INTO refresh_tokens (token_hash, user_id, expires_at, revoked, created_at)
SELECT 'seedhash_case_refresh_revoked_001', u.id, NOW()+INTERVAL '30 days', TRUE,  NOW()
FROM users u WHERE u.uid='usr_case_active_0000000000000000000001'
ON CONFLICT DO NOTHING;

INSERT INTO refresh_tokens (token_hash, user_id, expires_at, revoked, created_at)
SELECT 'seedhash_case_refresh_expired_001', u.id, NOW()-INTERVAL '30 days', FALSE, NOW()
FROM users u WHERE u.uid='usr_case_active_0000000000000000000001'
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Device Tokens (055)
-- ============================================================================
INSERT INTO device_tokens (uid, user_uid, token, platform, created_at, updated_at) VALUES
    ('dtok_case_android_00000000000000000001','usr_case_active_0000000000000000000001','seed-device-token-android-case-001','android',CURRENT_TIMESTAMP,CURRENT_TIMESTAMP),
    ('dtok_case_ios_000000000000000000000001','usr_case_active_0000000000000000000001','seed-device-token-ios-case-001',    'ios',    CURRENT_TIMESTAMP,CURRENT_TIMESTAMP)
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Approval Requests – case variants (055)
-- ============================================================================
INSERT INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at) VALUES
    ('apr_case_pending_0000000000000000000001',   'apf_leave_default','emp_00000000000000000000000000000002',1,1,'pending',   NOW()-INTERVAL '2 days',  NOW()-INTERVAL '2 days'),
    ('apr_case_approved_00000000000000000000001', 'apf_leave_default','emp_00000000000000000000000000000003',1,1,'approved',  NOW()-INTERVAL '10 days', NOW()-INTERVAL '9 days'),
    ('apr_case_rejected_00000000000000000000001', 'apf_leave_default','emp_00000000000000000000000000000004',1,1,'rejected',  NOW()-INTERVAL '8 days',  NOW()-INTERVAL '7 days'),
    ('apr_case_cancelled_0000000000000000000001', 'apf_leave_default','emp_00000000000000000000000000000005',1,1,'cancelled', NOW()-INTERVAL '6 days',  NOW()-INTERVAL '5 days'),
    ('apr_case_pending_step2_000000000000000001', 'apf_case_two_step','emp_00000000000000000000000000000006',2,2,'pending',   NOW()-INTERVAL '1 days',  NOW()-INTERVAL '1 days')
ON CONFLICT (uid) DO NOTHING;

-- Approval requests 056 (per-dept approved/rejected/pending)
INSERT INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
SELECT 'apr56_'||dc.dept_code||'_approved', 'apf_leave_default',
    (SELECT e.uid FROM employees e WHERE e.department_uid=dc.dept_uid AND e.status='active' ORDER BY (CASE WHEN e.uid LIKE 'emp56_%' THEN 1 ELSE 0 END), e.hire_date, e.id LIMIT 1 OFFSET 1),
    1,1,'approved', NOW()-INTERVAL '5 days', NOW()-INTERVAL '4 days'
FROM (VALUES
    ('dept_hr','HR'),('dept_it','IT'),('dept_fin','FIN'),('dept_ops','OPS'),
    ('dept_sec','SEC'),('dept_admin','ADM'),('dept_acad','ACA'),('dept_lib','LIB')
) AS dc(dept_uid, dept_code)
WHERE (SELECT e.uid FROM employees e WHERE e.department_uid=dc.dept_uid AND e.status='active' ORDER BY (CASE WHEN e.uid LIKE 'emp56_%' THEN 1 ELSE 0 END), e.hire_date, e.id LIMIT 1 OFFSET 1) IS NOT NULL
ON CONFLICT (uid) DO NOTHING;

INSERT INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
SELECT 'apr56_'||dc.dept_code||'_rejected', 'apf_leave_default',
    (SELECT e.uid FROM employees e WHERE e.department_uid=dc.dept_uid AND e.status='active' ORDER BY (CASE WHEN e.uid LIKE 'emp56_%' THEN 1 ELSE 0 END), e.hire_date, e.id LIMIT 1 OFFSET 2),
    1,1,'rejected', NOW()-INTERVAL '4 days', NOW()-INTERVAL '3 days'
FROM (VALUES
    ('dept_hr','HR'),('dept_it','IT'),('dept_fin','FIN'),('dept_ops','OPS'),
    ('dept_sec','SEC'),('dept_admin','ADM'),('dept_acad','ACA'),('dept_lib','LIB')
) AS dc(dept_uid, dept_code)
WHERE (SELECT e.uid FROM employees e WHERE e.department_uid=dc.dept_uid AND e.status='active' ORDER BY (CASE WHEN e.uid LIKE 'emp56_%' THEN 1 ELSE 0 END), e.hire_date, e.id LIMIT 1 OFFSET 2) IS NOT NULL
ON CONFLICT (uid) DO NOTHING;

INSERT INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
SELECT 'apr56_'||dc.dept_code||'_pending', 'apf_leave_default',
    (SELECT e.uid FROM employees e WHERE e.department_uid=dc.dept_uid AND e.status='active' ORDER BY (CASE WHEN e.uid LIKE 'emp56_%' THEN 1 ELSE 0 END), e.hire_date, e.id LIMIT 1 OFFSET 3),
    1,1,'pending', NOW()-INTERVAL '2 days', NOW()-INTERVAL '2 days'
FROM (VALUES
    ('dept_hr','HR'),('dept_it','IT'),('dept_fin','FIN'),('dept_ops','OPS'),
    ('dept_sec','SEC'),('dept_admin','ADM'),('dept_acad','ACA'),('dept_lib','LIB')
) AS dc(dept_uid, dept_code)
WHERE (SELECT e.uid FROM employees e WHERE e.department_uid=dc.dept_uid AND e.status='active' ORDER BY (CASE WHEN e.uid LIKE 'emp56_%' THEN 1 ELSE 0 END), e.hire_date, e.id LIMIT 1 OFFSET 3) IS NOT NULL
ON CONFLICT (uid) DO NOTHING;

-- Fixed-date pending requests 2026-04-19 (065)
INSERT INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
SELECT s.approval_uid, 'apf_leave_default', s.emp_uid, 1, 1, 'pending', s.submitted_at::TIMESTAMP, s.submitted_at::TIMESTAMP
FROM (VALUES
    ('apr65_20260419_ops02','emp56_OPS_02','2026-04-19 10:15:00'),
    ('apr65_20260419_sec02','emp56_SEC_02','2026-04-19 10:32:00'),
    ('apr65_20260419_adm02','emp56_ADM_02','2026-04-19 11:05:00'),
    ('apr65_20260419_lib02','emp56_LIB_02','2026-04-19 11:20:00')
) AS s(approval_uid, emp_uid, submitted_at)
JOIN employees e ON e.uid = s.emp_uid
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Leave Requests (055, 056, 065)
-- ============================================================================
INSERT INTO leave_requests (uid, employee_uid, leave_type_uid, start_date, end_date, days, notes, submitted_at, decided_at, approval_request_uid, created_at, updated_at) VALUES
    ('lrq_case_pending_0000000000000000000001',   'emp_00000000000000000000000000000002','ltype_00000000000000000000000000000002','2026-05-10'::DATE,'2026-05-10'::DATE,1,'Pending approval case',   NOW()-INTERVAL '2 days',  NULL,                    'apr_case_pending_0000000000000000000001',   NOW()-INTERVAL '2 days',  NOW()-INTERVAL '2 days'),
    ('lrq_case_approved_0000000000000000000001',  'emp_00000000000000000000000000000003','ltype_00000000000000000000000000000002','2026-05-12'::DATE,'2026-05-13'::DATE,2,'Approved request case',   NOW()-INTERVAL '10 days', NOW()-INTERVAL '9 days', 'apr_case_approved_00000000000000000000001', NOW()-INTERVAL '10 days', NOW()-INTERVAL '9 days'),
    ('lrq_case_rejected_0000000000000000000001',  'emp_00000000000000000000000000000004','ltype_00000000000000000000000000000003','2026-05-14'::DATE,'2026-05-14'::DATE,1,'Rejected request case',   NOW()-INTERVAL '8 days',  NOW()-INTERVAL '7 days', 'apr_case_rejected_00000000000000000000001', NOW()-INTERVAL '8 days',  NOW()-INTERVAL '7 days'),
    ('lrq_case_cancelled_0000000000000000000001', 'emp_00000000000000000000000000000005','ltype_00000000000000000000000000000001','2026-05-15'::DATE,'2026-05-15'::DATE,1,'Cancelled request case',  NOW()-INTERVAL '6 days',  NOW()-INTERVAL '5 days', 'apr_case_cancelled_0000000000000000000001', NOW()-INTERVAL '6 days',  NOW()-INTERVAL '5 days'),
    ('lrq_case_pending_step2_000000000000000001', 'emp_00000000000000000000000000000006','ltype_00000000000000000000000000000002','2026-05-16'::DATE,'2026-05-16'::DATE,1,'Pending at step 2 case',  NOW()-INTERVAL '1 days',  NULL,                    'apr_case_pending_step2_000000000000000001', NOW()-INTERVAL '1 days',  NOW()-INTERVAL '1 days')
ON CONFLICT (uid) DO NOTHING;

-- 056 per-dept leave requests
INSERT INTO leave_requests (uid, employee_uid, leave_type_uid, start_date, end_date, days, notes, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
SELECT 'lrq56_'||SPLIT_PART(ar.uid,'_',2)||'_approved', ar.requester_uid,
    'ltype_00000000000000000000000000000002','2026-05-10'::DATE,'2026-05-10'::DATE,1,'Approved leave case',
    NOW()-INTERVAL '5 days', NOW()-INTERVAL '4 days', ar.uid, NOW()-INTERVAL '5 days', NOW()-INTERVAL '4 days'
FROM approval_requests ar WHERE ar.uid LIKE 'apr56_%_approved' ON CONFLICT (uid) DO NOTHING;

INSERT INTO leave_requests (uid, employee_uid, leave_type_uid, start_date, end_date, days, notes, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
SELECT 'lrq56_'||SPLIT_PART(ar.uid,'_',2)||'_rejected', ar.requester_uid,
    'ltype_00000000000000000000000000000003','2026-05-12'::DATE,'2026-05-12'::DATE,1,'Rejected leave case',
    NOW()-INTERVAL '4 days', NOW()-INTERVAL '3 days', ar.uid, NOW()-INTERVAL '4 days', NOW()-INTERVAL '3 days'
FROM approval_requests ar WHERE ar.uid LIKE 'apr56_%_rejected' ON CONFLICT (uid) DO NOTHING;

INSERT INTO leave_requests (uid, employee_uid, leave_type_uid, start_date, end_date, days, notes, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
SELECT 'lrq56_'||SPLIT_PART(ar.uid,'_',2)||'_pending', ar.requester_uid,
    'ltype_00000000000000000000000000000001','2026-05-14'::DATE,'2026-05-14'::DATE,1,'Pending leave case',
    NOW()-INTERVAL '2 days', NULL, ar.uid, NOW()-INTERVAL '2 days', NOW()-INTERVAL '2 days'
FROM approval_requests ar WHERE ar.uid LIKE 'apr56_%_pending' ON CONFLICT (uid) DO NOTHING;

-- 065 fixed-date leave requests
INSERT INTO leave_requests (uid, employee_uid, leave_type_uid, start_date, end_date, days, notes, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
SELECT s.leave_uid, s.emp_uid, lt.uid, s.start_date::DATE, s.end_date::DATE, s.days, s.notes,
    s.submitted_at::TIMESTAMP, NULL, s.approval_uid, s.submitted_at::TIMESTAMP, s.submitted_at::TIMESTAMP
FROM (VALUES
    ('lrq65_20260419_ops02','apr65_20260419_ops02','emp56_OPS_02','2026-04-23','2026-04-24',2,'Seeded pending request from OPS.','2026-04-19 10:15:00'),
    ('lrq65_20260419_sec02','apr65_20260419_sec02','emp56_SEC_02','2026-04-24','2026-04-24',1,'Seeded pending request from SEC.','2026-04-19 10:32:00'),
    ('lrq65_20260419_adm02','apr65_20260419_adm02','emp56_ADM_02','2026-04-27','2026-04-29',3,'Seeded pending request from ADM.','2026-04-19 11:05:00'),
    ('lrq65_20260419_lib02','apr65_20260419_lib02','emp56_LIB_02','2026-04-25','2026-04-26',2,'Seeded pending request from LIB.','2026-04-19 11:20:00')
) AS s(leave_uid, approval_uid, emp_uid, start_date, end_date, days, notes, submitted_at)
JOIN employees e ON e.uid = s.emp_uid
JOIN leave_types lt ON lt.code = 'ANNUAL'
JOIN approval_requests ar ON ar.uid = s.approval_uid
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Approval Actions (055, 056, 065)
-- ============================================================================
INSERT INTO approval_actions (uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at) VALUES
    ('apa_case_submit_pending_0000000000000001',  'apr_case_pending_0000000000000000000001',   'submit',  NULL,'emp_00000000000000000000000000000002','Submitted request',                NOW()-INTERVAL '2 days', NOW()-INTERVAL '2 days'),
    ('apa_case_submit_approved_0000000000000001', 'apr_case_approved_00000000000000000000001', 'submit',  NULL,'emp_00000000000000000000000000000003','Submitted approved flow request',  NOW()-INTERVAL '10 days',NOW()-INTERVAL '10 days'),
    ('apa_case_approve_000000000000000000000001', 'apr_case_approved_00000000000000000000001', 'approve', 1,   'emp_00000000000000000000000000000001','Approved by manager',              NOW()-INTERVAL '9 days', NOW()-INTERVAL '9 days'),
    ('apa_case_submit_rejected_0000000000000001', 'apr_case_rejected_00000000000000000000001', 'submit',  NULL,'emp_00000000000000000000000000000004','Submitted rejected flow request',  NOW()-INTERVAL '8 days', NOW()-INTERVAL '8 days'),
    ('apa_case_reject_0000000000000000000000001', 'apr_case_rejected_00000000000000000000001', 'reject',  1,   'emp_00000000000000000000000000000001','Rejected due to policy',           NOW()-INTERVAL '7 days', NOW()-INTERVAL '7 days'),
    ('apa_case_submit_cancel_000000000000000001', 'apr_case_cancelled_0000000000000000000001', 'submit',  NULL,'emp_00000000000000000000000000000005','Submitted then cancelled',         NOW()-INTERVAL '6 days', NOW()-INTERVAL '6 days'),
    ('apa_case_cancel_0000000000000000000000001', 'apr_case_cancelled_0000000000000000000001', 'cancel',  NULL,'emp_00000000000000000000000000000005','Requester cancelled before review',NOW()-INTERVAL '5 days', NOW()-INTERVAL '5 days'),
    ('apa_case_submit_step2_0000000000000000001', 'apr_case_pending_step2_000000000000000001', 'submit',  NULL,'emp_00000000000000000000000000000006','Submitted multi-step request',     NOW()-INTERVAL '1 days', NOW()-INTERVAL '1 days'),
    ('apa_case_approve_step1_000000000000000001', 'apr_case_pending_step2_000000000000000001', 'approve', 1,   'emp_00000000000000000000000000000001','Step 1 approved, waiting step 2',  NOW()-INTERVAL '20 hours',NOW()-INTERVAL '20 hours')
ON CONFLICT (uid) DO NOTHING;

-- 056 submit actions for all apr56_*
INSERT INTO approval_actions (uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at)
SELECT 'apa56_submit_'||ar.uid, ar.uid, 'submit', NULL, ar.requester_uid, 'Submitted by employee',
    CASE WHEN ar.status='approved' THEN NOW()-INTERVAL '5 days' WHEN ar.status='rejected' THEN NOW()-INTERVAL '4 days' ELSE NOW()-INTERVAL '2 days' END,
    CASE WHEN ar.status='approved' THEN NOW()-INTERVAL '5 days' WHEN ar.status='rejected' THEN NOW()-INTERVAL '4 days' ELSE NOW()-INTERVAL '2 days' END
FROM approval_requests ar WHERE ar.uid LIKE 'apr56_%' ON CONFLICT (uid) DO NOTHING;

-- 056 decide actions (approve/reject only)
INSERT INTO approval_actions (uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at)
SELECT 'apa56_decide_'||ar.uid, ar.uid,
    CASE WHEN ar.status='approved' THEN 'approve' ELSE 'reject' END, 1,
    COALESCE(
        (SELECT em.uid FROM user_roles ur JOIN roles r ON r.id=ur.role_id JOIN users u ON u.id=ur.user_id JOIN employees em ON em.uid=u.employee_uid
         WHERE r.uid='role_department_manager' AND ur.department_uid=req.department_uid LIMIT 1),
        req.uid
    ),
    CASE WHEN ar.status='approved' THEN 'Approved by department manager' ELSE 'Rejected by department manager' END,
    CASE WHEN ar.status='approved' THEN NOW()-INTERVAL '4 days' ELSE NOW()-INTERVAL '3 days' END,
    CASE WHEN ar.status='approved' THEN NOW()-INTERVAL '4 days' ELSE NOW()-INTERVAL '3 days' END
FROM approval_requests ar JOIN employees req ON req.uid=ar.requester_uid
WHERE ar.uid LIKE 'apr56_%' AND ar.status IN ('approved','rejected') ON CONFLICT (uid) DO NOTHING;

-- 065 submit actions
INSERT INTO approval_actions (uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at)
SELECT s.action_uid, s.approval_uid, 'submit', NULL, s.emp_uid, 'Seeded submit action for pending request (2026-04-19).', s.submitted_at::TIMESTAMP, s.submitted_at::TIMESTAMP
FROM (VALUES
    ('apa65_20260419_ops02','apr65_20260419_ops02','emp56_OPS_02','2026-04-19 10:15:00'),
    ('apa65_20260419_sec02','apr65_20260419_sec02','emp56_SEC_02','2026-04-19 10:32:00'),
    ('apa65_20260419_adm02','apr65_20260419_adm02','emp56_ADM_02','2026-04-19 11:05:00'),
    ('apa65_20260419_lib02','apr65_20260419_lib02','emp56_LIB_02','2026-04-19 11:20:00')
) AS s(action_uid, approval_uid, emp_uid, submitted_at)
JOIN employees e ON e.uid = s.emp_uid
JOIN approval_requests ar ON ar.uid = s.approval_uid
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Leave Records – case + 056 approved (055, 056, 065)
-- ============================================================================
-- Case: approved leave record (055)
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at)
SELECT 'lrec_case_from_approved_request_000000001',
    e.id, lt.id, '2026-05-12','2026-05-13', 2,
    NOW()-INTERVAL '9 days', rb.id, 'Generated from approved request case',
    'lrq_case_approved_0000000000000000000001', NOW()-INTERVAL '9 days', NOW()-INTERVAL '9 days'
FROM employees e
JOIN leave_types lt ON lt.uid='ltype_00000000000000000000000000000002'
JOIN employees rb ON rb.uid='emp_00000000000000000000000000000001'
WHERE e.uid='emp_00000000000000000000000000000003'
ON CONFLICT (uid) DO NOTHING;

-- 056 approved leave records
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at)
SELECT 'lrec56_'||SPLIT_PART(ar.uid,'_',2)||'_approved',
    e_req.id, lt.id, '2026-05-10','2026-05-10', 1,
    NOW()-INTERVAL '4 days',
    COALESCE(
        (SELECT em.id FROM user_roles ur JOIN roles r ON r.id=ur.role_id JOIN users u ON u.id=ur.user_id JOIN employees em ON em.uid=u.employee_uid
         WHERE r.uid='role_department_manager' AND ur.department_uid=e_req.department_uid LIMIT 1),
        e_req.id
    ),
    'Approved leave recorded by manager', lr.uid,
    NOW()-INTERVAL '4 days', NOW()-INTERVAL '4 days'
FROM approval_requests ar
JOIN leave_requests lr ON lr.approval_request_uid=ar.uid
JOIN employees e_req ON e_req.uid=ar.requester_uid
JOIN leave_types lt ON lt.uid=lr.leave_type_uid
WHERE ar.uid LIKE 'apr56_%_approved'
ON CONFLICT (uid) DO NOTHING;

-- 065 leave records on 2026-04-19
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at)
SELECT s.uid, e.id, lt.id, '2026-04-19','2026-04-19', 1, '2026-04-19 09:30:00', rb.id, s.notes, NULL, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM (VALUES
    ('lrec65_20260419_hr03', 'emp56_HR_03',  'Seeded leave on 19 Apr 2026 (HR).'),
    ('lrec65_20260419_it03', 'emp56_IT_03',  'Seeded leave on 19 Apr 2026 (IT).'),
    ('lrec65_20260419_fin03','emp56_FIN_03', 'Seeded leave on 19 Apr 2026 (FIN).')
) AS s(uid, emp_uid, notes)
JOIN employees e  ON e.uid  = s.emp_uid
JOIN leave_types lt ON lt.code = 'CASUAL'
LEFT JOIN employees rb ON rb.uid = 'emp56_HR_01'
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Leave Balance Transactions (055, 056)
-- ============================================================================
INSERT INTO leave_balance_transactions (uid, balance_id, transaction_type, days, leave_record_id, notes, created_by, created_at)
SELECT 'lbt_case_initial_00000000000000000000001', lb.id, 'INITIAL', 21, NULL, 'Initial yearly balance case', e.id, NOW()-INTERVAL '30 days'
FROM leave_balances lb JOIN employees e ON e.id=lb.employee_id JOIN leave_types lt ON lt.id=lb.leave_type_id
WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' AND lb.year=2026 ON CONFLICT (uid) DO NOTHING;

INSERT INTO leave_balance_transactions (uid, balance_id, transaction_type, days, leave_record_id, notes, created_by, created_at)
SELECT 'lbt_case_deduct_000000000000000000000001', lb.id, 'DEDUCT', 2, lr.id, 'Deduct after approved leave request case', e.id, NOW()-INTERVAL '9 days'
FROM leave_balances lb JOIN employees e ON e.id=lb.employee_id JOIN leave_types lt ON lt.id=lb.leave_type_id
LEFT JOIN leave_records lr ON lr.uid='lrec_case_from_approved_request_000000001'
WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' AND lb.year=2026 ON CONFLICT (uid) DO NOTHING;

INSERT INTO leave_balance_transactions (uid, balance_id, transaction_type, days, leave_record_id, notes, created_by, created_at)
SELECT 'lbt_case_refund_000000000000000000000001', lb.id, 'REFUND', 1, NULL, 'Refund after manual correction case', e.id, NOW()-INTERVAL '8 days'
FROM leave_balances lb JOIN employees e ON e.id=lb.employee_id JOIN leave_types lt ON lt.id=lb.leave_type_id
WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' AND lb.year=2026 ON CONFLICT (uid) DO NOTHING;

INSERT INTO leave_balance_transactions (uid, balance_id, transaction_type, days, leave_record_id, notes, created_by, created_at)
SELECT 'lbt_case_adjustment_000000000000000000001', lb.id, 'ADJUSTMENT', 1, NULL, 'Administrative adjustment case', e.id, NOW()-INTERVAL '7 days'
FROM leave_balances lb JOIN employees e ON e.id=lb.employee_id JOIN leave_types lt ON lt.id=lb.leave_type_id
WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' AND lb.year=2026 ON CONFLICT (uid) DO NOTHING;

INSERT INTO leave_balance_transactions (uid, balance_id, transaction_type, days, leave_record_id, notes, created_by, created_at)
SELECT 'lbt_case_carry_over_00000000000000000001', lb.id, 'CARRY_OVER', 5, NULL, 'Year carry-over case', e.id, NOW()-INTERVAL '6 days'
FROM leave_balances lb JOIN employees e ON e.id=lb.employee_id JOIN leave_types lt ON lt.id=lb.leave_type_id
WHERE e.uid='emp_00000000000000000000000000000001' AND lt.code='ANNUAL' AND lb.year=2026 ON CONFLICT (uid) DO NOTHING;

-- 056 deduct transactions for approved leaves
INSERT INTO leave_balance_transactions (uid, balance_id, transaction_type, days, leave_record_id, notes, created_by, created_at)
SELECT 'lbt56_'||SPLIT_PART(ar.uid,'_',2)||'_approved_deduct',
    lb.id, 'DEDUCT', 1, lrec.id, 'Deduct due to approved leave request',
    COALESCE(
        (SELECT em.id FROM user_roles ur JOIN roles r ON r.id=ur.role_id JOIN users u ON u.id=ur.user_id JOIN employees em ON em.uid=u.employee_uid
         WHERE r.uid='role_department_manager' AND ur.department_uid=e_req.department_uid LIMIT 1),
        e_req.id
    ),
    NOW()-INTERVAL '4 days'
FROM approval_requests ar
JOIN leave_requests lr  ON lr.approval_request_uid=ar.uid
JOIN employees e_req    ON e_req.uid=ar.requester_uid
JOIN leave_types lt     ON lt.uid=lr.leave_type_uid
JOIN leave_balances lb  ON lb.employee_id=e_req.id AND lb.leave_type_id=lt.id AND lb.year=2026
JOIN leave_records lrec ON lrec.leave_request_uid=lr.uid
WHERE ar.uid LIKE 'apr56_%_approved' ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- Sub Leave Types
-- (076_seed_sub_leave_types, 20260428001721, 20260428001722)
-- ============================================================================

-- Sick leave sub-types (slt_001-004) – 076
INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar) VALUES
    ('slt_00000000000000000000000000000001','ltype_00000000000000000000000000000003','3 months with 75% pay','3 أشهر بأجر 75%'),
    ('slt_00000000000000000000000000000002','ltype_00000000000000000000000000000003','6 months with 50% pay for employees under 50 years old','6 أشهر بأجر 50% لمن أقل من 50 عام'),
    ('slt_00000000000000000000000000000003','ltype_00000000000000000000000000000003','3 months with 75% pay for employees over 50 years old','3 أشهر بأجر 75% لمن تجاوز 50 عام'),
    ('slt_00000000000000000000000000000004','ltype_00000000000000000000000000000003','3 months with full pay','3 أشهر بأجر كامل')
ON CONFLICT DO NOTHING;

-- SPECIAL_UNPAID_EXT sub-types (slt_005-014) – 076
INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT slt.uid, lt.uid, slt.name_en, slt.name_ar
FROM leave_types lt
JOIN (VALUES
    ('slt_00000000000000000000000000000005','Leave to Care for a Son/Daughter with Special Needs','إجازة رعاية ابن/ابنه من ذوي الاحتياجات الخاصة'),
    ('slt_00000000000000000000000000000006','Leave to Care for a Sister','إجازة رعاية أخت'),
    ('slt_00000000000000000000000000000007','Leave to Care for a Grandson/Granddaughter','إجازة رعاية حفيد/حفيدة'),
    ('slt_00000000000000000000000000000008','Family Reunification Leave','إجازة لجمع شمل الأسرة'),
    ('slt_00000000000000000000000000000009','Leave to Care for a Father/Mother','إجازة رعاية والد/والدة'),
    ('slt_00000000000000000000000000000010','Leave to Attend to Family Affairs','إجازة رعاية مصالح الأسرة'),
    ('slt_00000000000000000000000000000011','Leave to Care for a Husband','إجازة رعاية زوج'),
    ('slt_00000000000000000000000000000012','Leave for Academic Research','إجازة لجمع مادة علمية'),
    ('slt_00000000000000000000000000000013','Leave to Visit Beirut Arab University','إجازة لزيارة جامعة بيروت العربية'),
    ('slt_00000000000000000000000000000014','Leave to Visit Family Abroad','إجازة لزيارة الأسرة بالخارج')
) AS slt(uid, name_en, name_ar) ON TRUE
WHERE lt.code = 'SPECIAL_UNPAID_EXT'
ON CONFLICT DO NOTHING;

-- SPECIAL_PAID_EXT sub-types (slt_015-036) – 076
INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT slt.uid, lt.uid, slt.name_en, slt.name_ar
FROM leave_types lt
JOIN (VALUES
    ('slt_00000000000000000000000000000015','Conducting Laboratory Experiments','إجراء تجارب معملية'),
    ('slt_00000000000000000000000000000016','Undergoing Medical Examinations','إجراء فحوصات طبية'),
    ('slt_00000000000000000000000000000017','Performing Umrah','أداء العمرة'),
    ('slt_00000000000000000000000000000018','Reserve Officer Call-Up','استدعاء ضابط احتياط'),
    ('slt_00000000000000000000000000000019','Supervision / Research / Competition','الإشراف / بحث / مسابقة'),
    ('slt_00000000000000000000000000000020','Travel Abroad','السفر للخارج'),
    ('slt_00000000000000000000000000000021','Participation in an Olympic Games','المشاركة في دورة ألعاب أوليمبية'),
    ('slt_00000000000000000000000000000022','Participation in a Medical Convoy','المشاركة في قافلة طبية'),
    ('slt_00000000000000000000000000000023','Exceeding the Leave Duration','تجاوز مدة الإجازة'),
    ('slt_00000000000000000000000000000024','Attending a Meeting Abroad','حضور اجتماع خارج البلاد'),
    ('slt_00000000000000000000000000000025','Attending a Training Course','حضور دورة تدريبية'),
    ('slt_00000000000000000000000000000026','Attending a Scientific Forum','حضور ملتقى علمي'),
    ('slt_00000000000000000000000000000027','Attending a Scientific Symposium','حضور منتدى علمي'),
    ('slt_00000000000000000000000000000028','Attending a Scientific Conference','حضور مؤتمر علمي'),
    ('slt_00000000000000000000000000000029','Attending a Scientific Seminar','حضور ندوة علمية'),
    ('slt_00000000000000000000000000000030','Attending a Workshop','حضور ورشة عمل'),
    ('slt_00000000000000000000000000000031','Visiting the Holy Places - Jerusalem','زيارة الأماكن المقدسة-القدس'),
    ('slt_00000000000000000000000000000032','Practical Training','التدريب العملى'),
    ('slt_00000000000000000000000000000033','Visiting Lecturer','مدرس زائر'),
    ('slt_00000000000000000000000000000034','Exceeding the Leave Duration','تجاوز مدة الإجازة'),
    ('slt_00000000000000000000000000000035','Travel Abroad','السفر للخارج'),
    ('slt_00000000000000000000000000000036','Teaching Abroad','التدريس بالخارج')
) AS slt(uid, name_en, name_ar) ON TRUE
WHERE lt.code = 'SPECIAL_PAID_EXT'
ON CONFLICT DO NOTHING;

-- SPECIAL_PAID sub-types (slt_037-041) – 20260428001721
INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT slt.uid, lt.uid, slt.name_en, slt.name_ar
FROM leave_types lt
JOIN (VALUES
    ('slt_00000000000000000000000000000037','Leave to Perform Hajj','أجازة لأداء فريضة الحج.'),
    ('slt_00000000000000000000000000000038','Leave for Contact with a Sick Person','أجازة مخالط مريض.'),
    ('slt_00000000000000000000000000000039','Leave for Work Injury','أجازة اصابة عمل.'),
    ('slt_00000000000000000000000000000040','Leave for Taking Examinations','أجازة إداء الامتحانات.'),
    ('slt_00000000000000000000000000000041','Other','أخرى.')
) AS slt(uid, name_en, name_ar) ON TRUE
WHERE lt.code = 'SPECIAL_PAID'
  AND NOT EXISTS (SELECT 1 FROM sub_leave_types s WHERE s.leave_type_uid=lt.uid AND s.name_ar=slt.name_ar)
ON CONFLICT DO NOTHING;


-- SPECIAL_UNPAID sub-types (slt_042-055) – 20260428001722
INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT slt.uid, lt.uid, slt.name_en, slt.name_ar
FROM leave_types lt
JOIN (VALUES
    ('slt_00000000000000000000000000000042','Accompanying a Sick Relative','مرافقة مريض من الأقارب'),
    ('slt_00000000000000000000000000000043','Accompanying a Sick Spouse','مرافقة الزوج المريض'),
    ('slt_00000000000000000000000000000044','Accompanying a Sick Child','مرافقة الطفل المريض'),
    ('slt_00000000000000000000000000000045','Accompanying a Sick Parent','مرافقة الوالد المريض'),
    ('slt_00000000000000000000000000000046','Accompanying a Sick Sibling','مرافقة الأخ المريض'),
    ('slt_00000000000000000000000000000047','Accompanying a Sick Grandparent','مرافقة الجد المريض'),
    ('slt_00000000000000000000000000000048','Accompanying a Sick Grandchild','مرافقة الحفيد المريض'),
    ('slt_00000000000000000000000000000049','Accompanying a Sick Uncle/Aunt','مرافقة العم/الخال المريض'),
    ('slt_00000000000000000000000000000050','Accompanying a Sick Nephew/Niece','مرافقة ابن الأخ/الأخت المريض'),
    ('slt_00000000000000000000000000000051','Accompanying a Sick Cousin','مرافقة ابن العم/الخال المريض'),
    ('slt_00000000000000000000000000000052','Accompanying a Sick In-Law','مرافقة الحماة/الحمو المريض'),
    ('slt_00000000000000000000000000000053','Accompanying a Sick Brother/Sister-in-Law','مرافقة الصهر/الزوجة المريض'),
    ('slt_00000000000000000000000000000054','Accompanying a Sick Son/Daughter-in-Law','مرافقة الكنة/الصهر المريض'),
    ('slt_00000000000000000000000000000055','Other','أخرى')
) AS slt(uid, name_en, name_ar) ON TRUE
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (SELECT 1 FROM sub_leave_types s WHERE s.leave_type_uid=lt.uid AND s.name_ar=slt.name_ar)
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Shifts
-- (051_add_shifts_and_assignments)
-- ============================================================================
INSERT INTO shifts (uid, start_time, end_time, grace_minutes, created_at, updated_at) VALUES
    ('shf_general_seed', '08:00', '16:00', 15, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (uid) DO NOTHING;

-- ============================================================================
-- HR Staff Role
-- (20260508123000_create_hr_staff_role_for_profile_changes)
-- ============================================================================
INSERT INTO roles (uid, name, description, scope_type, is_system, created_at, updated_at)
VALUES (
    'role_hr_staff',
    'HR Staff',
    'HR staff with employee visibility and profile change submission access',
    'global',
    FALSE,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
) ON CONFLICT (uid) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code IN (
    'employees:read',
    'employee-profile-changes:read',
    'employee-profile-changes:write'
)
WHERE r.uid = 'role_hr_staff'
ON CONFLICT DO NOTHING;

-- Clean up profile change permissions from other roles (they should use HR Staff role)
DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id
    FROM roles
    WHERE uid IN ('role_university_human_resources', 'role_hr_manager', 'role_department_manager')
)
  AND permission_id IN (
    SELECT id
    FROM permissions
    WHERE code IN ('employee-profile-changes:read', 'employee-profile-changes:write')
);

-- ============================================================================
-- University Leadership Roles (Restored)
-- (20260510130000_restore_leadership_roles, 20260511090000_seed_university_leadership_roles)
-- ============================================================================
INSERT INTO roles (uid, name, description, is_system, created_at, updated_at)
VALUES
    ('role_university_president', 'University President', 'University President role', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('role_vice_president',       'Vice President',       'Vice President role',       TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (uid) DO NOTHING;

-- Grant department read access so leaders can view the org chart
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code IN ('departments:read')
WHERE r.uid IN ('role_university_president', 'role_vice_president', 'role_dean')
ON CONFLICT DO NOTHING;

-- ============================================================================
-- University Leadership People
-- (20260511100000_seed_university_leadership_people)
-- ============================================================================
INSERT INTO employees (
    uid, name, mobile, government_id, university_id, email,
    hire_date, status, department_uid, shift_uid, created_at, updated_at
)
VALUES
    (
        'emp_university_president_20260508', 'احمد عادل عبد الحكيم',
        '+201000000074', '299000000074', 'U20260508PRES',
        'president.seed@university.edu.eg',
        '2026-05-08', 'active', NULL, 'shf_general_seed',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    ),
    (
        'emp_vice_president_20260508', 'عفاف العوفي',
        '+201000000075', '299000000075', 'U20260508VP',
        'vice.president.seed@university.edu.eg',
        '2026-05-08', 'active', NULL, 'shf_general_seed',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    ),
    (
        'emp_dean_engineering_20260508', 'وائل المغلاني',
        '+201000000076', '299000000076', 'U20260508DEANENG',
        'dean.engineering.seed@university.edu.eg',
        '2026-05-08', 'active', NULL, 'shf_general_seed',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    )
ON CONFLICT (uid) DO NOTHING;

INSERT INTO users (uid, phone, employee_uid, is_active, created_at, updated_at)
VALUES
    ('usr_university_president_20260508', '+201000000074', 'emp_university_president_20260508', TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('usr_vice_president_20260508',       '+201000000075', 'emp_vice_president_20260508',       TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('usr_dean_engineering_20260508',     '+201000000076', 'emp_dean_engineering_20260508',     TRUE, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
ON CONFLICT (phone) DO NOTHING;

-- Assign base employee role
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP
FROM users u, roles r
WHERE u.uid IN ('usr_university_president_20260508', 'usr_vice_president_20260508', 'usr_dean_engineering_20260508')
  AND r.uid = 'role_employee'
ON CONFLICT DO NOTHING;

-- Assign leadership roles
INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP FROM users u, roles r
WHERE u.uid = 'usr_university_president_20260508' AND r.uid = 'role_university_president'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP FROM users u, roles r
WHERE u.uid = 'usr_vice_president_20260508' AND r.uid = 'role_vice_president'
ON CONFLICT DO NOTHING;

INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP FROM users u, roles r
WHERE u.uid = 'usr_dean_engineering_20260508' AND r.uid = 'role_dean'
ON CONFLICT DO NOTHING;

-- ============================================================================
-- Leadership Manager Relationships
-- (20260511110000_seed_leadership_manager_relationships)
-- ============================================================================
-- Wire the three seeded leaders into a reporting chain
UPDATE employees SET manager_uid = 'emp_university_president_20260508'
    WHERE uid = 'emp_vice_president_20260508';

UPDATE employees SET manager_uid = 'emp_vice_president_20260508'
    WHERE uid = 'emp_dean_engineering_20260508';

-- Bootstrap existing department managers → dean
UPDATE employees
SET manager_uid = 'emp_dean_engineering_20260508'
WHERE uid IN (
    SELECT e.uid FROM employees e
    JOIN users u ON u.employee_uid = e.uid
    JOIN user_roles ur ON ur.user_id = u.id
    JOIN roles r ON r.id = ur.role_id AND r.uid = 'role_department_manager'
);

-- Bootstrap regular employees → their department manager
UPDATE employees
SET manager_uid = (
    SELECT mgr_emp.uid FROM users mgr_u
    JOIN user_roles ur ON ur.user_id = mgr_u.id
    JOIN roles r ON r.id = ur.role_id AND r.uid = 'role_department_manager'
    JOIN employees mgr_emp ON mgr_emp.uid = mgr_u.employee_uid
    WHERE ur.department_uid = employees.department_uid
    LIMIT 1
)
WHERE manager_uid IS NULL
    AND department_uid IS NOT NULL
    AND uid NOT IN (
        SELECT e.uid FROM employees e
        JOIN users u ON u.employee_uid = e.uid
        JOIN user_roles ur ON ur.user_id = u.id
        JOIN roles r ON r.id = ur.role_id AND r.uid = 'role_department_manager'
    );

-- ============================================================================
-- END OF CONSOLIDATED SEED DATA
-- ============================================================================
