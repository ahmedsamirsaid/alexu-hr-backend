-- Comprehensive departmental seed set:
-- - 5 employees per department (1 manager + 4 staff)
-- - user accounts for all seeded employees
-- - department-scoped manager assignments
-- - attendance records for all seeded users with late/missed/absence cases
-- - leave requests with approved/rejected/pending mapped to each department manager

-- ---------------------------------------------------------------------------
-- 1) Seed 5 employees per department (1 manager + 4 staff)
-- ---------------------------------------------------------------------------
WITH RECURSIVE seq(n) AS (
    VALUES(1)
    UNION ALL
    SELECT n + 1 FROM seq WHERE n < 5
),
departments_seed(dept_uid, dept_code, dept_idx) AS (
    VALUES
        ('dept_hr', 'HR', 1),
        ('dept_it', 'IT', 2),
        ('dept_fin', 'FIN', 3),
        ('dept_ops', 'OPS', 4),
        ('dept_sec', 'SEC', 5),
        ('dept_admin', 'ADM', 6),
        ('dept_acad', 'ACA', 7),
        ('dept_lib', 'LIB', 8)
)
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
    shift_uid,
    created_at,
    updated_at
)
SELECT
    'emp56_' || ds.dept_code || '_' || printf('%02d', seq.n) AS uid,
    CASE
        WHEN seq.n = 1 THEN ds.dept_code || ' Department Manager'
        ELSE ds.dept_code || ' Staff ' || printf('%02d', seq.n)
    END AS name,
    '+201188' || printf('%04d', (ds.dept_idx * 10) + seq.n) || '00' AS mobile,
    '299' || printf('%011d', (ds.dept_idx * 10) + seq.n) AS government_id,
    'U56' || ds.dept_code || printf('%02d', seq.n) AS university_id,
    lower(ds.dept_code) || '.seed' || printf('%02d', seq.n) || '@university.edu.eg' AS email,
    date('2022-01-01', '+' || ((ds.dept_idx * 5) + seq.n) || ' days') AS hire_date,
    'active' AS status,
    ds.dept_uid,
    COALESCE((SELECT d.default_shift_uid FROM departments d WHERE d.uid = ds.dept_uid), 'shf_general_seed') AS shift_uid,
    datetime('now'),
    datetime('now')
FROM departments_seed ds
CROSS JOIN seq;

-- ---------------------------------------------------------------------------
-- 2) Seed user accounts for all new employees
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO users (uid, phone, employee_uid, is_active, created_at, updated_at)
SELECT
    'usr56_' || e.uid,
    e.mobile,
    e.uid,
    1,
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.status = 'active'
    AND e.department_uid IN ('dept_hr', 'dept_it', 'dept_fin', 'dept_ops', 'dept_sec', 'dept_admin', 'dept_acad', 'dept_lib');

-- ---------------------------------------------------------------------------
-- 3) Role assignments
--    - Everyone gets Employee role
--    - employee ##01 in each department gets Department Manager (scoped)
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT
    u.id,
    r.id,
    NULL,
    datetime('now')
FROM users u
JOIN roles r ON r.uid = 'role_employee'
WHERE u.uid LIKE 'usr56_emp56_%';

INSERT OR IGNORE INTO user_roles (user_id, role_id, department_uid, created_at)
SELECT
    u.id,
    r.id,
    e.department_uid,
    datetime('now')
FROM users u
JOIN employees e ON e.uid = u.employee_uid
JOIN roles r ON r.uid = 'role_department_manager'
WHERE u.uid LIKE 'usr56_emp56_%_01';

-- ---------------------------------------------------------------------------
-- 4) Attendance devices by department (one per department)
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO attendance_devices (uid, ip, port, name, location, serial_number, status)
VALUES
    ('adev56_hr', '192.168.2.11', 4370, 'HR Device', 'HR Floor', 'ZK-56-HR-001', 'online'),
    ('adev56_it', '192.168.2.12', 4370, 'IT Device', 'IT Floor', 'ZK-56-IT-001', 'online'),
    ('adev56_fin', '192.168.2.13', 4370, 'FIN Device', 'Finance Floor', 'ZK-56-FIN-001', 'online'),
    ('adev56_ops', '192.168.2.14', 4370, 'OPS Device', 'Operations Floor', 'ZK-56-OPS-001', 'offline'),
    ('adev56_sec', '192.168.2.15', 4370, 'SEC Device', 'Security Gate', 'ZK-56-SEC-001', 'online'),
    ('adev56_adm', '192.168.2.16', 4370, 'ADM Device', 'Admin Floor', 'ZK-56-ADM-001', 'offline'),
    ('adev56_aca', '192.168.2.17', 4370, 'ACA Device', 'Academic Floor', 'ZK-56-ACA-001', 'online'),
    ('adev56_lib', '192.168.2.18', 4370, 'LIB Device', 'Library Gate', 'ZK-56-LIB-001', 'online');

