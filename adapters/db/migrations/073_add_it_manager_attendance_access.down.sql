-- Remove attendance permissions from IT Manager.
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_it_manager')
  AND permission_id IN (
      SELECT id
      FROM permissions
      WHERE code IN (
          'attendance:read',
          'attendance:write',
          'attendance-devices:read',
          'employees:read',
          'departments:read'
      )
  );

-- Remove any assignments for the seeded IT Manager role.
DELETE FROM user_roles
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_it_manager');

-- Remove seeded IT Manager role.
DELETE FROM roles
WHERE uid = 'role_it_manager';
