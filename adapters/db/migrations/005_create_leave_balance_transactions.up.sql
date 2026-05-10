CREATE TABLE leave_balance_transactions (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT UNIQUE NOT NULL,
    balance_id INTEGER NOT NULL REFERENCES leave_balances(id) ON DELETE RESTRICT,
    transaction_type TEXT NOT NULL,
    days INTEGER NOT NULL,
    leave_record_id INTEGER REFERENCES leave_records(id),
    notes TEXT,
    created_by INTEGER REFERENCES employees(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_balance_transactions_balance ON leave_balance_transactions(balance_id);
