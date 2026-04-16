DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN ('attendance:read', 'attendance:write')
);

DELETE FROM permissions
WHERE code IN ('attendance:read', 'attendance:write');
