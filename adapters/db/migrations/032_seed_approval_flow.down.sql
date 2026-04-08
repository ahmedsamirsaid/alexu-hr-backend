DELETE FROM approval_flow_steps WHERE uid = 'afs_leave_step1';
DELETE FROM approval_flows WHERE uid = 'apf_leave_default';
DELETE FROM roles WHERE uid = 'role_department_manager';
