-- Seed Dean and University Human Resources roles.
INSERT OR IGNORE INTO roles (uid, name, description, is_system, created_at, updated_at)
VALUES
    ('role_dean', 'Dean', 'Dean role with approval and leave visibility access', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('role_university_human_resources', 'University Human Resources', 'University HR role with approval and leave visibility access', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Grant approval, document-read, and leave-read permissions to both roles.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code IN (
    'approval:read',
    'approval:write',
    'leave:approve',
    'leave:read',
    'documents:read'
)
WHERE r.uid IN ('role_dean', 'role_university_human_resources');

-- Create fallback users only when fewer than two unassigned existing users are available.
INSERT OR IGNORE INTO users (uid, phone, is_active, created_at, updated_at)
SELECT
    'usr_seed_dean_20260423',
    '+201000000072',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
WHERE (
    SELECT COUNT(*)
    FROM users u
    WHERE u.uid NOT IN ('usr_seed_dean_20260423', 'usr_seed_university_hr_20260423')
      AND NOT EXISTS (
          SELECT 1
          FROM user_roles ur
          WHERE ur.user_id = u.id
      )
) = 0;

INSERT OR IGNORE INTO users (uid, phone, is_active, created_at, updated_at)
SELECT
    'usr_seed_university_hr_20260423',
    '+201000000073',
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
WHERE (
    SELECT COUNT(*)
    FROM users u
    WHERE u.uid NOT IN ('usr_seed_dean_20260423', 'usr_seed_university_hr_20260423')
      AND NOT EXISTS (
          SELECT 1
          FROM user_roles ur
          WHERE ur.user_id = u.id
      )
) <= 1;

-- Assign Dean to the first existing unassigned user, or to the fallback user when none exists.
INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT chosen.user_id, r.id, NULL, CURRENT_TIMESTAMP
FROM roles r
CROSS JOIN (
    SELECT COALESCE(
        (
            SELECT u.id
            FROM users u
            WHERE u.uid NOT IN ('usr_seed_dean_20260423', 'usr_seed_university_hr_20260423')
              AND NOT EXISTS (
                  SELECT 1
                  FROM user_roles ur
                  WHERE ur.user_id = u.id
              )
            ORDER BY u.created_at, u.id
            LIMIT 1
        ),
        (
            SELECT u.id
            FROM users u
            WHERE u.uid = 'usr_seed_dean_20260423'
        )
    ) AS user_id
) chosen
WHERE r.uid = 'role_dean'
  AND chosen.user_id IS NOT NULL;

-- Assign University Human Resources to the second existing unassigned user, or to the fallback user when needed.
INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT chosen.user_id, r.id, NULL, CURRENT_TIMESTAMP
FROM roles r
CROSS JOIN (
    SELECT COALESCE(
        (
            SELECT u.id
            FROM users u
            WHERE u.uid NOT IN ('usr_seed_dean_20260423', 'usr_seed_university_hr_20260423')
              AND NOT EXISTS (
                  SELECT 1
                  FROM user_roles ur
                  WHERE ur.user_id = u.id
              )
            ORDER BY u.created_at, u.id
            LIMIT 1 OFFSET 1
        ),
        (
            SELECT u.id
            FROM users u
            WHERE u.uid = 'usr_seed_university_hr_20260423'
        )
    ) AS user_id
) chosen
WHERE r.uid = 'role_university_human_resources'
  AND chosen.user_id IS NOT NULL;
