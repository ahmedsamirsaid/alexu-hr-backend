INSERT INTO permissions (uid, code, description) VALUES
    ('perm_employees_read', 'employees:read', 'View employee list and details'),
    ('perm_employees_write', 'employees:write', 'Create/update employees'),
    ('perm_employees_import', 'employees:import', 'Bulk import employees'),
    ('perm_employees_export', 'employees:export', 'Export employees to Excel/PDF'),
    ('perm_leave_read', 'leave:read', 'View leave records'),
    ('perm_leave_record', 'leave:record', 'Record leave for employees'),
    ('perm_leave_approve', 'leave:approve', 'Approve/reject leave requests'),
    ('perm_users_read', 'users:read', 'View user accounts'),
    ('perm_users_write', 'users:write', 'Create/update user accounts'),
    ('perm_roles_read', 'roles:read', 'View roles and permissions'),
    ('perm_roles_write', 'roles:write', 'Create/update roles, assign permissions'),
    ('perm_documents_read', 'documents:read', 'Read Documents'),
    ('perm_documents_write', 'documents:write', 'Write Documents');
