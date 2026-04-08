-- Delete leave records
DELETE FROM leave_records WHERE uid LIKE 'lrec_%';

-- Delete leave balances
DELETE FROM leave_balances WHERE uid LIKE 'lbal_%';

-- Delete employees
DELETE FROM employees WHERE uid LIKE 'emp_%';

-- Delete additional leave types (keep CASUAL from 009)
DELETE FROM leave_types WHERE code IN ('ANNUAL', 'SICK');
