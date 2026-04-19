-- Seed fixed-date dashboard metrics for 2026-04-19.
-- Adds attendance (check-in/check-out), leave records on the day,
-- and pending leave requests submitted on the day.
--
-- This migration is additive and idempotent.

-- ---------------------------------------------------------------------------
-- 1) Attendance on 2026-04-19
-- ---------------------------------------------------------------------------
WITH attendance_seed(uid, employee_uid, punched_at, punch_type) AS (
    VALUES
        ('atr65_20260419_hr01_in',  'emp56_HR_01',  '2026-04-19T08:44:00Z', 'check_in'),
        ('atr65_20260419_it01_in',  'emp56_IT_01',  '2026-04-19T08:52:00Z', 'check_in'),
        ('atr65_20260419_fin01_in', 'emp56_FIN_01', '2026-04-19T08:59:00Z', 'check_in'),
        ('atr65_20260419_ops01_in', 'emp56_OPS_01', '2026-04-19T09:05:00Z', 'check_in'),
        ('atr65_20260419_sec01_in', 'emp56_SEC_01', '2026-04-19T08:35:00Z', 'check_in'),
        ('atr65_20260419_adm01_in', 'emp56_ADM_01', '2026-04-19T08:48:00Z', 'check_in'),
        ('atr65_20260419_aca01_in', 'emp56_ACA_01', '2026-04-19T08:57:00Z', 'check_in'),
        ('atr65_20260419_lib01_in', 'emp56_LIB_01', '2026-04-19T09:02:00Z', 'check_in'),
        ('atr65_20260419_hr01_out',  'emp56_HR_01',  '2026-04-19T17:01:00Z', 'check_out'),
        ('atr65_20260419_it01_out',  'emp56_IT_01',  '2026-04-19T17:07:00Z', 'check_out'),
        ('atr65_20260419_fin01_out', 'emp56_FIN_01', '2026-04-19T16:55:00Z', 'check_out'),
        ('atr65_20260419_sec01_out', 'emp56_SEC_01', '2026-04-19T16:46:00Z', 'check_out'),
        ('atr65_20260419_aca01_out', 'emp56_ACA_01', '2026-04-19T17:12:00Z', 'check_out')
),
attendance_mapped AS (
    SELECT
        s.uid,
        e.uid AS employee_uid,
        CASE
            WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
            WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
            ELSE 'adev56_' || lower(substr(e.department_uid, 6))
        END AS device_uid,
        e.university_id AS device_user_id,
        s.punched_at,
        s.punch_type
    FROM attendance_seed s
    JOIN employees e ON e.uid = s.employee_uid
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
    m.uid,
    m.employee_uid,
    m.device_uid,
    m.device_user_id,
    m.punched_at,
    m.punch_type,
    '{"source":"seed65","date":"2026-04-19"}',
    datetime('now'),
    datetime('now')
FROM attendance_mapped m
JOIN attendance_devices d ON d.uid = m.device_uid;

-- ---------------------------------------------------------------------------
-- 2) Leave records active on 2026-04-19
-- ---------------------------------------------------------------------------
WITH leave_seed(uid, employee_uid, start_date, end_date, days, notes) AS (
    VALUES
        ('lrec65_20260419_hr03',  'emp56_HR_03',  '2026-04-19', '2026-04-19', 1, 'Seeded leave on 19 Apr 2026 (HR).'),
        ('lrec65_20260419_it03',  'emp56_IT_03',  '2026-04-19', '2026-04-19', 1, 'Seeded leave on 19 Apr 2026 (IT).'),
        ('lrec65_20260419_fin03', 'emp56_FIN_03', '2026-04-19', '2026-04-19', 1, 'Seeded leave on 19 Apr 2026 (FIN).')
)
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
    s.uid,
    e.id,
    lt.id,
    s.start_date,
    s.end_date,
    s.days,
    '2026-04-19T09:30:00Z',
    rb.id,
    s.notes,
    NULL,
    datetime('now'),
    datetime('now')
FROM leave_seed s
JOIN employees e ON e.uid = s.employee_uid
JOIN leave_types lt ON lt.code = 'CASUAL'
LEFT JOIN employees rb ON rb.uid = 'emp56_HR_01';

