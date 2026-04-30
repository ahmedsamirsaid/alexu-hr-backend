DELETE FROM user_roles
WHERE role_id IN (
    SELECT id
    FROM roles
    WHERE uid IN ('role_dean', 'role_university_human_resources')
);

DELETE FROM users
WHERE uid IN ('usr_seed_dean_20260423', 'usr_seed_university_hr_20260423');

DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id
    FROM roles
    WHERE uid IN ('role_dean', 'role_university_human_resources')
);

DELETE FROM roles
WHERE uid IN ('role_dean', 'role_university_human_resources');
