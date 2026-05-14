CREATE TABLE rate_limit_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scope TEXT NOT NULL,
    subject_key TEXT NOT NULL,
    window_started_at TIMESTAMP NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_rate_limit_records_scope_subject
    ON rate_limit_records(scope, subject_key);

CREATE INDEX idx_rate_limit_records_locked_until
    ON rate_limit_records(locked_until);
