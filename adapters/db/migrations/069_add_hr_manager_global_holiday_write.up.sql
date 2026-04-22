-- Ensure HR Manager role exists as global scope.
INSERT OR IGNORE INTO roles (uid, name, description, scope_type, is_system, created_at, updated_at)
VALUES ('role_hr_manager', 'HR Manager', 'HR manager with global holiday management access', 'global', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Enforce global scope for HR Manager role.
UPDATE roles
SET scope_type = 'global',
    updated_at = CURRENT_TIMESTAMP
WHERE name = 'HR Manager';

-- Grant holiday write permission to HR Manager.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code = 'holidays:write'
WHERE r.name = 'HR Manager';
