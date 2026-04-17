-- Create a default admin user (phone: +201000000000, use OTP bypass to login)
-- This user has the admin role assigned
INSERT INTO users (uid, phone, is_active, created_at, updated_at) VALUES
    ('usr_admin_seed_00000000000000000000', '+201000000000', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP),
    ('usr_manager_seed_000000000000000000', '+201000000001', 1, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP);

-- Assign admin role to the seed user
INSERT INTO user_roles (user_id, role_id, created_at)
SELECT
    (SELECT id FROM users WHERE uid = 'usr_admin_seed_00000000000000000000'),
    (SELECT id FROM roles WHERE name = 'Admin'),
    CURRENT_TIMESTAMP;
