-- Backfill employee type/sub_type for all currently seeded employees.
-- This covers:
-- - the original 12 sample employees
-- - explicit case employees
-- - the departmental employee set from migration 056
-- The distribution intentionally includes every type/sub_type enum.

-- 1) Assign a baseline classification to every currently seeded employee.
UPDATE employees
SET
    type = 'permanent',
    sub_type = 'normal'
WHERE
    uid LIKE 'emp\_%' ESCAPE '\'
    OR uid LIKE 'emp56\_%' ESCAPE '\';

-- 2) Permanent / special_needs
UPDATE employees
SET
    type = 'permanent',
    sub_type = 'special_needs'
WHERE uid IN (
    'emp_00000000000000000000000000000002',
    'emp_00000000000000000000000000000006',
    'emp_00000000000000000000000000000010',
    'emp56_HR_02',
    'emp56_IT_02',
    'emp56_FIN_02',
    'emp56_OPS_02',
    'emp56_SEC_02',
    'emp56_ADM_02',
    'emp56_ACA_02',
    'emp56_LIB_02'
);

-- 3) Temporary / contract_employees
UPDATE employees
SET
    type = 'temporary',
    sub_type = 'contract_employees'
WHERE uid IN (
    'emp_00000000000000000000000000000003',
    'emp_00000000000000000000000000000008',
    'emp56_HR_03',
    'emp56_IT_03',
    'emp56_FIN_03',
    'emp56_OPS_03',
    'emp56_SEC_03',
    'emp56_ADM_03',
    'emp56_ACA_03',
    'emp56_LIB_03'
);

-- 4) Temporary / comprehensive_bonus
UPDATE employees
SET
    type = 'temporary',
    sub_type = 'comprehensive_bonus'
WHERE uid IN (
    'emp_00000000000000000000000000000004',
    'emp_00000000000000000000000000000009',
    'emp_case_inactive_000000000000000000001',
    'emp56_HR_04',
    'emp56_IT_04',
    'emp56_FIN_04',
    'emp56_OPS_04',
    'emp56_SEC_04',
    'emp56_ADM_04',
    'emp56_ACA_04',
    'emp56_LIB_04'
);

-- 5) Temporary / separation_termination_for_budget
UPDATE employees
SET
    type = 'temporary',
    sub_type = 'separation_termination_for_budget'
WHERE uid IN (
    'emp_00000000000000000000000000000005',
    'emp_00000000000000000000000000000011',
    'emp_case_terminated_0000000000000000001',
    'emp56_HR_05',
    'emp56_IT_05',
    'emp56_FIN_05',
    'emp56_OPS_05',
    'emp56_SEC_05',
    'emp56_ADM_05',
    'emp56_ACA_05',
    'emp56_LIB_05'
);
