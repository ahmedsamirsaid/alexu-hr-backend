-- Expand attendance reminder types to include 'missing_check_in' and 'missing_check_out'
-- PostgreSQL allows dropping and recreating constraints

-- Drop the old constraint
ALTER TABLE attendance_reminders 
DROP CONSTRAINT IF EXISTS attendance_reminders_reminder_type_check;

-- Add the new constraint with both reminder types
ALTER TABLE attendance_reminders
ADD CONSTRAINT attendance_reminders_reminder_type_check 
CHECK (reminder_type IN ('missing_check_in', 'missing_check_out'));

