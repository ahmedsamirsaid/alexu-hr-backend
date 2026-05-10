-- Ensure IT Manager role exists as global scope.
INSERT INTO roles (uid, name, description, scope_type, is_system, created_at, updated_at)
VALUES (
    'role_it_manager',
    'IT Manager',
    'IT manager with authority to manually correct attendance logs',
    'global',
    0,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Enforce global scope for IT Manager role.
UPDATE roles
SET scope_type = 'global',
    updated_at = CURRENT_TIMESTAMP
WHERE uid = 'role_it_manager' OR name = 'IT Manager';

-- Grant the read/write permissions needed for the attendance admin workflow.
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN (
    'attendance:read',
    'attendance:write',
    'attendance-devices:read',
    'employees:read',
    'departments:read'
)
WHERE r.uid = 'role_it_manager' OR r.name = 'IT Manager';
