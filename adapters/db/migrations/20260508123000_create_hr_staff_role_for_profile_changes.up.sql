INSERT OR IGNORE INTO roles (uid, name, description, scope_type, is_system, created_at, updated_at)
VALUES (
    'role_hr_staff',
    'HR Staff',
    'HR staff with employee visibility and profile change submission access',
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
    'employee-profile-changes:read',
    'employee-profile-changes:write'
)
WHERE r.uid = 'role_hr_staff';

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
