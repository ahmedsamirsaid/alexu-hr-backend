-- Dashboard runtime seed for realistic current-day metrics.
-- Additive + idempotent to support repeated local test resets.
--
-- Target signal for dashboard testing:
-- - Checked in today: +8 employees
-- - Leaves today: +3 employees
-- - Pending requests: +4 requests

-- ---------------------------------------------------------------------------
-- 1) Today's attendance check-ins (distinct employees)
-- ---------------------------------------------------------------------------
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
VALUES
    ('atr60_today_emp01_in', 'emp_00000000000000000000000000000001', 'adev_main_gate', 'UNI001', strftime('%Y-%m-%dT08:42:00Z', 'now'), 'check_in', '{"source":"seed60","note":"manager-arrival"}', datetime('now'), datetime('now')),
    ('atr60_today_emp03_in', 'emp_00000000000000000000000000000003', 'adev_seed_013', 'UNI003', strftime('%Y-%m-%dT08:55:00Z', 'now'), 'check_in', '{"source":"seed60","note":"it-arrival"}', datetime('now'), datetime('now')),
    ('atr60_today_emp04_in', 'emp_00000000000000000000000000000004', 'adev_seed_005', 'UNI004', strftime('%Y-%m-%dT09:06:00Z', 'now'), 'check_in', '{"source":"seed60","note":"finance-late"}', datetime('now'), datetime('now')),
    ('atr60_today_emp06_in', 'emp_00000000000000000000000000000006', 'adev_seed_014', 'UNI006', strftime('%Y-%m-%dT08:49:00Z', 'now'), 'check_in', '{"source":"seed60","note":"hr-annex"}', datetime('now'), datetime('now')),
    ('atr60_today_emp07_in', 'emp_00000000000000000000000000000007', 'adev_seed_020', 'UNI007', strftime('%Y-%m-%dT08:31:00Z', 'now'), 'check_in', '{"source":"seed60","note":"security-shift"}', datetime('now'), datetime('now')),
    ('atr60_today_emp10_in', 'emp_00000000000000000000000000000010', 'adev_seed_011', 'UNI010', strftime('%Y-%m-%dT08:58:00Z', 'now'), 'check_in', '{"source":"seed60","note":"academic-staff"}', datetime('now'), datetime('now')),
    ('atr60_today_emp11_unknown', 'emp_00000000000000000000000000000011', 'adev_seed_012', 'UNI011', strftime('%Y-%m-%dT09:01:00Z', 'now'), 'unknown', '{"source":"seed60","note":"unknown-punch-counted"}', datetime('now'), datetime('now')),
    ('atr60_today_emp12_in', 'emp_00000000000000000000000000000012', 'adev_main_gate', 'UNI012', strftime('%Y-%m-%dT08:46:00Z', 'now'), 'check_in', '{"source":"seed60","note":"library-staff"}', datetime('now'), datetime('now'));

-- Optional check-outs to keep the attendance picture realistic.
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
VALUES
    ('atr60_today_emp01_out', 'emp_00000000000000000000000000000001', 'adev_main_gate', 'UNI001', strftime('%Y-%m-%dT17:04:00Z', 'now'), 'check_out', '{"source":"seed60","note":"full-day"}', datetime('now'), datetime('now')),
    ('atr60_today_emp03_out', 'emp_00000000000000000000000000000003', 'adev_seed_013', 'UNI003', strftime('%Y-%m-%dT17:11:00Z', 'now'), 'check_out', '{"source":"seed60","note":"full-day"}', datetime('now'), datetime('now')),
    ('atr60_today_emp07_out', 'emp_00000000000000000000000000000007', 'adev_seed_020', 'UNI007', strftime('%Y-%m-%dT16:44:00Z', 'now'), 'check_out', '{"source":"seed60","note":"field-shift"}', datetime('now'), datetime('now'));

