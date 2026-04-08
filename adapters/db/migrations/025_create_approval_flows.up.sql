CREATE TABLE approval_flows (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    code TEXT UNIQUE NOT NULL,
    name_en TEXT NOT NULL,
    name_ar TEXT,
    description TEXT,
    is_active INTEGER DEFAULT 1,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE INDEX idx_approval_flows_code ON approval_flows(code);
CREATE INDEX idx_approval_flows_is_active ON approval_flows(is_active);