-- ---------------------------------------------------------------------------
-- 5) Attendance records
--    Day A: every user has check-in/check-out (baseline)
--    Day B: all attendance edge cases are represented
-- ---------------------------------------------------------------------------
-- Day A baseline for everyone
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
    'atr56_a_in_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-21T08:50:00Z',
    'check_in',
    '{"source":"seed56","case":"baseline"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.status = 'active'
    AND e.department_uid IN ('dept_hr', 'dept_it', 'dept_fin', 'dept_ops', 'dept_sec', 'dept_admin', 'dept_acad', 'dept_lib');

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
    'atr56_a_out_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-21T17:00:00Z',
    'check_out',
    '{"source":"seed56","case":"baseline"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%';

-- Day B case set
-- Managers (##01): normal with break punches
INSERT OR IGNORE INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_b_mgr_in_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-22T08:45:00Z',
    'check_in',
    '{"source":"seed56","case":"manager_normal"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_01';

INSERT OR IGNORE INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_b_mgr_break_s_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-22T12:30:00Z',
    'break_start',
    '{"source":"seed56","case":"manager_break"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_01';

INSERT OR IGNORE INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_b_mgr_break_e_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-22T13:00:00Z',
    'break_end',
    '{"source":"seed56","case":"manager_break"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_01';

INSERT OR IGNORE INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_b_mgr_out_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-22T17:05:00Z',
    'check_out',
    '{"source":"seed56","case":"manager_normal"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_01';

-- Staff ##02: late arrival (late case)
INSERT OR IGNORE INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_b_late_in_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-22T09:25:00Z',
    'check_in',
    '{"source":"seed56","case":"late_arrival"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_02';

INSERT OR IGNORE INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_b_late_out_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-22T17:02:00Z',
    'check_out',
    '{"source":"seed56","case":"late_arrival"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_02';

-- Staff ##03: missed punch out
INSERT OR IGNORE INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_b_miss_out_in_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-22T08:58:00Z',
    'check_in',
    '{"source":"seed56","case":"missed_punch_out"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_03';

-- Staff ##04: missed punch in (checkout only)
INSERT OR IGNORE INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_b_miss_in_out_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-22T16:55:00Z',
    'check_out',
    '{"source":"seed56","case":"missed_punch_in"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_04';

-- Staff ##05: absent (no day-B punches)
-- Also add one unknown punch sample for diagnostics
INSERT OR IGNORE INTO attendance_records (uid, employee_uid, device_uid, device_user_id, punched_at, punch_type, raw_payload, created_at, updated_at)
SELECT
    'atr56_b_unknown_' || e.uid,
    e.uid,
    CASE
        WHEN e.department_uid = 'dept_admin' THEN 'adev56_adm'
        WHEN e.department_uid = 'dept_acad' THEN 'adev56_aca'
        ELSE 'adev56_' || lower(substr(e.department_uid, 6))
    END,
    e.university_id,
    '2026-04-22T10:00:00Z',
    'unknown',
    '{"source":"seed56","case":"unknown_punch_type"}',
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid = 'emp56_IT_05';

-- Materialize attendance exceptions for all required cases
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
    'aex56_late_' || e.uid,
    e.uid,
    '2026-04-22',
    'late_arrival',
    '2026-04-22T09:25:00Z',
    '2026-04-22T17:02:00Z',
    15,
    10,
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_02';

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
    'aex56_miss_out_' || e.uid,
    e.uid,
    '2026-04-22',
    'missed_punch_out',
    '2026-04-22T08:58:00Z',
    NULL,
    15,
    NULL,
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_03';

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
    'aex56_miss_in_' || e.uid,
    e.uid,
    '2026-04-22',
    'missed_punch_in',
    NULL,
    '2026-04-22T16:55:00Z',
    15,
    NULL,
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_04';

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
    'aex56_absence_' || e.uid,
    e.uid,
    '2026-04-22',
    'absence',
    NULL,
    NULL,
    NULL,
    NULL,
    datetime('now'),
    datetime('now')
FROM employees e
WHERE e.uid LIKE 'emp56_%_05';

-- ---------------------------------------------------------------------------
-- 6) Leave balances for 2026 (to support request workflows)
-- ---------------------------------------------------------------------------
INSERT OR IGNORE INTO leave_balances (
    uid,
    employee_id,
    leave_type_id,
    year,
    total_days,
    used_days,
    created_at,
    updated_at
)
SELECT
    'lbal56_' || e.uid || '_' || lt.code || '_2026',
    e.id,
    lt.id,
    2026,
    lt.default_balance,
    0,
    datetime('now'),
    datetime('now')
