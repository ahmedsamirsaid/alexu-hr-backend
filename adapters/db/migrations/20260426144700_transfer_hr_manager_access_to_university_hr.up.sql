-- Transfer HR Manager access to University Human Resources and remove the HR Manager role.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code IN (
    'approval:read',
    'approval:write',
    'leave:approve',
    'leave:read',
    'documents:read',
    'leave-types:read',
    'leave-types:write',
    'attendance-devices:read',
    'attendance-devices:write',
    'attendance:read',
    'attendance:write',
    'holidays:write'
)
WHERE r.uid = 'role_university_human_resources';

INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT ur.user_id, uhr.id, ur.department_uid, CURRENT_TIMESTAMP
FROM user_roles ur
JOIN roles hr ON hr.id = ur.role_id AND hr.uid = 'role_hr_manager'
JOIN roles uhr ON uhr.uid = 'role_university_human_resources';

DELETE FROM user_roles
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_hr_manager');

DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_hr_manager');

DELETE FROM roles
WHERE uid = 'role_hr_manager';
