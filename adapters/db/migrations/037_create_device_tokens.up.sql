CREATE TABLE device_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    user_uid TEXT NOT NULL REFERENCES users(uid) ON DELETE CASCADE,
    token TEXT NOT NULL,
    platform TEXT NOT NULL CHECK (platform IN ('android', 'ios')),
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

CREATE UNIQUE INDEX idx_device_tokens_token ON device_tokens(token);
CREATE INDEX idx_device_tokens_user ON device_tokens(user_uid);
