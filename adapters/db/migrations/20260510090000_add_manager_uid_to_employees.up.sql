ALTER TABLE employees ADD COLUMN manager_uid TEXT REFERENCES employees(uid) ON DELETE SET NULL;
CREATE INDEX idx_employees_manager_uid ON employees(manager_uid);
