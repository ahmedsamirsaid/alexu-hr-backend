DELETE FROM user_roles WHERE role_id IN (
    SELECT id FROM roles WHERE uid IN ('role_university_president', 'role_vice_president', 'role_dean')
);
DELETE FROM role_permissions WHERE role_id IN (
    SELECT id FROM roles WHERE uid IN ('role_university_president', 'role_vice_president', 'role_dean')
);
DELETE FROM roles WHERE uid IN ('role_university_president', 'role_vice_president', 'role_dean');
