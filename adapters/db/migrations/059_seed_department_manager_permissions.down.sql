-- Revoke Department Manager permissions granted by 059 migration.
-- Note: leave:approve may have been granted by earlier migrations; this rollback removes it as part of this bundle.
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_department_manager')
  AND permission_id IN (
      SELECT id FROM permissions
      WHERE code IN (
          'employees:read',
          'departments:read',
          'attendance:read',
          'leave:read',
          'leave:request',
          'leave:approve'
      )
  );
