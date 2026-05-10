-- Add dedicated shift permissions
INSERT INTO permissions (uid, code, description) VALUES
    ('perm_shift_read', 'shift:read', 'View shift definitions and assignments'),
    ('perm_shift_write', 'shift:write', 'Create/update shift definitions and assignments');

-- Preserve existing access semantics by mapping attendance permissions to shift permissions.
INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p_shift_read.id
FROM role_permissions rp
JOIN permissions p_attendance_read ON p_attendance_read.id = rp.permission_id
JOIN permissions p_shift_read ON p_shift_read.code = 'shift:read'
WHERE p_attendance_read.code = 'attendance:read';

INSERT INTO role_permissions (role_id, permission_id)
SELECT rp.role_id, p_shift_write.id
FROM role_permissions rp
JOIN permissions p_attendance_write ON p_attendance_write.id = rp.permission_id
JOIN permissions p_shift_write ON p_shift_write.code = 'shift:write'
WHERE p_attendance_write.code = 'attendance:write';
