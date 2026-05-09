-- Add preferred_language column to users table.
-- Defaults to 'ar' (Arabic) for all existing users per product requirement.
ALTER TABLE users ADD COLUMN preferred_language TEXT NOT NULL DEFAULT 'ar';
