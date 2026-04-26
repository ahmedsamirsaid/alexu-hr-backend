-- Grant Employee role read access to employee data.
-- Uses the existing canonical permission code: employees:read.
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code = 'employees:read'
WHERE r.uid = 'role_employee';
