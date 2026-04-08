DROP INDEX IF EXISTS idx_user_roles_department;

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
CREATE TABLE user_roles_new (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

INSERT INTO user_roles_new (user_id, role_id, created_at)
SELECT user_id, role_id, created_at
FROM user_roles;

DROP TABLE user_roles;
ALTER TABLE user_roles_new RENAME TO user_roles;

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
