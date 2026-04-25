CREATE TABLE leave_request_documents (
    leave_request_uid TEXT NOT NULL REFERENCES leave_requests(uid) ON DELETE CASCADE,
    file_name TEXT NOT NULL,
    object_key TEXT NOT NULL UNIQUE
);

CREATE INDEX idx_leave_request_documents_leave_request_uid ON leave_request_documents(leave_request_uid);

ALTER TABLE leave_requests ADD COLUMN sub_leave_type_uid TEXT REFERENCES sub_leave_types(uid) ON DELETE SET NULL;

CREATE INDEX idx_leave_requests_sub_leave_type ON leave_requests(sub_leave_type_uid);
