ALTER TABLE roles ADD COLUMN scope_type TEXT NOT NULL DEFAULT 'global';

UPDATE roles
SET scope_type = 'global'
WHERE scope_type IS NULL OR TRIM(scope_type) = '';

UPDATE roles
SET scope_type = 'self'
WHERE uid = 'role_employee';

UPDATE roles
SET scope_type = 'department'
WHERE uid = 'role_department_manager';
