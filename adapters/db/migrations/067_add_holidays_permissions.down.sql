-- Remove holiday permissions from role mappings.
DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id
    FROM permissions
    WHERE code IN ('holidays:read', 'holidays:write')
);

-- Remove holiday permissions.
DELETE FROM permissions
WHERE code IN ('holidays:read', 'holidays:write');