-- ---------------------------------------------------------------------------
-- 2) Leave records active today (distinct employees)
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO leave_records (
    uid,
    employee_id,
    leave_type_id,
    start_date,
    end_date,
    days,
    recorded_at,
    recorded_by,
    notes,
    leave_request_uid,
    created_at,
    updated_at
)
SELECT
    'lrec60_today_emp02',
    e.id,
    lt.id,
    date('now'),
    date('now'),
    1,
    datetime('now', '-4 hours'),
    rb.id,
    'Seed runtime: one-day leave for dashboard realism',
    NULL,
    datetime('now'),
    datetime('now')
FROM employees e
JOIN (SELECT id FROM leave_types WHERE code = 'CASUAL' ORDER BY id LIMIT 1) lt
LEFT JOIN employees rb ON rb.uid = 'emp_00000000000000000000000000000001'
WHERE e.uid = 'emp_00000000000000000000000000000002';

INSERT OR IGNORE INTO leave_records (
    uid,
    employee_id,
    leave_type_id,
    start_date,
    end_date,
    days,
    recorded_at,
    recorded_by,
    notes,
    leave_request_uid,
    created_at,
    updated_at
)
SELECT
    'lrec60_today_emp05',
    e.id,
    lt.id,
    date('now'),
    date('now'),
    1,
    datetime('now', '-3 hours'),
    rb.id,
    'Seed runtime: short urgent leave',
    NULL,
    datetime('now'),
    datetime('now')
FROM employees e
JOIN (SELECT id FROM leave_types WHERE code = 'CASUAL' ORDER BY id LIMIT 1) lt
LEFT JOIN employees rb ON rb.uid = 'emp_00000000000000000000000000000001'
WHERE e.uid = 'emp_00000000000000000000000000000005';

INSERT OR IGNORE INTO leave_records (
    uid,
    employee_id,
    leave_type_id,
    start_date,
    end_date,
    days,
    recorded_at,
    recorded_by,
    notes,
    leave_request_uid,
    created_at,
    updated_at
)
SELECT
    'lrec60_today_emp08',
    e.id,
    lt.id,
    date('now'),
    date('now'),
    1,
    datetime('now', '-2 hours'),
    rb.id,
    'Seed runtime: personal leave day',
    NULL,
    datetime('now'),
    datetime('now')
FROM employees e
JOIN (SELECT id FROM leave_types WHERE code = 'CASUAL' ORDER BY id LIMIT 1) lt
LEFT JOIN employees rb ON rb.uid = 'emp_00000000000000000000000000000001'
WHERE e.uid = 'emp_00000000000000000000000000000008';

-- ---------------------------------------------------------------------------
-- 3) Pending leave requests (approval + request + submit action)
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO approval_requests (
    uid,
    approval_flow_uid,
    requester_uid,
    current_step,
    max_step,
    status,
    created_at,
    updated_at
)
VALUES
    ('apr60_pending_emp03', 'apf_leave_default', 'emp_00000000000000000000000000000003', 1, 1, 'pending', datetime('now', '-3 days'), datetime('now', '-3 days')),
    ('apr60_pending_emp04', 'apf_leave_default', 'emp_00000000000000000000000000000004', 1, 1, 'pending', datetime('now', '-2 days'), datetime('now', '-2 days')),
    ('apr60_pending_emp06', 'apf_leave_default', 'emp_00000000000000000000000000000006', 1, 1, 'pending', datetime('now', '-36 hours'), datetime('now', '-36 hours')),
    ('apr60_pending_emp10', 'apf_leave_default', 'emp_00000000000000000000000000000010', 1, 1, 'pending', datetime('now', '-24 hours'), datetime('now', '-24 hours'));

INSERT OR IGNORE INTO leave_requests (
    uid,
    employee_uid,
    leave_type_uid,
    start_date,
    end_date,
    days,
    notes,
    submitted_at,
    decided_at,
    approval_request_uid,
    created_at,
    updated_at
)
SELECT
    'lrq60_pending_emp03',
    'emp_00000000000000000000000000000003',
    lt.uid,
    date('now', '+2 days'),
    date('now', '+3 days'),
    2,
    'Seed runtime: pending annual leave request',
    datetime('now', '-3 days'),
    NULL,
    'apr60_pending_emp03',
    datetime('now', '-3 days'),
    datetime('now', '-3 days')
