INSERT OR IGNORE INTO permissions (uid, code, description) VALUES
    ('perm_attendance_read', 'attendance:read', 'View attendance records and summaries'),
    ('perm_attendance_write', 'attendance:write', 'Manage attendance work hours');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'Admin'),
    id
FROM permissions
WHERE code IN ('attendance:read', 'attendance:write');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'HR Manager'),
    id
FROM permissions
WHERE code IN ('attendance:read', 'attendance:write');

INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'Department Manager'),
    id
FROM permissions
WHERE code IN ('attendance:read', 'attendance:write');
