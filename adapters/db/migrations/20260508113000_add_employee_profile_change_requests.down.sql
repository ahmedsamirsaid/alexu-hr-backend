DROP TRIGGER IF EXISTS trg_employee_profile_change_requests_single_pending;
DROP INDEX IF EXISTS idx_employee_profile_change_requests_created_at;
DROP INDEX IF EXISTS idx_employee_profile_change_requests_submitted_by;
DROP INDEX IF EXISTS idx_employee_profile_change_requests_employee_uid;
DROP TABLE IF EXISTS employee_profile_change_requests;

DELETE FROM approval_flow_steps WHERE uid = 'afs_employee_profile_change_step1';
DELETE FROM approval_flows WHERE uid = 'apf_employee_profile_change';

DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_information_center')
  AND permission_id IN (
    SELECT id
    FROM permissions
    WHERE code IN (
        'employees:read',
        'employees:write',
        'employee-profile-changes:read',
        'employee-profile-changes:approve'
    )
);

DELETE FROM role_permissions
WHERE role_id IN (
    SELECT id
    FROM roles
    WHERE uid IN ('role_university_human_resources', 'role_hr_manager')
)
  AND permission_id = (
    SELECT id
    FROM permissions
    WHERE code = 'employee-profile-changes:write'
);

DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE uid = 'role_department_manager')
  AND permission_id IN (
    SELECT id
    FROM permissions
    WHERE code IN ('employee-profile-changes:read', 'employee-profile-changes:write')
);

DELETE FROM roles WHERE uid = 'role_information_center';

DELETE FROM permissions
WHERE code IN (
    'employee-profile-changes:read',
    'employee-profile-changes:write',
    'employee-profile-changes:approve'
);
