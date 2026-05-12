-- Wire the three seeded leaders into a reporting chain.
-- Runs after the employees exist, so UPDATEs are guaranteed to match rows.

UPDATE employees SET manager_uid = 'emp_university_president_20260508'
    WHERE uid = 'emp_vice_president_20260508';

UPDATE employees SET manager_uid = 'emp_vice_president_20260508'
    WHERE uid = 'emp_dean_engineering_20260508';

-- Bootstrap existing department managers → dean
UPDATE employees
SET manager_uid = 'emp_dean_engineering_20260508'
WHERE uid IN (
    SELECT e.uid FROM employees e
    JOIN users u ON u.employee_uid = e.uid
    JOIN user_roles ur ON ur.user_id = u.id
    JOIN roles r ON r.id = ur.role_id AND r.uid = 'role_department_manager'
);

-- Bootstrap regular employees → their department manager
UPDATE employees
SET manager_uid = (
    SELECT mgr_emp.uid FROM users mgr_u
    JOIN user_roles ur ON ur.user_id = mgr_u.id
    JOIN roles r ON r.id = ur.role_id AND r.uid = 'role_department_manager'
    JOIN employees mgr_emp ON mgr_emp.uid = mgr_u.employee_uid
    WHERE ur.department_uid = employees.department_uid
    LIMIT 1
)
WHERE manager_uid IS NULL
    AND department_uid IS NOT NULL
    AND uid NOT IN (
        SELECT e.uid FROM employees e
        JOIN users u ON u.employee_uid = e.uid
        JOIN user_roles ur ON ur.user_id = u.id
        JOIN roles r ON r.id = ur.role_id AND r.uid = 'role_department_manager'
    );
