CREATE TABLE rate_limit_records (
    id SERIAL PRIMARY KEY,
    scope TEXT NOT NULL,
    subject_key TEXT NOT NULL,
    window_started_at TIMESTAMPTZ NOT NULL,
    attempt_count INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMPTZ NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_rate_limit_records_scope_subject
    ON rate_limit_records(scope, subject_key);

CREATE INDEX idx_rate_limit_records_locked_until
    ON rate_limit_records(locked_until);
