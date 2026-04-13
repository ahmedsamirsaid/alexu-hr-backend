INSERT OR IGNORE INTO permissions (uid, code, description) VALUES
    ('perm_attendance_devices_read', 'attendance-devices:read', 'View attendance devices (ZKTeco)'),
    ('perm_attendance_devices_write', 'attendance-devices:write', 'Register/remove attendance devices (ZKTeco)');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN ('attendance-devices:read', 'attendance-devices:write')
WHERE r.name IN ('Admin', 'HR Manager');
