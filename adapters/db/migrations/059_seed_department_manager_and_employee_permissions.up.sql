-- Ensure core permissions used by Department Manager workflows exist.
INSERT OR IGNORE INTO permissions (uid, code, description) VALUES
    ('perm_employees_read', 'employees:read', 'View employee list and details'),
    ('perm_departments_read', 'departments:read', 'View departments'),
    ('perm_attendance_read', 'attendance:read', 'View attendance records and summaries'),
    ('perm_leave_read', 'leave:read', 'View leave records'),
    ('perm_leave_request', 'leave:request', 'Submit leave requests'),
    ('perm_leave_approve', 'leave:approve', 'Approve/reject leave requests');

-- Grant Department Manager the permissions needed for scoped team management.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN (
    'employees:read',
    'departments:read',
    'attendance:read',
    'leave:read',
    'leave:request',
    'leave:approve',
    'attendance-devices:read'
)
WHERE r.uid = 'role_department_manager';


-- Grant Employee role the permissions needed for self-service access.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN (
    'attendance:read',
    'leave:read',
    'leave:request'
)
WHERE r.uid = 'role_employee';