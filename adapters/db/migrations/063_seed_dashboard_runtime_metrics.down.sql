-- Rollback for 060_seed_dashboard_runtime_metrics.up.sql

DELETE FROM approval_actions
WHERE uid IN (
    'apa60_submit_emp03',
    'apa60_submit_emp04',
    'apa60_submit_emp06',
    'apa60_submit_emp10'
);

DELETE FROM leave_requests
WHERE uid IN (
    'lrq60_pending_emp03',
    'lrq60_pending_emp04',
    'lrq60_pending_emp06',
    'lrq60_pending_emp10'
);

DELETE FROM approval_requests
WHERE uid IN (
    'apr60_pending_emp03',
    'apr60_pending_emp04',
    'apr60_pending_emp06',
    'apr60_pending_emp10'
);

DELETE FROM leave_records
WHERE uid IN (
    'lrec60_today_emp02',
    'lrec60_today_emp05',
    'lrec60_today_emp08'
);

DELETE FROM attendance_records
WHERE uid IN (
    'atr60_today_emp01_in',
    'atr60_today_emp03_in',
    'atr60_today_emp04_in',
    'atr60_today_emp06_in',
    'atr60_today_emp07_in',
    'atr60_today_emp10_in',
    'atr60_today_emp11_unknown',
    'atr60_today_emp12_in',
    'atr60_today_emp01_out',
    'atr60_today_emp03_out',
    'atr60_today_emp07_out'
);
