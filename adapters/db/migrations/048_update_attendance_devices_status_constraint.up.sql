-- Update attendance_devices status constraint to include 'deactivated'
-- PostgreSQL allows dropping and recreating constraints

-- Drop the old constraint
ALTER TABLE attendance_devices 
DROP CONSTRAINT IF EXISTS attendance_devices_status_check;

-- Add the new constraint with 'deactivated' option
ALTER TABLE attendance_devices
ADD CONSTRAINT attendance_devices_status_check 
CHECK (status IN ('online', 'offline', 'deactivated'));
