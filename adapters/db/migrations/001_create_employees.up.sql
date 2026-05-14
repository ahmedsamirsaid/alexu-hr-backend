CREATE TABLE employees (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    mobile TEXT UNIQUE NOT NULL,
    government_id TEXT UNIQUE NOT NULL,
    university_id TEXT UNIQUE NOT NULL,
    email TEXT,
    hire_date DATE NOT NULL,
    status TEXT DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_employees_name ON employees(name);
CREATE INDEX idx_employees_status ON employees(status);
