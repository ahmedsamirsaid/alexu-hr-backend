DROP TABLE IF EXISTS holiday_instances;
DROP TABLE IF EXISTS holiday_definitions;

CREATE TABLE holiday_definitions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    code TEXT UNIQUE NOT NULL,
    name_en TEXT NOT NULL,
    name_ar TEXT NOT NULL,
    date TEXT NOT NULL,
    is_manual INTEGER NOT NULL DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX idx_holiday_definitions_date ON holiday_definitions(date);
