-- Restore the HR Manager role and remove the transferred permissions from University Human Resources.
INSERT OR IGNORE INTO roles (uid, name, description, scope_type, is_system, created_at, updated_at)
VALUES ('role_hr_manager', 'HR Manager', 'HR manager with global holiday management access', 'global', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

INSERT OR IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code IN (
    'approval:read',
    'approval:write',
    'leave:approve',
    'leave:read',
    'documents:read',
    'leave-types:read',
    'leave-types:write',
    'attendance-devices:read',
    'attendance-devices:write',
    'attendance:read',
    'attendance:write',
    'holidays:write'
)
WHERE r.uid = 'role_hr_manager';

DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_university_human_resources')
  AND permission_id IN (
      SELECT id FROM permissions
      WHERE code IN (
          'approval:read',
          'approval:write',
          'leave:approve',
          'leave:read',
          'documents:read',
          'leave-types:read',
          'leave-types:write',
          'attendance-devices:read',
          'attendance-devices:write',
          'attendance:read',
          'attendance:write',
          'holidays:write'
      )
  );
