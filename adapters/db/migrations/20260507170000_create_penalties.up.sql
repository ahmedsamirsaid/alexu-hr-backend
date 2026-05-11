CREATE TABLE penalties (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    employee_uid TEXT NOT NULL,
    penalty_type TEXT NOT NULL,
    penalty_reason TEXT NOT NULL,
    penalty_decision_number TEXT NOT NULL,
    penalty_decision_date DATE NOT NULL,
    penalty_decision_file_url TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (employee_uid) REFERENCES employees(uid) ON DELETE CASCADE
);

CREATE INDEX idx_penalties_employee_uid ON penalties(employee_uid);