FROM employees e
JOIN leave_types lt ON lt.code IN ('CASUAL', 'ANNUAL', 'SICK')
WHERE e.status = 'active'
    AND e.department_uid IN ('dept_hr', 'dept_it', 'dept_fin', 'dept_ops', 'dept_sec', 'dept_admin', 'dept_acad', 'dept_lib');

-- ---------------------------------------------------------------------------
-- 7) Leave cases by department with manager-correct mapping
--    - Reuses existing seeded managers/employees if present
--    - Falls back to emp56_* records when needed
-- ---------------------------------------------------------------------------
WITH dept_codes(dept_uid, dept_code) AS (
    VALUES
        ('dept_hr', 'HR'),
        ('dept_it', 'IT'),
        ('dept_fin', 'FIN'),
        ('dept_ops', 'OPS'),
        ('dept_sec', 'SEC'),
        ('dept_admin', 'ADM'),
        ('dept_acad', 'ACA'),
        ('dept_lib', 'LIB')
), ranked_employees AS (
    SELECT
        e.uid,
        e.id,
        e.department_uid,
        ROW_NUMBER() OVER (
            PARTITION BY e.department_uid
            ORDER BY CASE WHEN e.uid LIKE 'emp56_%' THEN 1 ELSE 0 END, e.hire_date, e.id
        ) AS rn
    FROM employees e
    WHERE e.status = 'active'
      AND e.department_uid IN (SELECT dept_uid FROM dept_codes)
), scenario_map AS (
    SELECT
        dc.dept_uid,
        dc.dept_code,
        COALESCE(
            (
                SELECT em.uid
                FROM user_roles ur
                JOIN roles r ON r.id = ur.role_id
                JOIN users u ON u.id = ur.user_id
                JOIN employees em ON em.uid = u.employee_uid
                WHERE r.uid = 'role_department_manager'
                  AND ur.department_uid = dc.dept_uid
                LIMIT 1
            ),
            (SELECT re1.uid FROM ranked_employees re1 WHERE re1.department_uid = dc.dept_uid AND re1.rn = 1)
        ) AS manager_uid,
        (SELECT re2.uid FROM ranked_employees re2 WHERE re2.department_uid = dc.dept_uid AND re2.rn = 2) AS approved_uid,
        (SELECT re3.uid FROM ranked_employees re3 WHERE re3.department_uid = dc.dept_uid AND re3.rn = 3) AS rejected_uid,
        (SELECT re4.uid FROM ranked_employees re4 WHERE re4.department_uid = dc.dept_uid AND re4.rn = 4) AS pending_uid
    FROM dept_codes dc
)
INSERT OR IGNORE INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
SELECT
    'apr56_' || sm.dept_code || '_approved',
    'apf_leave_default',
    sm.approved_uid,
    1,
    1,
    'approved',
    datetime('now', '-5 days'),
    datetime('now', '-4 days')
FROM scenario_map sm
WHERE sm.approved_uid IS NOT NULL;

WITH dept_codes(dept_uid, dept_code) AS (
    VALUES
        ('dept_hr', 'HR'),
        ('dept_it', 'IT'),
        ('dept_fin', 'FIN'),
        ('dept_ops', 'OPS'),
        ('dept_sec', 'SEC'),
        ('dept_admin', 'ADM'),
        ('dept_acad', 'ACA'),
        ('dept_lib', 'LIB')
), ranked_employees AS (
    SELECT
        e.uid,
        e.id,
        e.department_uid,
        ROW_NUMBER() OVER (
            PARTITION BY e.department_uid
            ORDER BY CASE WHEN e.uid LIKE 'emp56_%' THEN 1 ELSE 0 END, e.hire_date, e.id
        ) AS rn
    FROM employees e
    WHERE e.status = 'active'
      AND e.department_uid IN (SELECT dept_uid FROM dept_codes)
), scenario_map AS (
    SELECT
        dc.dept_uid,
        dc.dept_code,
        (SELECT re3.uid FROM ranked_employees re3 WHERE re3.department_uid = dc.dept_uid AND re3.rn = 3) AS rejected_uid
    FROM dept_codes dc
)
INSERT OR IGNORE INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
SELECT
    'apr56_' || sm.dept_code || '_rejected',
    'apf_leave_default',
    sm.rejected_uid,
    1,
    1,
    'rejected',
    datetime('now', '-4 days'),
    datetime('now', '-3 days')
