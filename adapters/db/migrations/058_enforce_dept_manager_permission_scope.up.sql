-- Enforce Department Manager leave approval permission even if prior migration was marked applied without data change.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code = 'leave:approve'
WHERE r.uid = 'role_department_manager';

-- Scope legacy seeded manager role assignment to SEC department when present.
UPDATE user_roles
SET department_uid = 'dept_sec'
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_department_manager')
  AND department_uid IS NULL
  AND user_id IN (
      SELECT id FROM users
      WHERE uid = 'usr_manager_seed_000000000000000000'
         OR phone = '+201000000001'
  )
  AND EXISTS (SELECT 1 FROM departments WHERE uid = 'dept_sec');
