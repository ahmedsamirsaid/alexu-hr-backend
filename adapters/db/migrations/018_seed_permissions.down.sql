DELETE FROM permissions WHERE code IN (
    'employees:read',
    'employees:write',
    'employees:import',
    'employees:export',
    'leave:read',
    'leave:record',
    'leave:approve',
    'users:read',
    'users:write',
    'roles:read',
    'roles:write'
);
