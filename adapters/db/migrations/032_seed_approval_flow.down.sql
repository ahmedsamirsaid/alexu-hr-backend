DELETE FROM user_roles
WHERE user_id = (SELECT id FROM users WHERE uid = 'usr_manager_seed_000000000000000000')
  AND role_id = (SELECT id FROM roles WHERE uid = 'role_department_manager');

DELETE FROM approval_flow_steps WHERE uid = 'afs_leave_step1';
DELETE FROM approval_flows WHERE uid = 'apf_leave_default';
DELETE FROM roles WHERE uid = 'role_department_manager';
