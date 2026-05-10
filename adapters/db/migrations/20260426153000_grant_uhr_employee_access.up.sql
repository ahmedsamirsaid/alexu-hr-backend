-- Grant University Human Resources department visibility, give Employee holiday visibility, and assign Employee role to UHR users.
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code = 'departments:read'
WHERE r.uid = 'role_university_human_resources';

INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT r.id, p.id, CURRENT_TIMESTAMP
FROM roles r
JOIN permissions p ON p.code = 'holidays:read'
WHERE r.uid = 'role_employee';

INSERT INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT uhr_users.user_id, employee_role.id, NULL, CURRENT_TIMESTAMP
FROM (
    SELECT DISTINCT ur.user_id
    FROM user_roles ur
    JOIN roles r ON r.id = ur.role_id
    WHERE r.uid = 'role_university_human_resources'
) uhr_users
JOIN roles employee_role ON employee_role.uid = 'role_employee';
