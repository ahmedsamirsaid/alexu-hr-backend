ALTER TABLE leave_records ADD COLUMN leave_request_uid TEXT REFERENCES leave_requests(uid) ON DELETE SET NULL;

CREATE INDEX idx_leave_records_request ON leave_records(leave_request_uid);
