-- Remove shift permissions from role mappings.
DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id
    FROM permissions
    WHERE code IN ('shift:read', 'shift:write')
);

-- Remove shift permissions.
DELETE FROM permissions
WHERE code IN ('shift:read', 'shift:write');
