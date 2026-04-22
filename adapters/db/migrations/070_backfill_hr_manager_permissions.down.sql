-- Roll back HR Manager backfilled permissions.
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_hr_manager')
  AND permission_id IN (
      SELECT id
      FROM permissions
      WHERE code IN (
          'leave-types:read',
          'leave-types:write',
          'attendance-devices:read',
          'attendance-devices:write',
          'attendance:read',
          'attendance:write'
      )
  );
