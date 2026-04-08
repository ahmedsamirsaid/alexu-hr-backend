-- Seed new permissions for leave requests and approvals
INSERT OR IGNORE INTO permissions (uid, code, description) VALUES
    ('perm_leave_request', 'leave:request', 'Submit leave requests'),
    ('perm_approval_read', 'approval:read', 'View approval flows and configuration'),
    ('perm_approval_write', 'approval:write', 'Create/update approval flows and steps');

-- Add new permissions to admin role (role_id=1 is the Admin role from 019_seed_admin_role)
INSERT OR IGNORE INTO role_permissions (role_id, permission_id)
SELECT
    (SELECT id FROM roles WHERE name = 'Admin'),
    id
FROM permissions
WHERE code IN ('leave:request', 'approval:read', 'approval:write');

-- Update existing leave types to add approval flow
UPDATE leave_types SET approval_flow_uid = 'apf_leave_default' WHERE code = 'ANNUAL';
UPDATE leave_types SET approval_flow_uid = 'apf_leave_default' WHERE code = 'SICK';

-- Seed additional leave types with approval flow (use INSERT OR IGNORE in case they exist)
INSERT OR IGNORE INTO leave_types (uid, code, name_en, name_ar, default_balance, max_consecutive, recording_deadline_days, advance_notice_days, is_active, approval_flow_uid)
VALUES
    ('ltype_00000000000000000000000000000004', 'STUDY', 'Study Leave', 'دراسية', 90, NULL, NULL, 30, 1, 'apf_leave_default'),
    ('ltype_00000000000000000000000000000005', 'INTERNAL_SECONDMENT', 'Internal Secondment', 'إعارة داخلية', 365, NULL, NULL, 30, 1, 'apf_leave_default'),
    ('ltype_00000000000000000000000000000006', 'EXTERNAL_SECONDMENT', 'External Secondment', 'إعارة خارجية', 365, NULL, NULL, 30, 1, 'apf_leave_default'),
    ('ltype_00000000000000000000000000000007', 'CHILD_CARE', 'Child Care Leave', 'رعاية طفل', 730, NULL, NULL, 14, 1, 'apf_leave_default'),
    ('ltype_00000000000000000000000000000008', 'SPOUSE_ACCOMPANIMENT', 'Spouse Accompaniment Leave', 'مرافقة زوج', 365, NULL, NULL, 14, 1, 'apf_leave_default');
