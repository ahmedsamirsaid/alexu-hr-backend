CREATE TABLE holiday_instances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    definition_id INTEGER NOT NULL REFERENCES holiday_definitions(id) ON DELETE CASCADE,
    year INTEGER NOT NULL,
    actual_date TEXT NOT NULL,
    observed_date TEXT NOT NULL,
    is_confirmed INTEGER DEFAULT 0,
    notes TEXT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),
    UNIQUE(definition_id, year)
);

CREATE INDEX idx_holiday_instances_year ON holiday_instances(year);
CREATE INDEX idx_holiday_instances_observed ON holiday_instances(observed_date);
