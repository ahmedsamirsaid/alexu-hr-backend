-- Revert leave approval permission grant from Department Manager role.
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_department_manager')
  AND permission_id = (SELECT id FROM permissions WHERE code = 'leave:approve');

-- Revert legacy seeded manager role assignment back to global scope.
UPDATE user_roles
SET department_uid = NULL
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_department_manager')
  AND department_uid = 'dept_sec'
  AND user_id IN (
      SELECT id FROM users
      WHERE uid = 'usr_manager_seed_000000000000000000'
         OR phone = '+201000000001'
  );
