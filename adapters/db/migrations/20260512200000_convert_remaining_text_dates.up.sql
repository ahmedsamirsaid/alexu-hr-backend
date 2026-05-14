-- Convert remaining TEXT date/time columns to proper PostgreSQL DATE/TIMESTAMP types

-- employees.hire_date TEXT → DATE
ALTER TABLE employees ALTER COLUMN hire_date TYPE DATE USING hire_date::DATE;

-- leave_records.start_date, end_date TEXT → DATE
ALTER TABLE leave_records ALTER COLUMN start_date TYPE DATE USING start_date::DATE;
ALTER TABLE leave_records ALTER COLUMN end_date TYPE DATE USING end_date::DATE;

-- leave_requests.start_date, end_date TEXT → DATE
-- First drop the CHECK constraint that compares these columns
ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS valid_date_range;
ALTER TABLE leave_requests ALTER COLUMN start_date TYPE DATE USING start_date::DATE;
ALTER TABLE leave_requests ALTER COLUMN end_date TYPE DATE USING end_date::DATE;
-- Recreate the constraint
ALTER TABLE leave_requests ADD CONSTRAINT valid_date_range CHECK (end_date >= start_date);

-- attendance_exceptions.attendance_date TEXT → DATE
ALTER TABLE attendance_exceptions ALTER COLUMN attendance_date TYPE DATE USING attendance_date::DATE;

-- attendance_exceptions.check_in, check_out TEXT → TIMESTAMP (nullable)
ALTER TABLE attendance_exceptions ALTER COLUMN check_in TYPE TIMESTAMP USING
    CASE
        WHEN check_in IS NULL OR check_in = '' THEN NULL
        ELSE check_in::TIMESTAMP
    END;
ALTER TABLE attendance_exceptions ALTER COLUMN check_out TYPE TIMESTAMP USING
    CASE
        WHEN check_out IS NULL OR check_out = '' THEN NULL
        ELSE check_out::TIMESTAMP
    END;

-- attendance_reminders.attendance_date TEXT → DATE
ALTER TABLE attendance_reminders ALTER COLUMN attendance_date TYPE DATE USING attendance_date::DATE;

-- holiday_definitions.date TEXT → DATE
ALTER TABLE holiday_definitions ALTER COLUMN date TYPE DATE USING date::DATE;

-- holiday_instances.actual_date, observed_date TEXT → DATE
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'holiday_instances'
    ) THEN
        ALTER TABLE holiday_instances ALTER COLUMN actual_date TYPE DATE USING actual_date::DATE;
        ALTER TABLE holiday_instances ALTER COLUMN observed_date TYPE DATE USING observed_date::DATE;
    END IF;
END $$;

-- permission_requests.permission_date TEXT → DATE
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'permission_requests' AND column_name = 'permission_date'
    ) THEN
        ALTER TABLE permission_requests ALTER COLUMN permission_date TYPE DATE USING permission_date::DATE;
    END IF;
END $$;
