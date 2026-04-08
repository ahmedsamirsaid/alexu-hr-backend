-- Remove department permissions from admin role
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN ('departments:read', 'departments:write')
);

-- Remove department permissions
DELETE FROM permissions WHERE code IN ('departments:read', 'departments:write');
