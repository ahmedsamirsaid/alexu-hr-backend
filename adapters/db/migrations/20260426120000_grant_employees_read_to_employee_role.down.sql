-- Revoke Employee role read access to employee data granted by migration 072.
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_employee')
  AND permission_id = (SELECT id FROM permissions WHERE code = 'employees:read');
