-- Drop permission requests infrastructure in reverse order

-- 1. Drop table and indexes
DROP INDEX IF EXISTS idx_permission_requests_type_date;
DROP INDEX IF EXISTS idx_permission_requests_approval;
DROP INDEX IF EXISTS idx_permission_requests_employee_date;
DROP TABLE IF EXISTS permission_requests;

-- 2. Drop approval flow steps and flow
DELETE FROM approval_flow_steps WHERE approval_flow_uid = 'apf_permission_default';
DELETE FROM approval_flows WHERE uid = 'apf_permission_default';

-- 3. Drop permissions and role assignments
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE code IN ('permission:request', 'permission:approve', 'permission:read')
);
DELETE FROM permissions WHERE code IN ('permission:request', 'permission:approve', 'permission:read');
