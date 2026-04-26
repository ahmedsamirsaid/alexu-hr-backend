-- Make University Human Resources editable in the frontend by marking it as non-system.
UPDATE roles
SET is_system = 0,
    updated_at = CURRENT_TIMESTAMP
WHERE uid = 'role_university_human_resources';
