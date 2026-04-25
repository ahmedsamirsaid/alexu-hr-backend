DELETE FROM attendance_edit_history
WHERE uid IN (
    'aeh_seed_it_manager_it02_in',
    'aeh_seed_admin_it02_out',
    'aeh_seed_it_manager_aca02_in',
    'aeh_seed_admin_aca04_out',
    'aeh_seed_admin_adm03_in',
    'aeh_seed_it_manager_adm03_out'
);

UPDATE attendance_records
SET punched_at = '2026-04-21T08:50:00Z',
    updated_at = CURRENT_TIMESTAMP
WHERE uid = 'atr56_a_in_emp56_IT_02';

UPDATE attendance_records
SET punched_at = '2026-04-21T17:00:00Z',
    updated_at = CURRENT_TIMESTAMP
WHERE uid = 'atr56_a_out_emp56_IT_02';

UPDATE attendance_records
SET punched_at = '2026-04-21T08:50:00Z',
    updated_at = CURRENT_TIMESTAMP
WHERE uid = 'atr56_a_in_emp56_ACA_02';

UPDATE attendance_records
SET punched_at = '2026-04-21T17:00:00Z',
    updated_at = CURRENT_TIMESTAMP
WHERE uid = 'atr56_a_out_emp56_ACA_04';

UPDATE attendance_records
SET punched_at = '2026-04-21T08:50:00Z',
    updated_at = CURRENT_TIMESTAMP
WHERE uid = 'atr56_a_in_emp56_ADM_03';

UPDATE attendance_records
SET punched_at = '2026-04-21T17:00:00Z',
    updated_at = CURRENT_TIMESTAMP
WHERE uid = 'atr56_a_out_emp56_ADM_03';

DELETE FROM user_roles
WHERE user_id = (SELECT id FROM users WHERE uid = 'usr56_emp56_IT_01')
  AND role_id = (SELECT id FROM roles WHERE uid = 'role_it_manager')
  AND department_uid IS NULL;
