-- Revert DATE/TIMESTAMP columns back to TEXT

-- employees.hire_date DATE → TEXT
ALTER TABLE employees ALTER COLUMN hire_date TYPE TEXT;

-- leave_records.start_date, end_date DATE → TEXT
ALTER TABLE leave_records ALTER COLUMN start_date TYPE TEXT;
ALTER TABLE leave_records ALTER COLUMN end_date TYPE TEXT;

-- leave_requests.start_date, end_date DATE → TEXT
ALTER TABLE leave_requests DROP CONSTRAINT IF EXISTS valid_date_range;
ALTER TABLE leave_requests ALTER COLUMN start_date TYPE TEXT;
ALTER TABLE leave_requests ALTER COLUMN end_date TYPE TEXT;
ALTER TABLE leave_requests ADD CONSTRAINT valid_date_range CHECK (end_date >= start_date);

-- attendance_exceptions.attendance_date DATE → TEXT
ALTER TABLE attendance_exceptions ALTER COLUMN attendance_date TYPE TEXT;

-- attendance_exceptions.check_in, check_out TIMESTAMP → TEXT
ALTER TABLE attendance_exceptions ALTER COLUMN check_in TYPE TEXT;
ALTER TABLE attendance_exceptions ALTER COLUMN check_out TYPE TEXT;

-- attendance_reminders.attendance_date DATE → TEXT
ALTER TABLE attendance_reminders ALTER COLUMN attendance_date TYPE TEXT;

-- holiday_definitions.date DATE → TEXT
ALTER TABLE holiday_definitions ALTER COLUMN date TYPE TEXT;

-- holiday_instances.actual_date, observed_date DATE → TEXT
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.tables WHERE table_name = 'holiday_instances'
    ) THEN
        ALTER TABLE holiday_instances ALTER COLUMN actual_date TYPE TEXT;
        ALTER TABLE holiday_instances ALTER COLUMN observed_date TYPE TEXT;
    END IF;
END $$;

-- permission_requests.permission_date DATE → TEXT
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'permission_requests' AND column_name = 'permission_date'
    ) THEN
        ALTER TABLE permission_requests ALTER COLUMN permission_date TYPE TEXT;
    END IF;
END $$;
