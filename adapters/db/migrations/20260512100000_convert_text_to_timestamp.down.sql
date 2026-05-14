-- Revert timestamp columns back to TEXT (for rollback)

-- audit_logs table
ALTER TABLE audit_logs ALTER COLUMN occurred_at DROP DEFAULT;
ALTER TABLE audit_logs ALTER COLUMN occurred_at TYPE TEXT USING occurred_at::TEXT;
ALTER TABLE audit_logs ALTER COLUMN occurred_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE audit_logs ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE audit_logs ALTER COLUMN created_at TYPE TEXT USING created_at::TEXT;
ALTER TABLE audit_logs ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;

-- approval_actions table
ALTER TABLE approval_actions ALTER COLUMN acted_at DROP DEFAULT;
ALTER TABLE approval_actions ALTER COLUMN acted_at TYPE TEXT USING acted_at::TEXT;
ALTER TABLE approval_actions ALTER COLUMN acted_at SET DEFAULT CURRENT_TIMESTAMP;

-- leave_requests table
ALTER TABLE leave_requests ALTER COLUMN submitted_at DROP DEFAULT;
ALTER TABLE leave_requests ALTER COLUMN submitted_at TYPE TEXT USING submitted_at::TEXT;
ALTER TABLE leave_requests ALTER COLUMN submitted_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE leave_requests ALTER COLUMN decided_at TYPE TEXT USING decided_at::TEXT;

-- permission_requests table
ALTER TABLE permission_requests ALTER COLUMN submitted_at DROP DEFAULT;
ALTER TABLE permission_requests ALTER COLUMN submitted_at TYPE TEXT USING submitted_at::TEXT;
ALTER TABLE permission_requests ALTER COLUMN submitted_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE permission_requests ALTER COLUMN decided_at TYPE TEXT USING decided_at::TEXT;

-- leave_records table
ALTER TABLE leave_records ALTER COLUMN recorded_at TYPE TEXT USING recorded_at::TEXT;

-- attendance_records table
ALTER TABLE attendance_records ALTER COLUMN punched_at TYPE TEXT USING punched_at::TEXT;

-- attendance_reminders table
ALTER TABLE attendance_reminders ALTER COLUMN sent_at TYPE TEXT USING sent_at::TEXT;
