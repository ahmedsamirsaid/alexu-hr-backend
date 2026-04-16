-- Make the sample active so absences appear in the seeded logs.
UPDATE employees
SET status = 'active'
WHERE uid = 'emp_00000000000000000000000000000011';

-- Backfill the deterministic sample exceptions once, so the app reads them
-- from attendance_exceptions without recalculating on every logs request.
INSERT OR IGNORE INTO attendance_exceptions (
    uid,
    employee_uid,
    attendance_date,
    exception_type,
    check_in,
    check_out,
    grace_minutes,
    minutes_delta,
    created_at,
    updated_at
)
VALUES
    ('aex_seed_20260413_emp02_late', 'emp_00000000000000000000000000000002', '2026-04-13', 'late_arrival', '2026-04-13T09:18:00Z', '2026-04-13T17:06:00Z', 15, 3, datetime('now'), datetime('now')),
    ('aex_seed_20260414_emp07_early', 'emp_00000000000000000000000000000007', '2026-04-14', 'early_departure', '2026-04-14T08:35:00Z', '2026-04-14T16:40:00Z', 15, 5, datetime('now'), datetime('now')),
    ('aex_seed_20260415_emp02_late', 'emp_00000000000000000000000000000002', '2026-04-15', 'late_arrival', '2026-04-15T09:24:00Z', '2026-04-15T16:56:00Z', 15, 9, datetime('now'), datetime('now')),
    ('aex_seed_20260415_emp04_miss_out', 'emp_00000000000000000000000000000004', '2026-04-15', 'missed_punch_out', '2026-04-15T08:58:00Z', NULL, 15, NULL, datetime('now'), datetime('now')),
    ('aex_seed_20260415_emp07_early', 'emp_00000000000000000000000000000007', '2026-04-15', 'early_departure', '2026-04-15T08:28:00Z', '2026-04-15T16:21:00Z', 15, 24, datetime('now'), datetime('now'));

WITH RECURSIVE sample_days(day) AS (
    VALUES(date('2026-04-13'))
    UNION ALL
    SELECT date(day, '+1 day')
    FROM sample_days
    WHERE day < date('2026-04-15')
)
INSERT OR IGNORE INTO attendance_exceptions (
    uid,
    employee_uid,
    attendance_date,
    exception_type,
    check_in,
    check_out,
    grace_minutes,
    minutes_delta,
    created_at,
    updated_at
)
SELECT
    'aex_seed_abs_' || replace(day, '-', '') || '_' || e.uid,
    e.uid,
    day,
    'absence',
    NULL,
    NULL,
    NULL,
    NULL,
    datetime('now'),
    datetime('now')
FROM sample_days
JOIN employees e
    ON e.status = 'active'
   AND date(e.hire_date) <= day
WHERE NOT EXISTS (
    SELECT 1
    FROM attendance_records ar
    WHERE ar.employee_uid = e.uid
      AND date(ar.punched_at) = day
)
AND NOT EXISTS (
    SELECT 1
    FROM leave_records lr
    WHERE lr.employee_id = e.id
      AND date(day) BETWEEN date(lr.start_date) AND date(lr.end_date)
);
