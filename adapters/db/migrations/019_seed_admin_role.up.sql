-- Create admin role as system role
INSERT INTO roles (uid, name, description, is_system) VALUES
    ('role_admin', 'Admin', 'System administrator with all permissions', 1);

-- Assign all permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'Admin'),
    id
FROM permissions;
