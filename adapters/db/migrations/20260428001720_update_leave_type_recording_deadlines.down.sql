UPDATE leave_types
SET recording_deadline_days = NULL;

UPDATE leave_types
SET recording_deadline_days = 2
WHERE code = 'CASUAL';

UPDATE leave_types
SET recording_deadline_days = 30
WHERE code = 'REGULAR';
