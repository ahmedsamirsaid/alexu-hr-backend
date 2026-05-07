-- Reverse: drop preferred_language column from users table.
-- SQLite does not support DROP COLUMN directly before version 3.35.0,
-- so we recreate the table without the column.
CREATE TABLE users_backup AS SELECT id, uid, phone, password_hash, employee_uid, is_active, created_at, updated_at FROM users;
DROP TABLE users;
ALTER TABLE users_backup RENAME TO users;
