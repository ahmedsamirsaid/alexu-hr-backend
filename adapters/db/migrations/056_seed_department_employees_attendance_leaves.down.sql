-- Rollback for 056_seed_department_employees_attendance_leaves.up.sql

DELETE FROM leave_balance_transactions
WHERE uid LIKE 'lbt56_%';

DELETE FROM leave_records
WHERE uid LIKE 'lrec56_%';

DELETE FROM approval_actions
WHERE uid LIKE 'apa56_%';

DELETE FROM leave_requests
WHERE uid LIKE 'lrq56_%';

DELETE FROM approval_requests
WHERE uid LIKE 'apr56_%';

DELETE FROM leave_balances
WHERE uid LIKE 'lbal56_%';

DELETE FROM attendance_exceptions
WHERE uid LIKE 'aex56_%';

DELETE FROM attendance_records
WHERE uid LIKE 'atr56_%';

DELETE FROM attendance_devices
WHERE uid LIKE 'adev56_%';

DELETE FROM user_roles
WHERE user_id IN (SELECT id FROM users WHERE uid LIKE 'usr56_%');

DELETE FROM users
WHERE uid LIKE 'usr56_%';

DELETE FROM employees
WHERE uid LIKE 'emp56_%';
