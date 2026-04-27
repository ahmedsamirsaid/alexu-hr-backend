-- Create Employee role as system role
INSERT INTO roles (uid, name, description, is_system) VALUES
    ('role_employee', 'Employee', 'Default role for all employees with basic self-service permissions', 1);

-- Assign leave and holiday permissions to Employee role
INSERT INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'Employee'),
    id
FROM permissions
WHERE code IN ('leave:request', 'leave:read', 'holidays:read');
