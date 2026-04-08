CREATE TABLE employees (
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

CREATE INDEX idx_employees_name ON employees(name);
CREATE INDEX idx_employees_status ON employees(status);
