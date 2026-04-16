DELETE FROM role_permissions
WHERE permission_id IN (
	SELECT id
	FROM permissions
	WHERE code IN ('attendance-devices:read', 'attendance-devices:write')
);

DELETE FROM permissions
WHERE code IN ('attendance-devices:read', 'attendance-devices:write');
