ALTER TABLE user_roles ADD COLUMN department_uid TEXT REFERENCES departments(uid);

CREATE INDEX idx_user_roles_department ON user_roles(department_uid);
