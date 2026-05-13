-- Revert BOOLEAN columns back to INTEGER

-- otp_codes.used
ALTER TABLE otp_codes ALTER COLUMN used TYPE INTEGER USING (CASE WHEN used THEN 1 ELSE 0 END);
ALTER TABLE otp_codes ALTER COLUMN used SET DEFAULT 0;

-- refresh_tokens.revoked
ALTER TABLE refresh_tokens ALTER COLUMN revoked TYPE INTEGER USING (CASE WHEN revoked THEN 1 ELSE 0 END);
ALTER TABLE refresh_tokens ALTER COLUMN revoked SET DEFAULT 0;

-- users.is_active
ALTER TABLE users ALTER COLUMN is_active TYPE INTEGER USING (CASE WHEN is_active THEN 1 ELSE 0 END);
ALTER TABLE users ALTER COLUMN is_active SET DEFAULT 1;

-- roles.is_system
ALTER TABLE roles ALTER COLUMN is_system TYPE INTEGER USING (CASE WHEN is_system THEN 1 ELSE 0 END);
ALTER TABLE roles ALTER COLUMN is_system SET DEFAULT 0;

-- leave_types.is_active
ALTER TABLE leave_types ALTER COLUMN is_active TYPE INTEGER USING (CASE WHEN is_active THEN 1 ELSE 0 END);
ALTER TABLE leave_types ALTER COLUMN is_active SET DEFAULT 1;

-- leave_types.is_paid
ALTER TABLE leave_types ALTER COLUMN is_paid TYPE INTEGER USING (CASE WHEN is_paid THEN 1 ELSE 0 END);
ALTER TABLE leave_types ALTER COLUMN is_paid SET DEFAULT 1;

-- departments.is_active
ALTER TABLE departments ALTER COLUMN is_active TYPE INTEGER USING (CASE WHEN is_active THEN 1 ELSE 0 END);
ALTER TABLE departments ALTER COLUMN is_active SET DEFAULT 1;

-- approval_flows.is_active
ALTER TABLE approval_flows ALTER COLUMN is_active TYPE INTEGER USING (CASE WHEN is_active THEN 1 ELSE 0 END);
ALTER TABLE approval_flows ALTER COLUMN is_active SET DEFAULT 1;

-- holiday_definitions.is_manual
ALTER TABLE holiday_definitions ALTER COLUMN is_manual TYPE INTEGER USING (CASE WHEN is_manual THEN 1 ELSE 0 END);
ALTER TABLE holiday_definitions ALTER COLUMN is_manual SET DEFAULT 1;

-- attendance_devices.is_active (if exists)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'attendance_devices' AND column_name = 'is_active'
    ) THEN
        ALTER TABLE attendance_devices ALTER COLUMN is_active TYPE INTEGER USING (CASE WHEN is_active THEN 1 ELSE 0 END);
        ALTER TABLE attendance_devices ALTER COLUMN is_active SET DEFAULT 1;
    END IF;
END $$;

-- shifts.is_active (if exists)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shifts' AND column_name = 'is_active'
    ) THEN
        ALTER TABLE shifts ALTER COLUMN is_active TYPE INTEGER USING (CASE WHEN is_active THEN 1 ELSE 0 END);
        ALTER TABLE shifts ALTER COLUMN is_active SET DEFAULT 1;
    END IF;
END $$;
