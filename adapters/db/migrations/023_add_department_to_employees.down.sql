DROP INDEX IF EXISTS idx_employees_department;

-- SQLite doesn't support DROP COLUMN directly, so we need to recreate the table
CREATE TABLE employees_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    mobile TEXT UNIQUE NOT NULL,
    government_id TEXT UNIQUE NOT NULL,
    university_id TEXT UNIQUE NOT NULL,
    email TEXT,
    hire_date TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT INTO employees_new (id, uid, name, mobile, government_id, university_id, email, hire_date, status, created_at, updated_at)
SELECT id, uid, name, mobile, government_id, university_id, email, hire_date, status, created_at, updated_at
FROM employees;

DROP TABLE employees;
ALTER TABLE employees_new RENAME TO employees;

CREATE INDEX idx_employees_name ON employees(name);
CREATE INDEX idx_employees_status ON employees(status);
