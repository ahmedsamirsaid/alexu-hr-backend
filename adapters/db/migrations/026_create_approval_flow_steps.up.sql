CREATE TABLE approval_flow_steps (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    approval_flow_uid TEXT NOT NULL REFERENCES approval_flows(uid) ON DELETE RESTRICT,
    step_order INTEGER NOT NULL CHECK (step_order > 0),
    role_uid TEXT NOT NULL REFERENCES roles(uid) ON DELETE RESTRICT,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now')),

    UNIQUE(approval_flow_uid, step_order)
);

CREATE INDEX idx_approval_flow_steps_flow ON approval_flow_steps(approval_flow_uid);
CREATE INDEX idx_approval_flow_steps_role ON approval_flow_steps(role_uid);