FROM leave_types lt
WHERE lt.code = 'ANNUAL'
ORDER BY lt.id
LIMIT 1;

INSERT OR IGNORE INTO leave_requests (
    uid,
    employee_uid,
    leave_type_uid,
    start_date,
    end_date,
    days,
    notes,
    submitted_at,
    decided_at,
    approval_request_uid,
    created_at,
    updated_at
)
SELECT
    'lrq60_pending_emp04',
    'emp_00000000000000000000000000000004',
    lt.uid,
    date('now', '+5 days'),
    date('now', '+5 days'),
    1,
    'Seed runtime: pending one-day request',
    datetime('now', '-2 days'),
    NULL,
    'apr60_pending_emp04',
    datetime('now', '-2 days'),
    datetime('now', '-2 days')
FROM leave_types lt
WHERE lt.code = 'ANNUAL'
ORDER BY lt.id
LIMIT 1;

INSERT OR IGNORE INTO leave_requests (
    uid,
    employee_uid,
    leave_type_uid,
    start_date,
    end_date,
    days,
    notes,
    submitted_at,
    decided_at,
    approval_request_uid,
    created_at,
    updated_at
)
SELECT
    'lrq60_pending_emp06',
    'emp_00000000000000000000000000000006',
    lt.uid,
    date('now', '+7 days'),
    date('now', '+8 days'),
    2,
    'Seed runtime: pending request awaiting manager',
    datetime('now', '-36 hours'),
    NULL,
    'apr60_pending_emp06',
    datetime('now', '-36 hours'),
    datetime('now', '-36 hours')
FROM leave_types lt
WHERE lt.code = 'ANNUAL'
ORDER BY lt.id
LIMIT 1;

INSERT OR IGNORE INTO leave_requests (
    uid,
    employee_uid,
    leave_type_uid,
    start_date,
    end_date,
    days,
    notes,
    submitted_at,
    decided_at,
    approval_request_uid,
    created_at,
    updated_at
)
SELECT
    'lrq60_pending_emp10',
    'emp_00000000000000000000000000000010',
    lt.uid,
    date('now', '+10 days'),
    date('now', '+11 days'),
    2,
    'Seed runtime: pending request for next week',
    datetime('now', '-24 hours'),
    NULL,
    'apr60_pending_emp10',
    datetime('now', '-24 hours'),
    datetime('now', '-24 hours')
FROM leave_types lt
WHERE lt.code = 'ANNUAL'
ORDER BY lt.id
LIMIT 1;

INSERT OR IGNORE INTO approval_actions (
    uid,
    approval_request_uid,
    action,
    step_order,
    actor_uid,
    comments,
    acted_at,
    created_at
)
VALUES
    ('apa60_submit_emp03', 'apr60_pending_emp03', 'submit', NULL, 'emp_00000000000000000000000000000003', 'Seed runtime submit', datetime('now', '-3 days'), datetime('now', '-3 days')),
    ('apa60_submit_emp04', 'apr60_pending_emp04', 'submit', NULL, 'emp_00000000000000000000000000000004', 'Seed runtime submit', datetime('now', '-2 days'), datetime('now', '-2 days')),
    ('apa60_submit_emp06', 'apr60_pending_emp06', 'submit', NULL, 'emp_00000000000000000000000000000006', 'Seed runtime submit', datetime('now', '-36 hours'), datetime('now', '-36 hours')),
    ('apa60_submit_emp10', 'apr60_pending_emp10', 'submit', NULL, 'emp_00000000000000000000000000000010', 'Seed runtime submit', datetime('now', '-24 hours'), datetime('now', '-24 hours'));
