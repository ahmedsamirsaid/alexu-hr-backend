-- Fix department_manager role name to be human-readable
UPDATE roles SET name = 'Department Manager' WHERE uid = 'role_department_manager';
