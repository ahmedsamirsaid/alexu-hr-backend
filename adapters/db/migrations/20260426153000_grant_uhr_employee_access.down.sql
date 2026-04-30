-- Roll back the University Human Resources employee access backfill.
DELETE FROM user_roles
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_employee')
  AND user_id IN (
      SELECT ur.user_id
      FROM user_roles ur
      JOIN roles r ON r.id = ur.role_id
      WHERE r.uid = 'role_university_human_resources'
  );

DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_university_human_resources')
  AND permission_id = (SELECT id FROM permissions WHERE code = 'departments:read');

DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_employee')
  AND permission_id = (SELECT id FROM permissions WHERE code = 'holidays:read');
