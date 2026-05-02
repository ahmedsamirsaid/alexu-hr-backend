-- Add audit:read permission
INSERT INTO permissions (uid, code, description) VALUES
    ('perm_audit_read', 'audit:read', 'View audit logs and system activity history');
