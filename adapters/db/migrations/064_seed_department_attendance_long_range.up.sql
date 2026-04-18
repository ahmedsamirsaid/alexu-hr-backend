-- Seed a long attendance timeline (180 days) for department dashboard trend testing.
-- Builds on employees/devices created by migration 056 (emp56_% / adev56_%).

WITH RECURSIVE days(day_date, day_index) AS (
    VALUES (date('now', '-179 days'), 0)
    UNION ALL
    SELECT date(day_date, '+1 day'), day_index + 1
    FROM days
    WHERE day_index < 179
),
workdays AS (
    SELECT day_date, day_index
    FROM days
    WHERE strftime('%w', day_date) NOT IN ('5', '6')
),
seed_employees AS (
    SELECT
        e.uid AS employee_uid,
        e.department_uid,
        e.university_id,
        CAST(substr(e.uid, -2) AS INTEGER) AS employee_slot,
        CASE
            WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
            WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
            ELSE 'adev56_' || lower(substr(e.department_uid, 6))
        END AS device_uid
    FROM employees e
    WHERE e.uid LIKE 'emp56_%'
      AND e.status = 'active'
),
base_rows AS (
    SELECT
        se.employee_uid,
        se.department_uid,
        se.university_id,
        se.device_uid,
        se.employee_slot,
        wd.day_date,
        wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 8 = 0) AS is_missed_punch_out,
        ((wd.day_index + se.employee_slot) % 13 = 0) AS is_missed_punch_in,
        ((wd.day_index + se.employee_slot) % 9 = 0) AS is_late_arrival,
        ((wd.day_index + se.employee_slot) % 17 = 0) AS is_early_departure
    FROM seed_employees se
    CROSS JOIN workdays wd
)
INSERT OR IGNORE INTO attendance_records (
    uid,
    employee_uid,
    device_uid,
    device_user_id,
    punched_at,
    punch_type,
    raw_payload,
    created_at,
    updated_at
)
SELECT
    'atr60_in_' || br.employee_uid || '_' || replace(br.day_date, '-', ''),
    br.employee_uid,
    br.device_uid,
    br.university_id,
    (br.day_date || 'T' || CASE
        WHEN br.is_late_arrival THEN '09:24:00'
        WHEN ((br.day_index + br.employee_slot) % 5 = 0) THEN '09:03:00'
        ELSE '08:47:00'
    END || 'Z'),
    'check_in',
    '{"source":"seed60","case":"long_range"}',
    datetime('now'),
    datetime('now')
FROM base_rows br
WHERE br.is_absent = 0
  AND br.is_missed_punch_in = 0;

WITH RECURSIVE days(day_date, day_index) AS (
    VALUES (date('now', '-179 days'), 0)
    UNION ALL
    SELECT date(day_date, '+1 day'), day_index + 1
    FROM days
    WHERE day_index < 179
),
workdays AS (
    SELECT day_date, day_index
    FROM days
    WHERE strftime('%w', day_date) NOT IN ('5', '6')
),
seed_employees AS (
    SELECT
        e.uid AS employee_uid,
        e.department_uid,
        e.university_id,
        CAST(substr(e.uid, -2) AS INTEGER) AS employee_slot,
        CASE
            WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
            WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
            ELSE 'adev56_' || lower(substr(e.department_uid, 6))
        END AS device_uid
    FROM employees e
    WHERE e.uid LIKE 'emp56_%'
      AND e.status = 'active'
),
base_rows AS (
    SELECT
        se.employee_uid,
        se.department_uid,
        se.university_id,
        se.device_uid,
        se.employee_slot,
        wd.day_date,
        wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 8 = 0) AS is_missed_punch_out,
        ((wd.day_index + se.employee_slot) % 13 = 0) AS is_missed_punch_in,
        ((wd.day_index + se.employee_slot) % 9 = 0) AS is_late_arrival,
        ((wd.day_index + se.employee_slot) % 17 = 0) AS is_early_departure
    FROM seed_employees se
    CROSS JOIN workdays wd
)
INSERT OR IGNORE INTO attendance_records (
    uid,
    employee_uid,
    device_uid,
    device_user_id,
    punched_at,
    punch_type,
    raw_payload,
    created_at,
    updated_at
)
SELECT
    'atr60_out_' || br.employee_uid || '_' || replace(br.day_date, '-', ''),
    br.employee_uid,
    br.device_uid,
    br.university_id,
    (br.day_date || 'T' || CASE
        WHEN br.is_early_departure THEN '15:38:00'
        WHEN ((br.day_index + br.employee_slot) % 6 = 0) THEN '17:19:00'
        ELSE '16:58:00'
    END || 'Z'),
    'check_out',
    '{"source":"seed60","case":"long_range"}',
    datetime('now'),
    datetime('now')
