CREATE TABLE roles_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    is_system INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO roles_new (id, uid, name, description, is_system, created_at, updated_at)
SELECT id, uid, name, description, is_system, created_at, updated_at
FROM roles;

DROP TABLE roles;
ALTER TABLE roles_new RENAME TO roles;

CREATE INDEX idx_roles_name ON roles(name);
