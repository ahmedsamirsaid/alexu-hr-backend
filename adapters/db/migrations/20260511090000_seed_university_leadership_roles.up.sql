-- Seed University President, Vice President, and Dean roles.
INSERT OR IGNORE INTO roles (uid, name, description, is_system, created_at, updated_at)
VALUES
    ('role_university_president', 'University President', 'University President role', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('role_vice_president',       'Vice President',       'Vice President role',       1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('role_dean',                 'Dean',                 'Dean role',                 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Grant department read access so leaders can view the org chart.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code IN ('departments:read')
WHERE r.uid IN ('role_university_president', 'role_vice_president', 'role_dean');
