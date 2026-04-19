-- Remove historical dashboard-oriented dummy seed data.
-- Keep dashboard values driven by real source tables and explicit fixed-date seeds.

DELETE FROM approval_actions
WHERE uid LIKE 'apa60_%';

DELETE FROM leave_requests
WHERE uid LIKE 'lrq60_%';

DELETE FROM approval_requests
WHERE uid LIKE 'apr60_%';

DELETE FROM leave_records
WHERE uid LIKE 'lrec60_%';

DELETE FROM attendance_exceptions
WHERE uid LIKE 'aex60_%';

DELETE FROM attendance_records
WHERE uid LIKE 'atr60_%';
