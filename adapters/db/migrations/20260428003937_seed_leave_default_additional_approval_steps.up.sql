INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
SELECT
    'afs_leave_step2_dean',
    'apf_leave_default',
    2,
    'role_dean'
WHERE EXISTS (
    SELECT 1
    FROM approval_flows
    WHERE uid = 'apf_leave_default'
)
AND EXISTS (
    SELECT 1
    FROM roles
    WHERE uid = 'role_dean'
)
AND NOT EXISTS (
    SELECT 1
    FROM approval_flow_steps
    WHERE approval_flow_uid = 'apf_leave_default'
      AND step_order = 2
);

INSERT INTO approval_flow_steps (uid, approval_flow_uid, step_order, role_uid)
SELECT
    'afs_leave_step3_university_hr',
    'apf_leave_default',
    3,
    'role_university_human_resources'
WHERE EXISTS (
    SELECT 1
    FROM approval_flows
    WHERE uid = 'apf_leave_default'
)
AND EXISTS (
    SELECT 1
    FROM roles
    WHERE uid = 'role_university_human_resources'
)
AND NOT EXISTS (
    SELECT 1
    FROM approval_flow_steps
    WHERE approval_flow_uid = 'apf_leave_default'
      AND step_order = 3
);
