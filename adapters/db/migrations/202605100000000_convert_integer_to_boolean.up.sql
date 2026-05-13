-- Convert INTEGER boolean columns to proper BOOLEAN type for PostgreSQL

-- otp_codes.used
ALTER TABLE otp_codes ALTER COLUMN used DROP DEFAULT;
ALTER TABLE otp_codes ALTER COLUMN used TYPE BOOLEAN USING (used::INTEGER != 0);
ALTER TABLE otp_codes ALTER COLUMN used SET DEFAULT FALSE;

-- refresh_tokens.revoked
ALTER TABLE refresh_tokens ALTER COLUMN revoked DROP DEFAULT;
ALTER TABLE refresh_tokens ALTER COLUMN revoked TYPE BOOLEAN USING (revoked::INTEGER != 0);
ALTER TABLE refresh_tokens ALTER COLUMN revoked SET DEFAULT FALSE;

-- users.is_active
ALTER TABLE users ALTER COLUMN is_active DROP DEFAULT;
ALTER TABLE users ALTER COLUMN is_active TYPE BOOLEAN USING (is_active::INTEGER != 0);
ALTER TABLE users ALTER COLUMN is_active SET DEFAULT TRUE;

-- roles.is_system
ALTER TABLE roles ALTER COLUMN is_system DROP DEFAULT;
ALTER TABLE roles ALTER COLUMN is_system TYPE BOOLEAN USING (is_system::INTEGER != 0);
ALTER TABLE roles ALTER COLUMN is_system SET DEFAULT FALSE;

-- leave_types.is_active
ALTER TABLE leave_types ALTER COLUMN is_active DROP DEFAULT;
ALTER TABLE leave_types ALTER COLUMN is_active TYPE BOOLEAN USING (is_active::INTEGER != 0);
ALTER TABLE leave_types ALTER COLUMN is_active SET DEFAULT TRUE;

-- leave_types.is_paid
ALTER TABLE leave_types ALTER COLUMN is_paid DROP DEFAULT;
ALTER TABLE leave_types ALTER COLUMN is_paid TYPE BOOLEAN USING (is_paid::INTEGER != 0);
ALTER TABLE leave_types ALTER COLUMN is_paid SET DEFAULT TRUE;

-- departments.is_active
ALTER TABLE departments ALTER COLUMN is_active DROP DEFAULT;
ALTER TABLE departments ALTER COLUMN is_active TYPE BOOLEAN USING (is_active::INTEGER != 0);
ALTER TABLE departments ALTER COLUMN is_active SET DEFAULT TRUE;

-- approval_flows.is_active
ALTER TABLE approval_flows ALTER COLUMN is_active DROP DEFAULT;
ALTER TABLE approval_flows ALTER COLUMN is_active TYPE BOOLEAN USING (is_active::INTEGER != 0);
ALTER TABLE approval_flows ALTER COLUMN is_active SET DEFAULT TRUE;

-- holiday_definitions.is_manual
ALTER TABLE holiday_definitions ALTER COLUMN is_manual DROP DEFAULT;
ALTER TABLE holiday_definitions ALTER COLUMN is_manual TYPE BOOLEAN USING (is_manual::INTEGER != 0);
ALTER TABLE holiday_definitions ALTER COLUMN is_manual SET DEFAULT TRUE;

-- attendance_devices.is_active (if exists)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'attendance_devices' AND column_name = 'is_active'
    ) THEN
        ALTER TABLE attendance_devices ALTER COLUMN is_active DROP DEFAULT;
        ALTER TABLE attendance_devices ALTER COLUMN is_active TYPE BOOLEAN USING (is_active::INTEGER != 0);
        ALTER TABLE attendance_devices ALTER COLUMN is_active SET DEFAULT TRUE;
    END IF;
END $$;

-- shifts.is_active (if exists)
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'shifts' AND column_name = 'is_active'
    ) THEN
        ALTER TABLE shifts ALTER COLUMN is_active DROP DEFAULT;
        ALTER TABLE shifts ALTER COLUMN is_active TYPE BOOLEAN USING (is_active::INTEGER != 0);
        ALTER TABLE shifts ALTER COLUMN is_active SET DEFAULT TRUE;
    END IF;
END $$;
