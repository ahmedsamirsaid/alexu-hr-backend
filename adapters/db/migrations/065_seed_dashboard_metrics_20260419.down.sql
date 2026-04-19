-- Rollback for 065_seed_dashboard_metrics_20260419.up.sql

DELETE FROM approval_actions
WHERE uid IN (
    'apa65_20260419_ops02',
    'apa65_20260419_sec02',
    'apa65_20260419_adm02',
    'apa65_20260419_lib02'
);

DELETE FROM leave_requests
WHERE uid IN (
    'lrq65_20260419_ops02',
    'lrq65_20260419_sec02',
    'lrq65_20260419_adm02',
    'lrq65_20260419_lib02'
);

DELETE FROM approval_requests
WHERE uid IN (
    'apr65_20260419_ops02',
    'apr65_20260419_sec02',
    'apr65_20260419_adm02',
    'apr65_20260419_lib02'
);

DELETE FROM leave_records
WHERE uid IN (
    'lrec65_20260419_hr03',
    'lrec65_20260419_it03',
    'lrec65_20260419_fin03'
);

DELETE FROM attendance_records
WHERE uid IN (
    'atr65_20260419_hr01_in',
    'atr65_20260419_it01_in',
    'atr65_20260419_fin01_in',
    'atr65_20260419_ops01_in',
    'atr65_20260419_sec01_in',
    'atr65_20260419_adm01_in',
    'atr65_20260419_aca01_in',
    'atr65_20260419_lib01_in',
    'atr65_20260419_hr01_out',
    'atr65_20260419_it01_out',
    'atr65_20260419_fin01_out',
    'atr65_20260419_sec01_out',
    'atr65_20260419_aca01_out'
);
