CREATE TABLE incentive_bonus (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    employee_uid TEXT NOT NULL,
    bonus_date DATE NOT NULL,
    decision_number TEXT NOT NULL,
    decision_date DATE NOT NULL,
    decision_image_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (employee_uid) REFERENCES employees(uid) ON DELETE CASCADE
);

CREATE INDEX idx_incentive_bonus_employee_uid ON incentive_bonus(employee_uid);
