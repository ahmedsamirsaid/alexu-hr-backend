ALTER TABLE employees ADD COLUMN type TEXT NOT NULL DEFAULT 'permanent'
    CHECK (type IN ('permanent', 'temporary'));

ALTER TABLE employees ADD COLUMN sub_type TEXT NOT NULL DEFAULT 'normal'
    CHECK (
        (type = 'permanent' AND sub_type IN ('normal', 'special_needs')) OR
        (type = 'temporary' AND sub_type IN (
            'separation_termination_for_budget',
            'comprehensive_bonus',
            'contract_employees'
        ))
    );
