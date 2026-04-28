-- Restore University Human Resources as a system role if this migration is rolled back.
UPDATE roles
SET is_system = 1,
    updated_at = CURRENT_TIMESTAMP
WHERE uid = 'role_university_human_resources';
