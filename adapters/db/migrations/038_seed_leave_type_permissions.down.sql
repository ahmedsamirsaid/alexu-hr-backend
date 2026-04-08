-- Remove leave type permissions from admin role
DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code IN ('leave-types:read', 'leave-types:write'));

-- Remove leave type permissions
DELETE FROM permissions WHERE code IN ('leave-types:read', 'leave-types:write');
