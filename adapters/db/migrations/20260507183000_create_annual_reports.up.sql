CREATE TABLE annual_reports (
    id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    employee_uid TEXT NOT NULL,
    report_year TEXT NOT NULL,
    report_grade TEXT NOT NULL,
    notes TEXT NOT NULL,
    report_image_url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (employee_uid) REFERENCES employees(uid) ON DELETE CASCADE
);

CREATE INDEX idx_annual_reports_employee_uid ON annual_reports(employee_uid);