FROM scenario_map sm
WHERE sm.rejected_uid IS NOT NULL;

WITH dept_codes(dept_uid, dept_code) AS (
    VALUES
        ('dept_hr', 'HR'),
        ('dept_it', 'IT'),
        ('dept_fin', 'FIN'),
        ('dept_ops', 'OPS'),
        ('dept_sec', 'SEC'),
        ('dept_admin', 'ADM'),
        ('dept_acad', 'ACA'),
        ('dept_lib', 'LIB')
), ranked_employees AS (
    SELECT
        e.uid,
        e.id,
        e.department_uid,
        ROW_NUMBER() OVER (
            PARTITION BY e.department_uid
            ORDER BY CASE WHEN e.uid LIKE 'emp56_%' THEN 1 ELSE 0 END, e.hire_date, e.id
        ) AS rn
    FROM employees e
    WHERE e.status = 'active'
      AND e.department_uid IN (SELECT dept_uid FROM dept_codes)
), scenario_map AS (
    SELECT
        dc.dept_uid,
        dc.dept_code,
        (SELECT re4.uid FROM ranked_employees re4 WHERE re4.department_uid = dc.dept_uid AND re4.rn = 4) AS pending_uid
    FROM dept_codes dc
)
INSERT OR IGNORE INTO approval_requests (uid, approval_flow_uid, requester_uid, current_step, max_step, status, created_at, updated_at)
SELECT
    'apr56_' || sm.dept_code || '_pending',
    'apf_leave_default',
    sm.pending_uid,
    1,
    1,
    'pending',
    datetime('now', '-2 days'),
    datetime('now', '-2 days')
FROM scenario_map sm
WHERE sm.pending_uid IS NOT NULL;

