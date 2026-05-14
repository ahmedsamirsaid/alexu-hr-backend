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
    type TEXT NOT NULL DEFAULT 'permanent'
        CHECK (type IN ('permanent', 'temporary')),
    sub_type TEXT NOT NULL DEFAULT 'normal'
        CHECK (
            (type = 'permanent' AND sub_type IN ('normal', 'special_needs')) OR
            (type = 'temporary' AND sub_type IN (
                'separation_termination_for_budget',
                'comprehensive_bonus',
                'contract_employees'
            ))
        ),
    department_uid TEXT REFERENCES departments(uid),
    shift_uid TEXT REFERENCES shifts(uid),
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT INTO employees_new (
    id, uid, name, mobile, government_id, university_id, email,
    hire_date, status, type, sub_type, department_uid, shift_uid, created_at, updated_at
)
SELECT
    id, uid, name, mobile, government_id, university_id, email,
    hire_date, status, type, sub_type, department_uid, shift_uid, created_at, updated_at
FROM employees;

DROP TABLE employees;
ALTER TABLE employees_new RENAME TO employees;

CREATE INDEX idx_employees_name ON employees(name);
CREATE INDEX idx_employees_status ON employees(status);
CREATE INDEX idx_employees_department ON employees(department_uid);
