-- Convert TEXT timestamp columns to proper TIMESTAMP type for PostgreSQL

-- audit_logs table
ALTER TABLE audit_logs ALTER COLUMN occurred_at DROP DEFAULT;
ALTER TABLE audit_logs ALTER COLUMN occurred_at TYPE TIMESTAMP USING occurred_at::TIMESTAMP;
ALTER TABLE audit_logs ALTER COLUMN occurred_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE audit_logs ALTER COLUMN created_at DROP DEFAULT;
ALTER TABLE audit_logs ALTER COLUMN created_at TYPE TIMESTAMP USING created_at::TIMESTAMP;
ALTER TABLE audit_logs ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP;

-- approval_actions table
ALTER TABLE approval_actions ALTER COLUMN acted_at DROP DEFAULT;
ALTER TABLE approval_actions ALTER COLUMN acted_at TYPE TIMESTAMP USING acted_at::TIMESTAMP;
ALTER TABLE approval_actions ALTER COLUMN acted_at SET DEFAULT CURRENT_TIMESTAMP;

-- leave_requests table
ALTER TABLE leave_requests ALTER COLUMN submitted_at DROP DEFAULT;
ALTER TABLE leave_requests ALTER COLUMN submitted_at TYPE TIMESTAMP USING submitted_at::TIMESTAMP;
ALTER TABLE leave_requests ALTER COLUMN submitted_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE leave_requests ALTER COLUMN decided_at TYPE TIMESTAMP USING 
    CASE 
        WHEN decided_at IS NULL OR decided_at = '' THEN NULL 
        ELSE decided_at::TIMESTAMP 
    END;

-- permission_requests table
ALTER TABLE permission_requests ALTER COLUMN submitted_at DROP DEFAULT;
ALTER TABLE permission_requests ALTER COLUMN submitted_at TYPE TIMESTAMP USING submitted_at::TIMESTAMP;
ALTER TABLE permission_requests ALTER COLUMN submitted_at SET DEFAULT CURRENT_TIMESTAMP;

ALTER TABLE permission_requests ALTER COLUMN decided_at TYPE TIMESTAMP USING 
    CASE 
        WHEN decided_at IS NULL OR decided_at = '' THEN NULL 
        ELSE decided_at::TIMESTAMP 
    END;

-- leave_records table
ALTER TABLE leave_records ALTER COLUMN recorded_at TYPE TIMESTAMP USING recorded_at::TIMESTAMP;

-- attendance_records table
ALTER TABLE attendance_records ALTER COLUMN punched_at TYPE TIMESTAMP USING punched_at::TIMESTAMP;

-- attendance_reminders table
ALTER TABLE attendance_reminders ALTER COLUMN sent_at TYPE TIMESTAMP USING sent_at::TIMESTAMP;
