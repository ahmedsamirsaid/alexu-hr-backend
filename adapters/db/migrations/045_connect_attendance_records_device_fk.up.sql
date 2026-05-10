-- Add foreign key constraint to attendance_records.device_uid
-- PostgreSQL allows adding constraints without recreating the table

-- First, ensure the foreign key constraint doesn't already exist
ALTER TABLE attendance_records 
DROP CONSTRAINT IF EXISTS attendance_records_device_uid_fkey;

-- Add the foreign key constraint
ALTER TABLE attendance_records
ADD CONSTRAINT attendance_records_device_uid_fkey 
FOREIGN KEY (device_uid) REFERENCES attendance_devices(uid) ON DELETE CASCADE;

