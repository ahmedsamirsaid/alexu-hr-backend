CREATE TABLE penalties_removed (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    penalty_id INTEGER NOT NULL UNIQUE,
    penalty_removal_type TEXT NOT NULL,
    penalty_removal_number TEXT NOT NULL,
    penalty_removal_date DATE NOT NULL,
    notes TEXT NOT NULL,
    penalty_withdrawal_decision_file_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    FOREIGN KEY (penalty_id) REFERENCES penalties(id) ON DELETE CASCADE
);

CREATE INDEX idx_penalties_removed_penalty_id ON penalties_removed(penalty_id);
