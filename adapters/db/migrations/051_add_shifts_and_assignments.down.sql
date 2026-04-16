DROP TABLE IF EXISTS shifts;

CREATE TABLE departments_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    code TEXT UNIQUE NOT NULL,
    name_en TEXT NOT NULL,
    name_ar TEXT,
    is_active INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT INTO departments_new (id, uid, code, name_en, name_ar, is_active, created_at, updated_at)
SELECT id, uid, code, name_en, name_ar, is_active, created_at, updated_at
FROM departments;

DROP TABLE departments;
ALTER TABLE departments_new RENAME TO departments;

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
    department_uid TEXT REFERENCES departments(uid),
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT INTO employees_new (id, uid, name, mobile, government_id, university_id, email, hire_date, status, department_uid, created_at, updated_at)
SELECT id, uid, name, mobile, government_id, university_id, email, hire_date, status, department_uid, created_at, updated_at
FROM employees;

DROP TABLE employees;
ALTER TABLE employees_new RENAME TO employees;
