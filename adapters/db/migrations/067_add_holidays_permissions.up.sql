-- Add dedicated holiday permissions
INSERT INTO permissions (uid, code, description) VALUES
    ('perm_holidays_read', 'holidays:read', 'View holidays'),
    ('perm_holidays_write', 'holidays:write', 'Create/update holidays');

-- Preserve existing access semantics by mapping department permissions to holiday permissions.
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p_holidays_read.id
FROM role_permissions rp
JOIN permissions p_departments_read ON p_departments_read.id = rp.permission_id
JOIN permissions p_holidays_read ON p_holidays_read.code = 'holidays:read'
WHERE p_departments_read.code = 'departments:read';

INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p_holidays_write.id
FROM role_permissions rp
JOIN permissions p_departments_write ON p_departments_write.id = rp.permission_id
JOIN permissions p_holidays_write ON p_holidays_write.code = 'holidays:write'
WHERE p_departments_write.code = 'departments:write';
