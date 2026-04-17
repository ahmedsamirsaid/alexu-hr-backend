-- Grant leave approval permission to Department Manager role.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE uid = 'role_department_manager'),
    id
FROM permissions
WHERE code = 'leave:approve';

-- Scope legacy seeded department manager role assignment to SEC department.
UPDATE user_roles
SET department_uid = 'dept_sec'
WHERE user_id = (SELECT id FROM users WHERE uid = 'usr_manager_seed_000000000000000000')
  AND role_id = (SELECT id FROM roles WHERE uid = 'role_department_manager')
  AND department_uid IS NULL
  AND EXISTS (SELECT 1 FROM departments WHERE uid = 'dept_sec');
