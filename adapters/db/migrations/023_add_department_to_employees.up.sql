ALTER TABLE employees ADD COLUMN department_uid TEXT REFERENCES departments(uid);

CREATE INDEX idx_employees_department ON employees(department_uid);