FROM base_rows br
WHERE br.is_absent = 0
  AND br.is_missed_punch_out = 0;

WITH RECURSIVE days(day_date, day_index) AS (
    VALUES (date('now', '-179 days'), 0)
    UNION ALL
    SELECT date(day_date, '+1 day'), day_index + 1
    FROM days
    WHERE day_index < 179
),
workdays AS (
    SELECT day_date, day_index
    FROM days
    WHERE strftime('%w', day_date) NOT IN ('5', '6')
),
seed_employees AS (
    SELECT
        e.uid AS employee_uid,
        CAST(substr(e.uid, -2) AS INTEGER) AS employee_slot
    FROM employees e
    WHERE e.uid LIKE 'emp56_%'
      AND e.status = 'active'
),
base_rows AS (
    SELECT
        se.employee_uid,
        se.employee_slot,
        wd.day_date,
        wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 8 = 0) AS is_missed_punch_out,
        ((wd.day_index + se.employee_slot) % 13 = 0) AS is_missed_punch_in,
        ((wd.day_index + se.employee_slot) % 9 = 0) AS is_late_arrival,
        ((wd.day_index + se.employee_slot) % 17 = 0) AS is_early_departure
    FROM seed_employees se
    CROSS JOIN workdays wd
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
    'aex60_abs_' || br.employee_uid || '_' || replace(br.day_date, '-', ''),
    br.employee_uid,
    br.day_date,
    'absence',
    NULL,
    NULL,
    NULL,
    NULL,
    datetime('now'),
    datetime('now')
FROM base_rows br
WHERE br.is_absent = 1;

WITH RECURSIVE days(day_date, day_index) AS (
    VALUES (date('now', '-179 days'), 0)
    UNION ALL
    SELECT date(day_date, '+1 day'), day_index + 1
    FROM days
    WHERE day_index < 179
),
workdays AS (
    SELECT day_date, day_index
    FROM days
    WHERE strftime('%w', day_date) NOT IN ('5', '6')
),
seed_employees AS (
    SELECT
        e.uid AS employee_uid,
        CAST(substr(e.uid, -2) AS INTEGER) AS employee_slot
    FROM employees e
    WHERE e.uid LIKE 'emp56_%'
      AND e.status = 'active'
),
base_rows AS (
    SELECT
        se.employee_uid,
        se.employee_slot,
        wd.day_date,
        wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 9 = 0) AS is_late_arrival
    FROM seed_employees se
    CROSS JOIN workdays wd
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
    'aex60_late_' || br.employee_uid || '_' || replace(br.day_date, '-', ''),
    br.employee_uid,
    br.day_date,
    'late_arrival',
    (br.day_date || 'T09:24:00Z'),
    (br.day_date || 'T16:58:00Z'),
    15,
    9,
    datetime('now'),
    datetime('now')
FROM base_rows br
WHERE br.is_absent = 0
  AND br.is_late_arrival = 1;

