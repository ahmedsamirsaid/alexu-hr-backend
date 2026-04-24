-- Remove holiday write permission from HR Manager.
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_hr_manager')
  AND permission_id = (SELECT id FROM permissions WHERE code = 'holidays:write');

-- Remove any assignments for the seeded HR Manager role.
DELETE FROM user_roles
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_hr_manager');

-- Remove seeded HR Manager role.
DELETE FROM roles
WHERE uid = 'role_hr_manager';
