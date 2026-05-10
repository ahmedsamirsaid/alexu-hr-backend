CREATE TABLE approval_requests (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    approval_flow_uid TEXT NOT NULL REFERENCES approval_flows(uid) ON DELETE RESTRICT,
    requester_uid TEXT NOT NULL REFERENCES employees(uid) ON DELETE RESTRICT,
    current_step INTEGER NOT NULL DEFAULT 1,
    max_step INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_approval_requests_flow ON approval_requests(approval_flow_uid);
CREATE INDEX idx_approval_requests_requester ON approval_requests(requester_uid);
CREATE INDEX idx_approval_requests_status ON approval_requests(status);
CREATE INDEX idx_approval_requests_pending ON approval_requests(status, current_step) WHERE status = 'pending';
