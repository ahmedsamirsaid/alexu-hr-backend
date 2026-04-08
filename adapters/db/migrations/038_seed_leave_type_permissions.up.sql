-- Seed permissions for leave type management
INSERT OR IGNORE INTO permissions (uid, code, description) VALUES
    ('perm_leave_types_read', 'leave-types:read', 'View leave types'),
    ('perm_leave_types_write', 'leave-types:write', 'Enable/disable leave types');

-- Add new permissions to Admin role
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'Admin'),
    id
FROM permissions
WHERE code IN ('leave-types:read', 'leave-types:write');

-- Add new permissions to HR Manager role
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'HR Manager'),
    id
FROM permissions
WHERE code IN ('leave-types:read', 'leave-types:write');
