CREATE TABLE IF NOT EXISTS shifts (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    grace_minutes INTEGER NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE departments ADD COLUMN default_shift_uid TEXT REFERENCES shifts(uid);
ALTER TABLE employees ADD COLUMN shift_uid TEXT REFERENCES shifts(uid);

