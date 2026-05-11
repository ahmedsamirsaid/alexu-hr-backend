CREATE TABLE employees_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    uid TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    mobile TEXT UNIQUE NOT NULL,
    government_id TEXT UNIQUE NOT NULL,
    university_id TEXT UNIQUE NOT NULL,
    email TEXT,
    hire_date TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    type TEXT NOT NULL DEFAULT 'permanent'
        CHECK (type IN ('permanent', 'temporary')),
    sub_type TEXT NOT NULL DEFAULT 'normal'
        CHECK (
            (type = 'permanent' AND sub_type IN ('normal', 'special_needs')) OR
            (type = 'temporary' AND sub_type IN (
                'separation_termination_for_budget',
                'comprehensive_bonus',
                'contract_employees'
            ))
        ),
    telephone_number TEXT,
    person_with_special_needs INTEGER NOT NULL DEFAULT 0,
    date_of_birth TEXT,
    gender TEXT,
    religion TEXT,
    marital_status TEXT,
    address TEXT,
    place_of_birth TEXT,
    place_of_residence TEXT,
    police_station TEXT,
    id_card_valid_until TEXT,
    academic_level TEXT,
    educational_qualification TEXT,
    university_name TEXT,
    faculty TEXT,
    specialization TEXT,
    year_obtained INTEGER,
    grade TEXT,
    work_entity TEXT,
    employee_file_number TEXT,
    insurance_number TEXT,
    employment_status TEXT,
    solidarity_fund INTEGER NOT NULL DEFAULT 0,
    subscription_date TEXT,
    military_status TEXT,
    medical_cadre TEXT,
    personal_photo_url TEXT,
    national_card_image_url TEXT,
    qualification_certificate_image_url TEXT,
    cv_url TEXT,
    member_number TEXT,
    insurance_code TEXT,
    job_group TEXT,
    qualitative_group TEXT,
    job_title_at_level TEXT,
    job_title_before_placement TEXT,
    financial_grade TEXT,
    previous_financial_grade TEXT,
    job_level TEXT,
    decision_date TEXT,
    decision_number INTEGER,
    grade_grant_date TEXT,
    notes TEXT,
    decision_file_url TEXT,
    appointment_seniority_or_grade_withdrawal INTEGER NOT NULL DEFAULT 0,
    department_uid TEXT REFERENCES departments(uid),
    shift_uid TEXT REFERENCES shifts(uid),
    created_at TEXT DEFAULT (datetime('now')),
    updated_at TEXT DEFAULT (datetime('now'))
);

INSERT INTO employees_new (
    id, uid, name, mobile, government_id, university_id, email,
    hire_date, status, type, sub_type,
    telephone_number, person_with_special_needs, date_of_birth, gender, religion, marital_status,
    address, place_of_birth, place_of_residence, police_station, id_card_valid_until,
    academic_level, educational_qualification, university_name, faculty, specialization,
    year_obtained, grade, work_entity, employee_file_number, insurance_number, employment_status,
    solidarity_fund, subscription_date, military_status, medical_cadre,
    personal_photo_url, national_card_image_url, qualification_certificate_image_url, cv_url,
    member_number, insurance_code, job_group, qualitative_group, job_title_at_level,
    job_title_before_placement, financial_grade, previous_financial_grade, job_level,
    decision_date, decision_number, grade_grant_date, notes, decision_file_url,
    appointment_seniority_or_grade_withdrawal, department_uid, shift_uid, created_at, updated_at
)
SELECT
    id, uid, name, mobile, government_id, university_id, email,
    hire_date, status, type, sub_type,
    telephone_number, person_with_special_needs, date_of_birth, gender, religion, marital_status,
    address, place_of_birth, place_of_residence, police_station, id_card_valid_until,
    academic_level, educational_qualification, university_name, faculty, specialization,
    year_obtained, grade, work_entity, employee_file_number, insurance_number, employment_status,
    solidarity_fund, subscription_date, military_status, medical_cadre,
    personal_photo_url, national_card_image_url, qualification_certificate_image_url, cv_url,
    member_number, insurance_code, job_group, qualitative_group, job_title_at_level,
    job_title_before_placement, financial_grade, previous_financial_grade, job_level,
    decision_date, decision_number, grade_grant_date, notes, decision_file_url,
    appointment_seniority_or_grade_withdrawal, department_uid, shift_uid, created_at, updated_at
FROM employees;

DROP TABLE employees;
ALTER TABLE employees_new RENAME TO employees;

CREATE INDEX idx_employees_name ON employees(name);
CREATE INDEX idx_employees_status ON employees(status);
CREATE INDEX idx_employees_department ON employees(department_uid);