INSERT OR IGNORE INTO leave_requests (uid, employee_uid, leave_type_uid, start_date, end_date, days, notes, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
SELECT
    'lrq56_' || substr(ar.uid, 7, instr(substr(ar.uid, 7), '_') - 1) || '_approved',
    ar.requester_uid,
    'ltype_00000000000000000000000000000002',
    '2026-05-10',
    '2026-05-10',
    1,
    'Approved leave case',
    datetime('now', '-5 days'),
    datetime('now', '-4 days'),
    ar.uid,
    datetime('now', '-5 days'),
    datetime('now', '-4 days')
FROM approval_requests ar
WHERE ar.uid LIKE 'apr56_%_approved';

INSERT OR IGNORE INTO leave_requests (uid, employee_uid, leave_type_uid, start_date, end_date, days, notes, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
SELECT
    'lrq56_' || substr(ar.uid, 7, instr(substr(ar.uid, 7), '_') - 1) || '_rejected',
    ar.requester_uid,
    'ltype_00000000000000000000000000000003',
    '2026-05-12',
    '2026-05-12',
    1,
    'Rejected leave case',
    datetime('now', '-4 days'),
    datetime('now', '-3 days'),
    ar.uid,
    datetime('now', '-4 days'),
    datetime('now', '-3 days')
FROM approval_requests ar
WHERE ar.uid LIKE 'apr56_%_rejected';

INSERT OR IGNORE INTO leave_requests (uid, employee_uid, leave_type_uid, start_date, end_date, days, notes, submitted_at, decided_at, approval_request_uid, created_at, updated_at)
SELECT
    'lrq56_' || substr(ar.uid, 7, instr(substr(ar.uid, 7), '_') - 1) || '_pending',
    ar.requester_uid,
    'ltype_00000000000000000000000000000001',
    '2026-05-14',
    '2026-05-14',
    1,
    'Pending leave case',
    datetime('now', '-2 days'),
    NULL,
    ar.uid,
    datetime('now', '-2 days'),
    datetime('now', '-2 days')
FROM approval_requests ar
WHERE ar.uid LIKE 'apr56_%_pending';

INSERT OR IGNORE INTO approval_actions (uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at)
SELECT
    'apa56_submit_' || ar.uid,
    ar.uid,
    'submit',
    NULL,
    ar.requester_uid,
    'Submitted by employee',
    CASE
        WHEN ar.status = 'approved' THEN datetime('now', '-5 days')
        WHEN ar.status = 'rejected' THEN datetime('now', '-4 days')
        ELSE datetime('now', '-2 days')
    END,
    CASE
        WHEN ar.status = 'approved' THEN datetime('now', '-5 days')
        WHEN ar.status = 'rejected' THEN datetime('now', '-4 days')
        ELSE datetime('now', '-2 days')
    END
FROM approval_requests ar
WHERE ar.uid LIKE 'apr56_%';

INSERT OR IGNORE INTO approval_actions (uid, approval_request_uid, action, step_order, actor_uid, comments, acted_at, created_at)
SELECT
    'apa56_decide_' || ar.uid,
    ar.uid,
    CASE WHEN ar.status = 'approved' THEN 'approve' ELSE 'reject' END,
    1,
    COALESCE(
        (
            SELECT em.uid
            FROM user_roles ur
            JOIN roles r ON r.id = ur.role_id
            JOIN users u ON u.id = ur.user_id
            JOIN employees em ON em.uid = u.employee_uid
            WHERE r.uid = 'role_department_manager'
              AND ur.department_uid = req.department_uid
            LIMIT 1
        ),
        req.uid
    ),
    CASE WHEN ar.status = 'approved' THEN 'Approved by department manager' ELSE 'Rejected by department manager' END,
    CASE WHEN ar.status = 'approved' THEN datetime('now', '-4 days') ELSE datetime('now', '-3 days') END,
    CASE WHEN ar.status = 'approved' THEN datetime('now', '-4 days') ELSE datetime('now', '-3 days') END
FROM approval_requests ar
JOIN employees req ON req.uid = ar.requester_uid
WHERE ar.uid LIKE 'apr56_%'
  AND ar.status IN ('approved', 'rejected');

INSERT OR IGNORE INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes, leave_request_uid, created_at, updated_at)
SELECT
    'lrec56_' || substr(ar.uid, 7, instr(substr(ar.uid, 7), '_') - 1) || '_approved',
    e_req.id,
    lt.id,
    '2026-05-10',
    '2026-05-10',
    1,
    datetime('now', '-4 days'),
    COALESCE(
        (
            SELECT em.id
            FROM user_roles ur
            JOIN roles r ON r.id = ur.role_id
            JOIN users u ON u.id = ur.user_id
            JOIN employees em ON em.uid = u.employee_uid
            WHERE r.uid = 'role_department_manager'
              AND ur.department_uid = e_req.department_uid
            LIMIT 1
        ),
        e_req.id
    ),
    'Approved leave recorded by manager',
    lr.uid,
    datetime('now', '-4 days'),
    datetime('now', '-4 days')
FROM approval_requests ar
JOIN leave_requests lr ON lr.approval_request_uid = ar.uid
JOIN employees e_req ON e_req.uid = ar.requester_uid
JOIN leave_types lt ON lt.uid = lr.leave_type_uid
WHERE ar.uid LIKE 'apr56_%_approved';

INSERT OR IGNORE INTO leave_balance_transactions (uid, balance_id, transaction_type, days, leave_record_id, notes, created_by, created_at)
SELECT
    'lbt56_' || substr(ar.uid, 7, instr(substr(ar.uid, 7), '_') - 1) || '_approved_deduct',
    lb.id,
    'DEDUCT',
    1,
    lrec.id,
    'Deduct due to approved leave request',
    COALESCE(
        (
            SELECT em.id
            FROM user_roles ur
            JOIN roles r ON r.id = ur.role_id
            JOIN users u ON u.id = ur.user_id
            JOIN employees em ON em.uid = u.employee_uid
            WHERE r.uid = 'role_department_manager'
              AND ur.department_uid = e_req.department_uid
            LIMIT 1
        ),
        e_req.id
    ),
    datetime('now', '-4 days')
FROM approval_requests ar
JOIN leave_requests lr ON lr.approval_request_uid = ar.uid
JOIN employees e_req ON e_req.uid = ar.requester_uid
JOIN leave_types lt ON lt.uid = lr.leave_type_uid
JOIN leave_balances lb ON lb.employee_id = e_req.id AND lb.leave_type_id = lt.id AND lb.year = 2026
JOIN leave_records lrec ON lrec.leave_request_uid = lr.uid
WHERE ar.uid LIKE 'apr56_%_approved';
