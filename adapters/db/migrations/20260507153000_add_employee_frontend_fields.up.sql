ALTER TABLE employees ADD COLUMN actual_appointment_reappointment_date TEXT;
ALTER TABLE employees ADD COLUMN appointment_decision_date TEXT;
ALTER TABLE employees ADD COLUMN appointment_decision_number TEXT;
ALTER TABLE employees ADD COLUMN appointment_type TEXT;
ALTER TABLE employees ADD COLUMN department_name TEXT;
ALTER TABLE employees ADD COLUMN nature_of_appointment TEXT;
ALTER TABLE employees ADD COLUMN reappointment INTEGER NOT NULL DEFAULT 0;
