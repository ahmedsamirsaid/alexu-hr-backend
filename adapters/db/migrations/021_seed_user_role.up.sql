-- Create a default 'user' role for regular employees
INSERT INTO roles (uid, name, description, is_system, created_at, updated_at) VALUES
    ('role_user_00000000000000000000000', 'user', 'Regular user role for employees', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Assign basic read permissions to user role
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT 
    (SELECT id FROM roles WHERE name = 'user'),
    id,
    CURRENT_TIMESTAMP
FROM permissions
WHERE code IN ('employees:read', 'leave:read');
