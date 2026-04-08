-- Remove new leave types
DELETE FROM leave_types WHERE uid IN (
    'ltype_00000000000000000000000000000002',
    'ltype_00000000000000000000000000000003',
    'ltype_00000000000000000000000000000004',
    'ltype_00000000000000000000000000000005',
    'ltype_00000000000000000000000000000006',
    'ltype_00000000000000000000000000000007',
    'ltype_00000000000000000000000000000008'
);

-- Remove permissions from admin role
DELETE FROM role_permissions
WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN ('leave:request', 'approval:read', 'approval:write')
);

-- Remove new permissions
DELETE FROM permissions WHERE code IN ('leave:request', 'approval:read', 'approval:write');
