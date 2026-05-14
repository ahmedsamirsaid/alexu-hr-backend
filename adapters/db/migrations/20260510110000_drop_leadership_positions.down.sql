CREATE TABLE IF NOT EXISTS leadership_positions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL UNIQUE,
    role_uid TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_leadership_positions_user_id ON leadership_positions(user_id);
CREATE INDEX IF NOT EXISTS idx_leadership_positions_role_uid ON leadership_positions(role_uid);
