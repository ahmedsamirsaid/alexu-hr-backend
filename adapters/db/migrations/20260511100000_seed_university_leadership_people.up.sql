-- Seed the university leadership employees and user accounts used by the org chart.

INSERT OR IGNORE INTO employees (
    uid, name, mobile, government_id, university_id, email,
    hire_date, status, department_uid, shift_uid, created_at, updated_at
)
VALUES
    (
        'emp_university_president_20260508', 'احمد عادل عبد الحكيم',
        '+201000000074', '299000000074', 'U20260508PRES',
        'president.seed@university.edu.eg',
        '2026-05-08', 'active', NULL, 'shf_general_seed',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    ),
    (
        'emp_vice_president_20260508', 'عفاف العوفي',
        '+201000000075', '299000000075', 'U20260508VP',
        'vice.president.seed@university.edu.eg',
        '2026-05-08', 'active', NULL, 'shf_general_seed',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    ),
    (
        'emp_dean_engineering_20260508', 'وائل المغلاني',
        '+201000000076', '299000000076', 'U20260508DEANENG',
        'dean.engineering.seed@university.edu.eg',
        '2026-05-08', 'active', NULL, 'shf_general_seed',
        CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
    );

INSERT OR IGNORE INTO users (uid, phone, employee_uid, is_active, created_at, updated_at)
VALUES
    ('usr_university_president_20260508', '+201000000074', 'emp_university_president_20260508', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('usr_vice_president_20260508',       '+201000000075', 'emp_vice_president_20260508',       1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('usr_dean_engineering_20260508',     '+201000000076', 'emp_dean_engineering_20260508',     1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Assign base employee role
INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP
FROM users u, roles r
WHERE u.uid IN ('usr_university_president_20260508', 'usr_vice_president_20260508', 'usr_dean_engineering_20260508')
  AND r.uid = 'role_employee';

-- Assign leadership roles
INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP FROM users u, roles r
WHERE u.uid = 'usr_university_president_20260508' AND r.uid = 'role_university_president';

INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP FROM users u, roles r
WHERE u.uid = 'usr_vice_president_20260508' AND r.uid = 'role_vice_president';

INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT u.id, r.id, NULL, CURRENT_TIMESTAMP FROM users u, roles r
WHERE u.uid = 'usr_dean_engineering_20260508' AND r.uid = 'role_dean';
