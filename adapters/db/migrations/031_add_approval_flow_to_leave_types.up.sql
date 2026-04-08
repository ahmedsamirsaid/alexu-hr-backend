ALTER TABLE leave_types ADD COLUMN approval_flow_uid TEXT REFERENCES approval_flows(uid);

CREATE INDEX idx_leave_types_approval_flow ON leave_types(approval_flow_uid);
