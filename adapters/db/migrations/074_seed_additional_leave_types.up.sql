INSERT OR IGNORE INTO leave_types (
    uid,
    code,
    name_en,
    name_ar,
    default_balance,
    max_consecutive,
    recording_deadline_days,
    advance_notice_days,
    is_active,
    approval_flow_uid
)
VALUES
    ('ltype_00000000000000000000000000000009', 'SPECIAL_PAID', 'Special Paid Leave', 'إجازة خاصة بمرتب', 0, NULL, NULL, NULL, 1, 'apf_leave_default'),
    ('ltype_00000000000000000000000000000010', 'SPECIAL_UNPAID', 'Special Unpaid Leave', 'إجازة خاصة بدون مرتب', 0, NULL, NULL, NULL, 1, 'apf_leave_default'),
    ('ltype_00000000000000000000000000000011', 'REGULAR', 'Regular Leave', 'إجازة اعتيادى', 21, NULL, 30, 7, 1, 'apf_leave_default'),
    ('ltype_00000000000000000000000000000012', 'MATERNITY', 'Maternity Leave', 'إجازة وضع', 90, NULL, NULL, NULL, 1, 'apf_leave_default');
