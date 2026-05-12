DELETE FROM user_roles WHERE user_id IN (
    SELECT id FROM users WHERE uid IN (
        'usr_university_president_20260508',
        'usr_vice_president_20260508',
        'usr_dean_engineering_20260508'
    )
);
DELETE FROM users WHERE uid IN (
    'usr_university_president_20260508',
    'usr_vice_president_20260508',
    'usr_dean_engineering_20260508'
);
DELETE FROM employees WHERE uid IN (
    'emp_university_president_20260508',
    'emp_vice_president_20260508',
    'emp_dean_engineering_20260508'
);
