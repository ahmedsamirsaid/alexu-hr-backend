-- The initial drop-leadership-positions migration accidentally deleted these
-- roles. Re-add them so their names can be displayed on the org chart cards.
-- The org chart hierarchy is determined by manager_uid, NOT by these roles.

INSERT OR IGNORE INTO roles (uid, name, description, is_system, created_at, updated_at)
VALUES
    ('role_university_president', 'University President', 'University President role', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('role_vice_president',       'Vice President',       'Vice President role',       1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Re-assign to the seeded leadership users
INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP
FROM users u, roles r
WHERE u.uid = 'usr_university_president_20260508' AND r.uid = 'role_university_president';

INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP
FROM users u, roles r
WHERE u.uid = 'usr_vice_president_20260508' AND r.uid = 'role_vice_president';
