-- Rollback for 060_seed_department_attendance_long_range.up.sql

DELETE FROM attendance_exceptions
WHERE uid LIKE 'aex60_%';

DELETE FROM attendance_records
WHERE uid LIKE 'atr60_%';
