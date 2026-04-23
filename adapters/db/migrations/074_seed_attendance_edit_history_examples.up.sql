-- Give one seeded IT user the IT Manager role for realistic demo data.
INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT
    u.id,
    r.id,
    NULL,
    CURRENT_TIMESTAMP
FROM users u
JOIN roles r ON r.uid = 'role_it_manager'
WHERE u.uid = 'usr56_emp56_IT_01';

-- Make a few seeded daily punches reflect manual corrections so the UI
-- immediately shows mixed edited and untouched rows.
UPDATE attendance_records
SET punched_at = '2026-04-21T08:55:00Z',
    updated_at = '2026-04-21T09:10:00Z'
WHERE uid = 'atr56_a_in_emp56_IT_02';

UPDATE attendance_records
SET punched_at = '2026-04-21T17:12:00Z',
    updated_at = '2026-04-21T17:20:00Z'
WHERE uid = 'atr56_a_out_emp56_IT_02';

UPDATE attendance_records
SET punched_at = '2026-04-21T09:06:00Z',
    updated_at = '2026-04-21T09:24:00Z'
WHERE uid = 'atr56_a_in_emp56_ACA_02';

UPDATE attendance_records
SET punched_at = '2026-04-21T16:42:00Z',
    updated_at = '2026-04-21T16:50:00Z'
WHERE uid = 'atr56_a_out_emp56_ACA_04';

UPDATE attendance_records
SET punched_at = '2026-04-21T08:58:00Z',
    updated_at = '2026-04-21T09:05:00Z'
WHERE uid = 'atr56_a_in_emp56_ADM_03';

UPDATE attendance_records
SET punched_at = '2026-04-21T17:18:00Z',
    updated_at = '2026-04-21T17:25:00Z'
WHERE uid = 'atr56_a_out_emp56_ADM_03';

INSERT OR IGNORE INTO attendance_edit_history (
    uid,
    attendance_record_uid,
    field_changed,
    old_value,
    new_value,
    reason,
    edited_by_uid,
    created_at
)
VALUES
    (
        'aeh_seed_it_manager_it02_in',
        'atr56_a_in_emp56_IT_02',
        'punched_at',
        '2026-04-21T08:50:00Z',
        '2026-04-21T08:55:00Z',
        'Adjusted after reviewing the IT floor device sync report.',
        'usr56_emp56_IT_01',
        '2026-04-21T09:10:00Z'
    ),
    (
        'aeh_seed_admin_it02_out',
        'atr56_a_out_emp56_IT_02',
        'punched_at',
        '2026-04-21T17:00:00Z',
        '2026-04-21T17:12:00Z',
        'Approved a manual correction after confirming the missed checkout.',
        'usr_admin_seed_00000000000000000000',
        '2026-04-21T17:20:00Z'
    ),
    (
        'aeh_seed_it_manager_aca02_in',
        'atr56_a_in_emp56_ACA_02',
        'punched_at',
        '2026-04-21T08:50:00Z',
        '2026-04-21T09:06:00Z',
        'Adjusted after the academic floor reader synced late.',
        'usr56_emp56_IT_01',
        '2026-04-21T09:24:00Z'
    ),
    (
        'aeh_seed_admin_aca04_out',
        'atr56_a_out_emp56_ACA_04',
        'punched_at',
        '2026-04-21T17:00:00Z',
        '2026-04-21T16:42:00Z',
        'Recorded an approved early departure for an off-site committee meeting.',
        'usr_admin_seed_00000000000000000000',
        '2026-04-21T16:50:00Z'
    ),
    (
        'aeh_seed_admin_adm03_in',
        'atr56_a_in_emp56_ADM_03',
        'punched_at',
        '2026-04-21T08:50:00Z',
        '2026-04-21T08:58:00Z',
        'Corrected the first admin-floor punch after paper sign-in verification.',
        'usr_admin_seed_00000000000000000000',
        '2026-04-21T09:05:00Z'
    ),
    (
        'aeh_seed_it_manager_adm03_out',
        'atr56_a_out_emp56_ADM_03',
        'punched_at',
        '2026-04-21T17:00:00Z',
        '2026-04-21T17:18:00Z',
        'Extended checkout after confirming the network outage on the shared device.',
        'usr56_emp56_IT_01',
        '2026-04-21T17:25:00Z'
    );
