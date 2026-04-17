-- Roll back comprehensive case seed data inserted in 055_seed_all_cases.up.sql.

DELETE FROM leave_balance_transactions
WHERE uid IN (
    'lbt_case_initial_00000000000000000000001',
    'lbt_case_deduct_000000000000000000000001',
    'lbt_case_refund_000000000000000000000001',
    'lbt_case_adjustment_000000000000000000001',
    'lbt_case_carry_over_00000000000000000001'
);

DELETE FROM leave_records
WHERE uid = 'lrec_case_from_approved_request_000000001';

DELETE FROM approval_actions
WHERE uid IN (
    'apa_case_submit_pending_0000000000000001',
    'apa_case_submit_approved_0000000000000001',
    'apa_case_approve_000000000000000000000001',
    'apa_case_submit_rejected_0000000000000001',
    'apa_case_reject_0000000000000000000000001',
    'apa_case_submit_cancel_000000000000000001',
    'apa_case_cancel_0000000000000000000000001',
    'apa_case_submit_step2_0000000000000000001',
    'apa_case_approve_step1_000000000000000001'
);

DELETE FROM leave_requests
WHERE uid IN (
    'lrq_case_pending_0000000000000000000001',
    'lrq_case_approved_0000000000000000000001',
    'lrq_case_rejected_0000000000000000000001',
    'lrq_case_cancelled_0000000000000000000001',
    'lrq_case_pending_step2_000000000000000001'
);

DELETE FROM approval_requests
WHERE uid IN (
    'apr_case_pending_0000000000000000000001',
    'apr_case_approved_00000000000000000000001',
    'apr_case_rejected_00000000000000000000001',
    'apr_case_cancelled_0000000000000000000001',
    'apr_case_pending_step2_000000000000000001'
);

DELETE FROM approval_flow_steps
WHERE uid IN (
    'afs_case_two_step_1',
    'afs_case_two_step_2'
);

DELETE FROM approval_flows
WHERE uid = 'apf_case_two_step';

DELETE FROM attendance_exceptions
WHERE uid IN (
    'aex_case_20260420_missed_in',
    'aex_case_20260420_missed_out',
    'aex_case_20260420_late',
    'aex_case_20260420_early',
    'aex_case_20260420_absence'
);

DELETE FROM attendance_devices
WHERE uid IN (
    'adev_case_online_000000000000000000001',
    'adev_case_offline_00000000000000000001',
    'adev_case_deact_0000000000000000000001'
);

DELETE FROM refresh_tokens
WHERE token_hash IN (
    'seedhash_case_refresh_active_001',
    'seedhash_case_refresh_revoked_001',
    'seedhash_case_refresh_expired_001'
);

DELETE FROM otp_codes
WHERE id IN (900001, 900002, 900003);

DELETE FROM device_tokens
WHERE uid IN (
    'dtok_case_android_00000000000000000001',
    'dtok_case_ios_000000000000000000000001'
);

DELETE FROM user_roles
WHERE user_id IN (
    SELECT id
    FROM users
    WHERE uid IN (
        'usr_case_active_0000000000000000000001',
        'usr_case_inactive_000000000000000000001',
        'usr_case_mgr_it_0000000000000000000001'
    )
);

DELETE FROM users
WHERE uid IN (
    'usr_case_active_0000000000000000000001',
    'usr_case_inactive_000000000000000000001',
    'usr_case_mgr_it_0000000000000000000001'
);

DELETE FROM employees
WHERE uid IN (
    'emp_case_inactive_000000000000000000001',
    'emp_case_terminated_0000000000000000001'
);
