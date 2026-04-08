CREATE TABLE holiday_definitions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    code TEXT UNIQUE NOT NULL,
    name_en TEXT NOT NULL,
    name_ar TEXT NOT NULL,
    default_month INTEGER,
    default_day INTEGER,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);
