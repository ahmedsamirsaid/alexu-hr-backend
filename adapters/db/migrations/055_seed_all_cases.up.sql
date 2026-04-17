-- Seed data that covers all explicit status/enum branches used by the backend.
-- This migration is additive and idempotent.

-- ---------------------------------------------------------------------------
-- Employee status variants
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO employees (
    uid,
    name,
    mobile,
    government_id,
    university_id,
    email,
    hire_date,
    status,
    department_uid,
    shift_uid
)
VALUES
    ('emp_case_inactive_000000000000000000001', 'Case Employee Inactive', '+201199100001', '29901010000001', 'UNICASE001', 'case.inactive@university.edu.eg', '2024-01-15', 'inactive', 'dept_admin', 'shf_general_seed'),
    ('emp_case_terminated_0000000000000000001', 'Case Employee Terminated', '+201199100002', '29901010000002', 'UNICASE002', 'case.terminated@university.edu.eg', '2023-06-01', 'terminated', 'dept_ops', 'shf_general_seed');

-- ---------------------------------------------------------------------------
-- User/account and role-scope variants
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO users (uid, phone, employee_uid, is_active, created_at, updated_at)
VALUES
    ('usr_case_active_0000000000000000000001', '+201199200001', 'emp_00000000000000000000000000000001', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('usr_case_inactive_000000000000000000001', '+201199200002', 'emp_case_inactive_000000000000000000001', 0, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('usr_case_mgr_it_0000000000000000000001', '+201199200003', 'emp_00000000000000000000000000000003', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT
    u.id,
    r.id,
    'dept_it',
    CURRENT_TIMESTAMP
FROM users u
JOIN roles r ON r.uid = 'role_department_manager'
WHERE u.uid = 'usr_case_mgr_it_0000000000000000000001';

INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT
    u.id,
    r.id,
    NULL,
    CURRENT_TIMESTAMP
FROM users u
JOIN roles r ON r.uid = 'role_employee'
WHERE u.uid = 'usr_case_active_0000000000000000000001';

-- ---------------------------------------------------------------------------
-- Device token platform variants
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO device_tokens (uid, user_uid, token, platform, created_at, updated_at)
VALUES
    ('dtok_case_android_00000000000000000001', 'usr_case_active_0000000000000000000001', 'seed-device-token-android-case-001', 'android', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('dtok_case_ios_000000000000000000000001', 'usr_case_active_0000000000000000000001', 'seed-device-token-ios-case-001', 'ios', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- ---------------------------------------------------------------------------
-- OTP variants: valid-unused, valid-used, expired
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO otp_codes (id, phone, code, expires_at, used, created_at)
VALUES
    (900001, '+201199300001', '123456', datetime('now', '+20 minutes'), 0, datetime('now')),
    (900002, '+201199300002', '654321', datetime('now', '+20 minutes'), 1, datetime('now')),
    (900003, '+201199300003', '111111', datetime('now', '-20 minutes'), 0, datetime('now'));

-- ---------------------------------------------------------------------------
-- Refresh token variants: active, revoked, expired
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO refresh_tokens (token_hash, user_id, expires_at, revoked, created_at)
SELECT
    'seedhash_case_refresh_active_001',
    u.id,
    datetime('now', '+30 days'),
    0,
    datetime('now')
FROM users u
WHERE u.uid = 'usr_case_active_0000000000000000000001';

INSERT OR IGNORE INTO refresh_tokens (token_hash, user_id, expires_at, revoked, created_at)
SELECT
    'seedhash_case_refresh_revoked_001',
    u.id,
    datetime('now', '+30 days'),
    1,
    datetime('now')
FROM users u
WHERE u.uid = 'usr_case_active_0000000000000000000001';

INSERT OR IGNORE INTO refresh_tokens (token_hash, user_id, expires_at, revoked, created_at)
SELECT
    'seedhash_case_refresh_expired_001',
    u.id,
    datetime('now', '-30 days'),
    0,
    datetime('now')
FROM users u
WHERE u.uid = 'usr_case_active_0000000000000000000001';

-- ---------------------------------------------------------------------------
-- Attendance device status variants
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO attendance_devices (uid, ip, port, name, location, serial_number, status)
VALUES
    ('adev_case_online_000000000000000000001', '192.168.1.240', 4370, 'Case Online Device', 'Case Lab', 'ZK-CASE-ONLINE-001', 'online'),
    ('adev_case_offline_00000000000000000001', '192.168.1.241', 4370, 'Case Offline Device', 'Case Lab', 'ZK-CASE-OFFLINE-001', 'offline'),
    ('adev_case_deact_0000000000000000000001', '192.168.1.242', 4370, 'Case Deactivated Device', 'Case Lab', 'ZK-CASE-DEACT-001', 'deactivated');

-- ---------------------------------------------------------------------------
-- Attendance exception variants (all enum values)
-- ---------------------------------------------------------------------------
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
    ('aex_case_20260420_missed_in',  'emp_00000000000000000000000000000001', '2026-04-20', 'missed_punch_in',  NULL, '2026-04-20T17:03:00Z', 15, NULL, datetime('now'), datetime('now')),
    ('aex_case_20260420_missed_out', 'emp_00000000000000000000000000000002', '2026-04-20', 'missed_punch_out', '2026-04-20T09:01:00Z', NULL, 15, NULL, datetime('now'), datetime('now')),
    ('aex_case_20260420_late',       'emp_00000000000000000000000000000003', '2026-04-20', 'late_arrival',    '2026-04-20T09:22:00Z', '2026-04-20T17:06:00Z', 15, 7, datetime('now'), datetime('now')),
    ('aex_case_20260420_early',      'emp_00000000000000000000000000000004', '2026-04-20', 'early_departure', '2026-04-20T08:55:00Z', '2026-04-20T16:34:00Z', 15, 11, datetime('now'), datetime('now')),
    ('aex_case_20260420_absence',    'emp_00000000000000000000000000000005', '2026-04-20', 'absence',         NULL, NULL, NULL, NULL, datetime('now'), datetime('now'));

-- ---------------------------------------------------------------------------
-- Additional approval flow to represent multi-step pending state
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO approval_flows (uid, code, name_en, name_ar, description, is_active)
VALUES (
    'apf_case_two_step',
    'case_two_step',
    'Case Two-Step Flow',
    'تدفق اعتمادات تجريبي من مرحلتين',
    'Seed flow to cover current_step progression and pending-at-step-2 cases',
    1
);

INSERT OR IGNORE INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
VALUES
    ('afs_case_two_step_1', 'apf_case_two_step', 1, 'role_department_manager'),
    ('afs_case_two_step_2', 'apf_case_two_step', 2, 'role_department_manager');

-- ---------------------------------------------------------------------------
-- Approval request status variants
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
VALUES
    ('apr_case_pending_0000000000000000000001',   'apf_leave_default', 'emp_00000000000000000000000000000002', 1, 1, 'pending',   datetime('now', '-2 days'), datetime('now', '-2 days')),
    ('apr_case_approved_00000000000000000000001', 'apf_leave_default', 'emp_00000000000000000000000000000003', 1, 1, 'approved',  datetime('now', '-10 days'), datetime('now', '-9 days')),
    ('apr_case_rejected_00000000000000000000001', 'apf_leave_default', 'emp_00000000000000000000000000000004', 1, 1, 'rejected',  datetime('now', '-8 days'), datetime('now', '-7 days')),
    ('apr_case_cancelled_0000000000000000000001', 'apf_leave_default', 'emp_00000000000000000000000000000005', 1, 1, 'cancelled', datetime('now', '-6 days'), datetime('now', '-5 days')),
    ('apr_case_pending_step2_000000000000000001', 'apf_case_two_step', 'emp_00000000000000000000000000000006', 2, 2, 'pending',   datetime('now', '-1 days'), datetime('now', '-1 days'));

-- ---------------------------------------------------------------------------
-- Leave request variants linked to each approval status
-- ---------------------------------------------------------------------------
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
VALUES
    ('lrq_case_pending_0000000000000000000001',   'emp_00000000000000000000000000000002', 'ltype_00000000000000000000000000000002', '2026-05-10', '2026-05-10', 1, 'Pending approval case', datetime('now', '-2 days'), NULL,                    'apr_case_pending_0000000000000000000001',   datetime('now', '-2 days'), datetime('now', '-2 days')),
    ('lrq_case_approved_0000000000000000000001',  'emp_00000000000000000000000000000003', 'ltype_00000000000000000000000000000002', '2026-05-12', '2026-05-13', 2, 'Approved request case', datetime('now', '-10 days'), datetime('now', '-9 days'), 'apr_case_approved_00000000000000000000001', datetime('now', '-10 days'), datetime('now', '-9 days')),
    ('lrq_case_rejected_0000000000000000000001',  'emp_00000000000000000000000000000004', 'ltype_00000000000000000000000000000003', '2026-05-14', '2026-05-14', 1, 'Rejected request case', datetime('now', '-8 days'), datetime('now', '-7 days'), 'apr_case_rejected_00000000000000000000001', datetime('now', '-8 days'), datetime('now', '-7 days')),
    ('lrq_case_cancelled_0000000000000000000001', 'emp_00000000000000000000000000000005', 'ltype_00000000000000000000000000000001', '2026-05-15', '2026-05-15', 1, 'Cancelled request case', datetime('now', '-6 days'), datetime('now', '-5 days'), 'apr_case_cancelled_0000000000000000000001', datetime('now', '-6 days'), datetime('now', '-5 days')),
    ('lrq_case_pending_step2_000000000000000001', 'emp_00000000000000000000000000000006', 'ltype_00000000000000000000000000000002', '2026-05-16', '2026-05-16', 1, 'Pending at step 2 case', datetime('now', '-1 days'), NULL,                    'apr_case_pending_step2_000000000000000001', datetime('now', '-1 days'), datetime('now', '-1 days'));

-- ---------------------------------------------------------------------------
-- Approval action variants (all action types)
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO approval_actions (uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at)
VALUES
    ('apa_case_submit_pending_0000000000000001',   'apr_case_pending_0000000000000000000001',   'submit',  NULL, 'emp_00000000000000000000000000000002', 'Submitted request',                 datetime('now', '-2 days'), datetime('now', '-2 days')),
    ('apa_case_submit_approved_0000000000000001',  'apr_case_approved_00000000000000000000001', 'submit',  NULL, 'emp_00000000000000000000000000000003', 'Submitted approved flow request',   datetime('now', '-10 days'), datetime('now', '-10 days')),
    ('apa_case_approve_000000000000000000000001',  'apr_case_approved_00000000000000000000001', 'approve', 1,    'emp_00000000000000000000000000000001', 'Approved by manager',               datetime('now', '-9 days'), datetime('now', '-9 days')),
    ('apa_case_submit_rejected_0000000000000001',  'apr_case_rejected_00000000000000000000001', 'submit',  NULL, 'emp_00000000000000000000000000000004', 'Submitted rejected flow request',   datetime('now', '-8 days'), datetime('now', '-8 days')),
    ('apa_case_reject_0000000000000000000000001',  'apr_case_rejected_00000000000000000000001', 'reject',  1,    'emp_00000000000000000000000000000001', 'Rejected due to policy',            datetime('now', '-7 days'), datetime('now', '-7 days')),
    ('apa_case_submit_cancel_000000000000000001',  'apr_case_cancelled_0000000000000000000001', 'submit',  NULL, 'emp_00000000000000000000000000000005', 'Submitted then cancelled',          datetime('now', '-6 days'), datetime('now', '-6 days')),
    ('apa_case_cancel_0000000000000000000000001',  'apr_case_cancelled_0000000000000000000001', 'cancel',  NULL, 'emp_00000000000000000000000000000005', 'Requester cancelled before review', datetime('now', '-5 days'), datetime('now', '-5 days')),
    ('apa_case_submit_step2_0000000000000000001',  'apr_case_pending_step2_000000000000000001', 'submit',  NULL, 'emp_00000000000000000000000000000006', 'Submitted multi-step request',      datetime('now', '-1 days'), datetime('now', '-1 days')),
    ('apa_case_approve_step1_000000000000000001',  'apr_case_pending_step2_000000000000000001', 'approve', 1,    'emp_00000000000000000000000000000001', 'Step 1 approved, waiting step 2',   datetime('now', '-20 hours'), datetime('now', '-20 hours'));

-- ---------------------------------------------------------------------------
-- Leave record for approved request case
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
    'lrec_case_from_approved_request_000000001',
    e.id,
    lt.id,
    '2026-05-12',
    '2026-05-13',
    2,
    datetime('now', '-9 days'),
    rb.id,
    'Generated from approved request case',
    'lrq_case_approved_0000000000000000000001',
    datetime('now', '-9 days'),
    datetime('now', '-9 days')
FROM employees e
JOIN leave_types lt ON lt.uid = 'ltype_00000000000000000000000000000002'
JOIN employees rb ON rb.uid = 'emp_00000000000000000000000000000001'
WHERE e.uid = 'emp_00000000000000000000000000000003';

-- ---------------------------------------------------------------------------
-- Leave balance transaction variants (all domain transaction types)
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO leave_balance_transactions (
    uid,
    balance_id,
    transaction_type,
    days,
    leave_record_id,
    notes,
    created_by,
    created_at
)
SELECT
    'lbt_case_initial_00000000000000000000001',
    lb.id,
    'INITIAL',
    21,
    NULL,
    'Initial yearly balance case',
    e.id,
    datetime('now', '-30 days')
FROM leave_balances lb
JOIN employees e ON e.id = lb.employee_id
JOIN leave_types lt ON lt.id = lb.leave_type_id
WHERE e.uid = 'emp_00000000000000000000000000000001'
  AND lt.code = 'ANNUAL'
  AND lb.year = 2026;

INSERT OR IGNORE INTO leave_balance_transactions (
    uid,
    balance_id,
    transaction_type,
    days,
    leave_record_id,
    notes,
    created_by,
    created_at
)
SELECT
    'lbt_case_deduct_000000000000000000000001',
    lb.id,
    'DEDUCT',
    2,
    lr.id,
    'Deduct after approved leave request case',
    e.id,
    datetime('now', '-9 days')
FROM leave_balances lb
JOIN employees e ON e.id = lb.employee_id
JOIN leave_types lt ON lt.id = lb.leave_type_id
LEFT JOIN leave_records lr ON lr.uid = 'lrec_case_from_approved_request_000000001'
WHERE e.uid = 'emp_00000000000000000000000000000001'
  AND lt.code = 'ANNUAL'
  AND lb.year = 2026;

INSERT OR IGNORE INTO leave_balance_transactions (
    uid,
    balance_id,
    transaction_type,
    days,
    leave_record_id,
    notes,
    created_by,
    created_at
)
SELECT
    'lbt_case_refund_000000000000000000000001',
    lb.id,
    'REFUND',
    1,
    NULL,
    'Refund after manual correction case',
    e.id,
    datetime('now', '-8 days')
FROM leave_balances lb
JOIN employees e ON e.id = lb.employee_id
JOIN leave_types lt ON lt.id = lb.leave_type_id
WHERE e.uid = 'emp_00000000000000000000000000000001'
  AND lt.code = 'ANNUAL'
  AND lb.year = 2026;

INSERT OR IGNORE INTO leave_balance_transactions (
    uid,
    balance_id,
    transaction_type,
    days,
    leave_record_id,
    notes,
    created_by,
    created_at
)
SELECT
    'lbt_case_adjustment_000000000000000000001',
    lb.id,
    'ADJUSTMENT',
    1,
    NULL,
    'Administrative adjustment case',
    e.id,
    datetime('now', '-7 days')
FROM leave_balances lb
JOIN employees e ON e.id = lb.employee_id
JOIN leave_types lt ON lt.id = lb.leave_type_id
WHERE e.uid = 'emp_00000000000000000000000000000001'
  AND lt.code = 'ANNUAL'
  AND lb.year = 2026;

INSERT OR IGNORE INTO leave_balance_transactions (
    uid,
    balance_id,
    transaction_type,
    days,
    leave_record_id,
    notes,
    created_by,
    created_at
)
SELECT
    'lbt_case_carry_over_00000000000000000001',
    lb.id,
    'CARRY_OVER',
    5,
    NULL,
    'Year carry-over case',
    e.id,
    datetime('now', '-6 days')
FROM leave_balances lb
JOIN employees e ON e.id = lb.employee_id
JOIN leave_types lt ON lt.id = lb.leave_type_id
WHERE e.uid = 'emp_00000000000000000000000000000001'
  AND lt.code = 'ANNUAL'
  AND lb.year = 2026;
