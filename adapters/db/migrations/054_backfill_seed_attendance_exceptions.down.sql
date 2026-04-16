DELETE FROM attendance_exceptions
WHERE uid IN (
    'aex_seed_20260413_emp02_late',
    'aex_seed_20260414_emp07_early',
    'aex_seed_20260415_emp02_late',
    'aex_seed_20260415_emp04_miss_out',
    'aex_seed_20260415_emp07_early'
)
OR uid LIKE 'aex_seed_abs_%';

UPDATE employees
SET status = 'inactive'
WHERE uid = 'emp_00000000000000000000000000000011';
