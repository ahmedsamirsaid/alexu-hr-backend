CREATE TABLE audit_logs (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    uid         TEXT    UNIQUE NOT NULL,
    actor_uid   TEXT    NOT NULL,                     -- employee UID who performed the action
    action      TEXT    NOT NULL,                     -- e.g. approve, update, create
    entity_type TEXT    NOT NULL,                     -- e.g. leave_request, attendance_record
    entity_uid  TEXT    NOT NULL,                     -- UID of the affected entity
    meta        TEXT,                                 -- optional JSON blob (old/new values, comments…)
    occurred_at TEXT    NOT NULL DEFAULT (datetime('now')),
    created_at  TEXT    NOT NULL DEFAULT (datetime('now'))
);

-- Fast look-up by entity (most common query: "show history of this record")
CREATE INDEX idx_audit_logs_entity     ON audit_logs(entity_type, entity_uid);

-- Fast look-up by actor (e.g. "show everything Ahmed did today")
CREATE INDEX idx_audit_logs_actor      ON audit_logs(actor_uid);

-- Fast look-up by action type
CREATE INDEX idx_audit_logs_action     ON audit_logs(action);

-- Fast time-range filtering
CREATE INDEX idx_audit_logs_occurred   ON audit_logs(occurred_at);
