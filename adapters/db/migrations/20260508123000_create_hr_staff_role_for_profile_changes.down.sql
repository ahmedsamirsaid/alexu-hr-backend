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

DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_hr_staff')
  AND permission_id IN (
    SELECT id
    FROM permissions
    WHERE code IN ('employees:read', 'employee-profile-changes:read', 'employee-profile-changes:write')
);

DELETE FROM roles WHERE uid = 'role_hr_staff';
