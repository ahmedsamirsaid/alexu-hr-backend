DELETE FROM role_permissions WHERE role_id = (SELECT id FROM roles WHERE name = 'user');
DELETE FROM roles WHERE name = 'user';
