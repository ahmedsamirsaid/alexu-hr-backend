CREATE TABLE IF NOT EXISTS shifts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    grace_minutes INTEGER NOT NULL,
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT OR IGNORE INTO shifts (id, uid, start_time, end_time, grace_minutes)
SELECT 1, 'shf_general', work_day_start, work_day_end, late_grace_minutes
FROM work_hours_configs
WHERE id = 1;

INSERT OR IGNORE INTO shifts (id, uid, start_time, end_time, grace_minutes)
VALUES (2, 'shf_general_seed', '09:00', '17:00', 15);

ALTER TABLE departments ADD COLUMN default_shift_uid TEXT REFERENCES shifts(uid);
ALTER TABLE employees ADD COLUMN shift_uid TEXT REFERENCES shifts(uid);

UPDATE departments
SET default_shift_uid = COALESCE(default_shift_uid, (SELECT uid FROM shifts ORDER BY id ASC LIMIT 1));

UPDATE employees
SET shift_uid = COALESCE(shift_uid, (SELECT uid FROM shifts ORDER BY id ASC LIMIT 1));
