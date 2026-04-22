-- Backfill HR Manager permissions that were seeded before the role existed.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN (
    'leave-types:read',
    'leave-types:write',
    'attendance-devices:read',
    'attendance-devices:write',
    'attendance:read',
    'attendance:write'
)
WHERE r.uid = 'role_hr_manager';
