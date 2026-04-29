UPDATE leave_types
SET recording_deadline_days = 30;

UPDATE leave_types
SET recording_deadline_days = 2
WHERE code IN ('CASUAL', 'REGULAR');

UPDATE leave_types
SET max_consecutive = 90
WHERE code = 'REGULAR';

UPDATE leave_types
SET max_consecutive = 730
WHERE code = 'CHILD_CARE';

UPDATE leave_types
SET default_balance = 120
WHERE code = 'MATERNITY';

UPDATE leave_types
SET default_balance = 365
WHERE code = 'SPECIAL_PAID' OR code = 'SPECIAL_UNPAID';

UPDATE leave_types
SET max_consecutive = 365
WHERE code != 'REGULAR' AND code != 'CASUAL' AND code != 'CHILD_CARE';