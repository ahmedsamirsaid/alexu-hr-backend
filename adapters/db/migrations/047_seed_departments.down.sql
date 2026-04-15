DELETE FROM departments
WHERE uid IN (
    'dept_hr',
    'dept_it',
    'dept_fin',
    'dept_ops',
    'dept_sec',
    'dept_admin',
    'dept_acad',
    'dept_lib'
);
