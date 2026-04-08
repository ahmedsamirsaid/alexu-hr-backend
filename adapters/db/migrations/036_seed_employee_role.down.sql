-- Remove permissions from Employee role
DELETE FROM role_permissions WHERE role_id = (SELECT id FROM roles WHERE name = 'Employee');

-- Delete Employee role
DELETE FROM roles WHERE name = 'Employee';
