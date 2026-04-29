-- Remove audit:read permission
DELETE FROM permissions WHERE code = 'audit:read';