WITH RECURSIVE days(day_date, day_index) AS (
    VALUES (date('now', '-179 days'), 0)
    UNION ALL
    SELECT date(day_date, '+1 day'), day_index + 1
    FROM days
    WHERE day_index < 179
),
workdays AS (
    SELECT day_date, day_index
    FROM days
    WHERE strftime('%w', day_date) NOT IN ('5', '6')
),
seed_employees AS (
    SELECT
        e.uid AS employee_uid,
        CAST(substr(e.uid, -2) AS INTEGER) AS employee_slot
    FROM employees e
    WHERE e.uid LIKE 'emp56_%'
      AND e.status = 'active'
),
base_rows AS (
    SELECT
        se.employee_uid,
        se.employee_slot,
        wd.day_date,
        wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 8 = 0) AS is_missed_punch_out
    FROM seed_employees se
    CROSS JOIN workdays wd
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
    'aex60_mout_' || br.employee_uid || '_' || replace(br.day_date, '-', ''),
    br.employee_uid,
    br.day_date,
    'missed_punch_out',
    (br.day_date || 'T08:47:00Z'),
    NULL,
    15,
    NULL,
    datetime('now'),
    datetime('now')
FROM base_rows br
WHERE br.is_absent = 0
  AND br.is_missed_punch_out = 1;

WITH RECURSIVE days(day_date, day_index) AS (
    VALUES (date('now', '-179 days'), 0)
    UNION ALL
    SELECT date(day_date, '+1 day'), day_index + 1
    FROM days
    WHERE day_index < 179
),
workdays AS (
    SELECT day_date, day_index
    FROM days
    WHERE strftime('%w', day_date) NOT IN ('5', '6')
),
seed_employees AS (
    SELECT
        e.uid AS employee_uid,
        CAST(substr(e.uid, -2) AS INTEGER) AS employee_slot
    FROM employees e
    WHERE e.uid LIKE 'emp56_%'
      AND e.status = 'active'
),
base_rows AS (
    SELECT
        se.employee_uid,
        se.employee_slot,
        wd.day_date,
        wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 8 = 0) AS is_missed_punch_out,
        ((wd.day_index + se.employee_slot) % 13 = 0) AS is_missed_punch_in
    FROM seed_employees se
    CROSS JOIN workdays wd
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
    'aex60_min_' || br.employee_uid || '_' || replace(br.day_date, '-', ''),
    br.employee_uid,
    br.day_date,
    'missed_punch_in',
    NULL,
    (br.day_date || 'T16:58:00Z'),
    15,
    NULL,
    datetime('now'),
    datetime('now')
FROM base_rows br
WHERE br.is_absent = 0
  AND br.is_missed_punch_out = 0
  AND br.is_missed_punch_in = 1;

WITH RECURSIVE days(day_date, day_index) AS (
    VALUES (date('now', '-179 days'), 0)
    UNION ALL
    SELECT date(day_date, '+1 day'), day_index + 1
    FROM days
    WHERE day_index < 179
),
workdays AS (
    SELECT day_date, day_index
    FROM days
    WHERE strftime('%w', day_date) NOT IN ('5', '6')
),
seed_employees AS (
    SELECT
        e.uid AS employee_uid,
        CAST(substr(e.uid, -2) AS INTEGER) AS employee_slot
    FROM employees e
    WHERE e.uid LIKE 'emp56_%'
      AND e.status = 'active'
),
base_rows AS (
    SELECT
        se.employee_uid,
        se.employee_slot,
        wd.day_date,
        wd.day_index,
        ((wd.day_index + se.employee_slot) % 11 = 0) AS is_absent,
        ((wd.day_index + se.employee_slot) % 8 = 0) AS is_missed_punch_out,
        ((wd.day_index + se.employee_slot) % 13 = 0) AS is_missed_punch_in,
        ((wd.day_index + se.employee_slot) % 17 = 0) AS is_early_departure
    FROM seed_employees se
    CROSS JOIN workdays wd
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
    'aex60_early_' || br.employee_uid || '_' || replace(br.day_date, '-', ''),
    br.employee_uid,
    br.day_date,
    'early_departure',
    (br.day_date || 'T08:47:00Z'),
    (br.day_date || 'T15:38:00Z'),
    0,
    82,
    datetime('now'),
    datetime('now')
FROM base_rows br
WHERE br.is_absent = 0
  AND br.is_missed_punch_out = 0
  AND br.is_missed_punch_in = 0
  AND br.is_early_departure = 1;
