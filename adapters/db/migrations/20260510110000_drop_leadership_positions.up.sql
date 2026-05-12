-- Drop the leadership_positions sidecar table; the roles themselves
-- (role_university_president, role_vice_president, role_dean, role_department_manager)
-- are intentionally kept in the roles/user_roles tables so their names can
-- be displayed on the org chart cards.
DROP TABLE IF EXISTS leadership_positions;
