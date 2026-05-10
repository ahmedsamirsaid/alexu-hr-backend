DROP TABLE IF EXISTS holiday_instances;
DROP TABLE IF EXISTS holiday_definitions;

CREATE TABLE holiday_definitions (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    code TEXT UNIQUE NOT NULL,
    name_en TEXT NOT NULL,
    name_ar TEXT NOT NULL,
    date TEXT NOT NULL,
    is_manual INTEGER NOT NULL DEFAULT 1,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_holiday_definitions_date ON holiday_definitions(date);

CREATE TABLE holiday_definition_departments (
    holiday_definition_id INTEGER NOT NULL,
    department_uid TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (holiday_definition_id, department_uid),
    FOREIGN KEY (holiday_definition_id) REFERENCES holiday_definitions(id) ON DELETE CASCADE,
    FOREIGN KEY (department_uid) REFERENCES departments(uid) ON DELETE CASCADE
);

CREATE INDEX idx_holiday_definition_departments_department_uid ON holiday_definition_departments(department_uid);