-- ---------------------------------------------------------------------------
-- 3) Pending requests submitted on 2026-04-19
-- ---------------------------------------------------------------------------
WITH request_seed(approval_uid, leave_uid, action_uid, employee_uid, start_date, end_date, days, notes, submitted_at) AS (
    VALUES
        ('apr65_20260419_ops02', 'lrq65_20260419_ops02', 'apa65_20260419_ops02', 'emp56_OPS_02', '2026-04-23', '2026-04-24', 2, 'Seeded pending request from OPS for dashboard checks.', '2026-04-19T10:15:00Z'),
        ('apr65_20260419_sec02', 'lrq65_20260419_sec02', 'apa65_20260419_sec02', 'emp56_SEC_02', '2026-04-24', '2026-04-24', 1, 'Seeded pending request from SEC for dashboard checks.', '2026-04-19T10:32:00Z'),
        ('apr65_20260419_adm02', 'lrq65_20260419_adm02', 'apa65_20260419_adm02', 'emp56_ADM_02', '2026-04-27', '2026-04-29', 3, 'Seeded pending request from ADM for dashboard checks.', '2026-04-19T11:05:00Z'),
        ('apr65_20260419_lib02', 'lrq65_20260419_lib02', 'apa65_20260419_lib02', 'emp56_LIB_02', '2026-04-25', '2026-04-26', 2, 'Seeded pending request from LIB for dashboard checks.', '2026-04-19T11:20:00Z')
)
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
SELECT
    s.approval_uid,
    af.uid,
    s.employee_uid,
    1,
    1,
    'pending',
    s.submitted_at,
    s.submitted_at
FROM request_seed s
JOIN employees e ON e.uid = s.employee_uid
JOIN approval_flows af ON af.uid = 'apf_leave_default';

WITH request_seed(approval_uid, leave_uid, action_uid, employee_uid, start_date, end_date, days, notes, submitted_at) AS (
    VALUES
        ('apr65_20260419_ops02', 'lrq65_20260419_ops02', 'apa65_20260419_ops02', 'emp56_OPS_02', '2026-04-23', '2026-04-24', 2, 'Seeded pending request from OPS for dashboard checks.', '2026-04-19T10:15:00Z'),
        ('apr65_20260419_sec02', 'lrq65_20260419_sec02', 'apa65_20260419_sec02', 'emp56_SEC_02', '2026-04-24', '2026-04-24', 1, 'Seeded pending request from SEC for dashboard checks.', '2026-04-19T10:32:00Z'),
        ('apr65_20260419_adm02', 'lrq65_20260419_adm02', 'apa65_20260419_adm02', 'emp56_ADM_02', '2026-04-27', '2026-04-29', 3, 'Seeded pending request from ADM for dashboard checks.', '2026-04-19T11:05:00Z'),
        ('apr65_20260419_lib02', 'lrq65_20260419_lib02', 'apa65_20260419_lib02', 'emp56_LIB_02', '2026-04-25', '2026-04-26', 2, 'Seeded pending request from LIB for dashboard checks.', '2026-04-19T11:20:00Z')
)
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
    s.leave_uid,
    s.employee_uid,
    lt.uid,
    s.start_date,
    s.end_date,
    s.days,
    s.notes,
    s.submitted_at,
    NULL,
    s.approval_uid,
    s.submitted_at,
    s.submitted_at
FROM request_seed s
JOIN employees e ON e.uid = s.employee_uid
JOIN leave_types lt ON lt.code = 'ANNUAL'
JOIN approval_requests ar ON ar.uid = s.approval_uid;

WITH request_seed(approval_uid, leave_uid, action_uid, employee_uid, start_date, end_date, days, notes, submitted_at) AS (
    VALUES
        ('apr65_20260419_ops02', 'lrq65_20260419_ops02', 'apa65_20260419_ops02', 'emp56_OPS_02', '2026-04-23', '2026-04-24', 2, 'Seeded pending request from OPS for dashboard checks.', '2026-04-19T10:15:00Z'),
        ('apr65_20260419_sec02', 'lrq65_20260419_sec02', 'apa65_20260419_sec02', 'emp56_SEC_02', '2026-04-24', '2026-04-24', 1, 'Seeded pending request from SEC for dashboard checks.', '2026-04-19T10:32:00Z'),
        ('apr65_20260419_adm02', 'lrq65_20260419_adm02', 'apa65_20260419_adm02', 'emp56_ADM_02', '2026-04-27', '2026-04-29', 3, 'Seeded pending request from ADM for dashboard checks.', '2026-04-19T11:05:00Z'),
        ('apr65_20260419_lib02', 'lrq65_20260419_lib02', 'apa65_20260419_lib02', 'emp56_LIB_02', '2026-04-25', '2026-04-26', 2, 'Seeded pending request from LIB for dashboard checks.', '2026-04-19T11:20:00Z')
)
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
SELECT
    s.action_uid,
    s.approval_uid,
    'submit',
    NULL,
    s.employee_uid,
    'Seeded submit action for pending request (2026-04-19).',
    s.submitted_at,
    s.submitted_at
FROM request_seed s
JOIN employees e ON e.uid = s.employee_uid
JOIN approval_requests ar ON ar.uid = s.approval_uid;
