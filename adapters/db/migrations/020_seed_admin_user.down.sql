DELETE FROM user_roles WHERE user_id = (SELECT id FROM users WHERE uid = 'usr_admin_seed_00000000000000000000');
DELETE FROM user_roles WHERE user_id = (SELECT id FROM users WHERE uid = 'usr_manager_seed_000000000000000000');
DELETE FROM users WHERE uid = 'usr_admin_seed_00000000000000000000';
DELETE FROM users WHERE uid = 'usr_manager_seed_000000000000000000';
