-- Seed permissions for department management
INSERT OR IGNORE INTO permissions (uid, code, description) VALUES
    ('perm_departments_read', 'departments:read', 'View departments'),
    ('perm_departments_write', 'departments:write', 'Create/update departments and assign managers');

-- Add new permissions to admin role
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'Admin'),
    id
FROM permissions
WHERE code IN ('departments:read', 'departments:write');
