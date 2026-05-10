CREATE TABLE approval_actions (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    approval_request_uid TEXT NOT NULL REFERENCES approval_requests(uid) ON DELETE RESTRICT,
    action TEXT NOT NULL CHECK (action IN ('submit', 'approve', 'reject', 'cancel')),
    step_order INTEGER,
    actor_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE RESTRICT,
    comments TEXT,
    acted_at TEXT NOT NULL DEFAULT (CURRENT_TIMESTAMP),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_approval_actions_request ON approval_actions(approval_request_uid);
CREATE INDEX idx_approval_actions_actor ON approval_actions(actor_uid);
CREATE INDEX idx_approval_actions_action ON approval_actions(action);
