-- Seed additional leave types
INSERT INTO leave_types (uid, code, name_en, name_ar, default_balance, max_consecutive, recording_deadline_days, advance_notice_days, is_active)
VALUES
    ('ltype_00000000000000000000000000000002', 'ANNUAL', 'Annual Leave', 'سنوية', 21, NULL, 30, 7, 1),
    ('ltype_00000000000000000000000000000003', 'SICK', 'Sick Leave', 'مرضية', 30, NULL, 3, NULL, 1);

-- Seed sample employees
INSERT INTO employees (uid, name, mobile, government_id, university_id, email, hire_date, status)
VALUES
    ('emp_00000000000000000000000000000001', 'أحمد محمد حسن', '+201001234567', '28501151234567', 'UNI001', 'ahmed.hassan@university.edu.eg', '2020-01-15', 'active'),
    ('emp_00000000000000000000000000000002', 'فاطمة علي إبراهيم', '+201012345678', '29002281234568', 'UNI002', 'fatma.ibrahim@university.edu.eg', '2019-06-01', 'active'),
    ('emp_00000000000000000000000000000003', 'محمد محمود سعيد', '+201023456789', '28803151234569', 'UNI003', 'mohamed.said@university.edu.eg', '2021-09-01', 'active'),
    ('emp_00000000000000000000000000000004', 'نادية خالد عمر', '+201034567890', '29105201234570', 'UNI004', 'nadia.omar@university.edu.eg', '2018-03-15', 'active'),
    ('emp_00000000000000000000000000000005', 'يوسف إبراهيم فاروق', '+201045678901', '28707101234571', 'UNI005', 'youssef.farouk@university.edu.eg', '2022-02-01', 'active'),
    ('emp_00000000000000000000000000000006', 'مريم أحمد نور', '+201056789012', '29209251234572', 'UNI006', 'mariam.nour@university.edu.eg', '2020-11-15', 'active'),
    ('emp_00000000000000000000000000000007', 'حسن علي مصطفى', '+201067890123', '28604181234573', 'UNI007', 'hassan.mostafa@university.edu.eg', '2017-08-01', 'active'),
    ('emp_00000000000000000000000000000008', 'سارة محمد عزت', '+201078901234', '29401051234574', 'UNI008', 'sara.ezzat@university.edu.eg', '2023-01-15', 'active'),
    ('emp_00000000000000000000000000000009', 'خالد عبدالرحمن', '+201089012345', '28211221234575', 'UNI009', 'khaled.rahman@university.edu.eg', '2016-04-01', 'active'),
    ('emp_00000000000000000000000000000010', 'ليلى سمير حلمي', '+201090123456', '29308121234576', 'UNI010', 'layla.helmy@university.edu.eg', '2021-07-01', 'active'),
    ('emp_00000000000000000000000000000011', 'عمر طارق زكي', '+201101234567', '28909301234577', 'UNI011', 'omar.zaki@university.edu.eg', '2019-12-01', 'inactive'),
    ('emp_00000000000000000000000000000012', 'هبة فتحي صالح', '+201112345678', '29106151234578', 'UNI012', 'heba.saleh@university.edu.eg', '2020-05-15', 'active');

-- Seed leave balances for 2025 and 2026
-- Employee 1 balances
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_casual_2025', e.id, lt.id, 2025, 7, 3
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_annual_2025', e.id, lt.id, 2025, 21, 5
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_sick_2025', e.id, lt.id, 2025, 30, 2
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_casual_2026', e.id, lt.id, 2026, 7, 1
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_annual_2026', e.id, lt.id, 2026, 21, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp01_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'SICK';

-- Employee 2 balances
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_casual_2025', e.id, lt.id, 2025, 7, 5
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_annual_2025', e.id, lt.id, 2025, 21, 10
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_sick_2025', e.id, lt.id, 2025, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_casual_2026', e.id, lt.id, 2026, 7, 2
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_annual_2026', e.id, lt.id, 2026, 21, 3
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp02_sick_2026', e.id, lt.id, 2026, 30, 1
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'SICK';

-- Employee 3 balances
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_casual_2025', e.id, lt.id, 2025, 7, 7
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_annual_2025', e.id, lt.id, 2025, 21, 15
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_sick_2025', e.id, lt.id, 2025, 30, 5
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_casual_2026', e.id, lt.id, 2026, 7, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_annual_2026', e.id, lt.id, 2026, 21, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp03_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'SICK';

-- Employee 4-12 balances (2025 and 2026)
INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_casual_2025', e.id, lt.id, 2025, 7, 2
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000004' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_annual_2025', e.id, lt.id, 2025, 21, 8
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000004' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_sick_2025', e.id, lt.id, 2025, 30, 3
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000004' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_casual_2026', e.id, lt.id, 2026, 7, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000004' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_annual_2026', e.id, lt.id, 2026, 21, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000004' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp04_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000004' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_casual_2025', e.id, lt.id, 2025, 7, 4
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000005' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_annual_2025', e.id, lt.id, 2025, 21, 12
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000005' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_sick_2025', e.id, lt.id, 2025, 30, 1
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000005' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_casual_2026', e.id, lt.id, 2026, 7, 1
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000005' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_annual_2026', e.id, lt.id, 2026, 21, 2
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000005' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp05_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000005' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_casual_2025', e.id, lt.id, 2025, 7, 1
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000006' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_annual_2025', e.id, lt.id, 2025, 21, 6
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000006' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_sick_2025', e.id, lt.id, 2025, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000006' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_casual_2026', e.id, lt.id, 2026, 7, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000006' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_annual_2026', e.id, lt.id, 2026, 21, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000006' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp06_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000006' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_casual_2025', e.id, lt.id, 2025, 7, 6
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_annual_2025', e.id, lt.id, 2025, 21, 18
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_sick_2025', e.id, lt.id, 2025, 30, 10
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_casual_2026', e.id, lt.id, 2026, 7, 2
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_annual_2026', e.id, lt.id, 2026, 21, 5
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp07_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_casual_2025', e.id, lt.id, 2025, 7, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000008' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_annual_2025', e.id, lt.id, 2025, 21, 3
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000008' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_sick_2025', e.id, lt.id, 2025, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000008' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_casual_2026', e.id, lt.id, 2026, 7, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000008' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_annual_2026', e.id, lt.id, 2026, 21, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000008' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp08_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000008' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_casual_2025', e.id, lt.id, 2025, 7, 5
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_annual_2025', e.id, lt.id, 2025, 21, 21
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_sick_2025', e.id, lt.id, 2025, 30, 7
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_casual_2026', e.id, lt.id, 2026, 7, 1
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_annual_2026', e.id, lt.id, 2026, 21, 3
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp09_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_casual_2025', e.id, lt.id, 2025, 7, 3
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000010' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_annual_2025', e.id, lt.id, 2025, 21, 7
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000010' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_sick_2025', e.id, lt.id, 2025, 30, 2
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000010' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_casual_2026', e.id, lt.id, 2026, 7, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000010' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_annual_2026', e.id, lt.id, 2026, 21, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000010' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp10_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000010' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp11_casual_2025', e.id, lt.id, 2025, 7, 2
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000011' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp11_annual_2025', e.id, lt.id, 2025, 21, 4
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000011' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp11_sick_2025', e.id, lt.id, 2025, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000011' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_casual_2025', e.id, lt.id, 2025, 7, 4
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000012' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_annual_2025', e.id, lt.id, 2025, 21, 9
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000012' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_sick_2025', e.id, lt.id, 2025, 30, 4
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000012' AND lt.code = 'SICK';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_casual_2026', e.id, lt.id, 2026, 7, 1
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000012' AND lt.code = 'CASUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_annual_2026', e.id, lt.id, 2026, 21, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000012' AND lt.code = 'ANNUAL';

INSERT INTO leave_balances (uid, employee_id, leave_type_id, year, total_days, used_days)
SELECT 'lbal_emp12_sick_2026', e.id, lt.id, 2026, 30, 0
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000012' AND lt.code = 'SICK';

-- Seed leave records
-- Employee 1 leave records
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000001', e.id, lt.id, '2025-02-15', '2025-02-16', 2, '2025-02-17 09:00:00', e.id, 'عارضة لظروف عائلية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000002', e.id, lt.id, '2025-03-10', '2025-03-10', 1, '2025-03-11 10:30:00', e.id, 'عارضة'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000003', e.id, lt.id, '2025-04-20', '2025-04-24', 5, '2025-04-10 14:00:00', e.id, 'إجازة سنوية - سفر'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000004', e.id, lt.id, '2025-06-01', '2025-06-02', 2, '2025-06-03 08:30:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'SICK';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000005', e.id, lt.id, '2026-01-05', '2026-01-05', 1, '2026-01-06 09:15:00', e.id, 'عارضة'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'CASUAL';

-- Employee 2 leave records
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000006', e.id, lt.id, '2025-01-20', '2025-01-21', 2, '2025-01-22 10:00:00', e.id, 'عارضة طارئة'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000007', e.id, lt.id, '2025-03-01', '2025-03-02', 2, '2025-03-03 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000008', e.id, lt.id, '2025-05-15', '2025-05-15', 1, '2025-05-16 08:45:00', e.id, 'عارضة'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000009', e.id, lt.id, '2025-07-01', '2025-07-10', 10, '2025-06-15 11:00:00', e.id, 'إجازة سنوية - صيف'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000010', e.id, lt.id, '2026-01-02', '2026-01-03', 2, '2026-01-04 09:30:00', e.id, 'عارضة'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000011', e.id, lt.id, '2026-01-12', '2026-01-14', 3, '2026-01-05 14:00:00', e.id, 'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000012', e.id, lt.id, '2026-01-20', '2026-01-20', 1, '2026-01-21 08:00:00', e.id, 'مرضية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000002' AND lt.code = 'SICK';

-- Employee 3 leave records (used all casual leave)
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000013', e.id, lt.id, '2025-01-10', '2025-01-11', 2, '2025-01-12 10:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000014', e.id, lt.id, '2025-02-05', '2025-02-06', 2, '2025-02-07 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000015', e.id, lt.id, '2025-03-20', '2025-03-21', 2, '2025-03-22 08:30:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000016', e.id, lt.id, '2025-04-15', '2025-04-15', 1, '2025-04-16 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000017', e.id, lt.id, '2025-05-01', '2025-05-15', 15, '2025-04-15 14:00:00', e.id, 'إجازة سنوية طويلة'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000018', e.id, lt.id, '2025-08-10', '2025-08-14', 5, '2025-08-15 09:00:00', e.id, 'مرضية - شهادة طبية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000003' AND lt.code = 'SICK';

-- Employee 7 leave records (senior employee with more history)
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000019', e.id, lt.id, '2025-01-05', '2025-01-06', 2, '2025-01-07 10:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000020', e.id, lt.id, '2025-02-10', '2025-02-11', 2, '2025-02-12 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000021', e.id, lt.id, '2025-04-01', '2025-04-02', 2, '2025-04-03 08:30:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000022', e.id, lt.id, '2025-03-01', '2025-03-07', 7, '2025-02-20 14:00:00', e.id, 'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000023', e.id, lt.id, '2025-06-15', '2025-06-25', 11, '2025-06-01 10:00:00', e.id, 'إجازة سنوية - صيف'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000024', e.id, lt.id, '2025-05-05', '2025-05-14', 10, '2025-05-15 09:00:00', e.id, 'مرضية - عملية جراحية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'SICK';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000025', e.id, lt.id, '2026-01-02', '2026-01-03', 2, '2026-01-04 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000026', e.id, lt.id, '2026-01-19', '2026-01-23', 5, '2026-01-10 11:00:00', e.id, 'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000007' AND lt.code = 'ANNUAL';

-- Employee 9 leave records (exhausted annual leave)
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000027', e.id, lt.id, '2025-02-01', '2025-02-07', 7, '2025-01-20 14:00:00', e.id, 'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000028', e.id, lt.id, '2025-04-10', '2025-04-16', 7, '2025-04-01 10:00:00', e.id, 'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000029', e.id, lt.id, '2025-08-01', '2025-08-07', 7, '2025-07-15 11:00:00', e.id, 'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000030', e.id, lt.id, '2025-03-15', '2025-03-16', 2, '2025-03-17 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000031', e.id, lt.id, '2025-06-01', '2025-06-02', 2, '2025-06-03 08:30:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000032', e.id, lt.id, '2025-09-10', '2025-09-10', 1, '2025-09-11 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000033', e.id, lt.id, '2025-07-20', '2025-07-26', 7, '2025-07-27 10:00:00', e.id, 'مرضية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'SICK';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000034', e.id, lt.id, '2026-01-05', '2026-01-05', 1, '2026-01-06 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000035', e.id, lt.id, '2026-01-12', '2026-01-14', 3, '2026-01-05 14:00:00', e.id, 'إجازة سنوية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000009' AND lt.code = 'ANNUAL';

-- Additional leave records for pagination testing (Employee 1)
INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000036', e.id, lt.id, '2024-01-15', '2024-01-16', 2, '2024-01-17 09:00:00', e.id, 'عارضة 2024'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000037', e.id, lt.id, '2024-03-10', '2024-03-14', 5, '2024-03-01 10:00:00', e.id, 'إجازة سنوية 2024'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000038', e.id, lt.id, '2024-05-20', '2024-05-21', 2, '2024-05-22 08:30:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000039', e.id, lt.id, '2024-07-01', '2024-07-10', 10, '2024-06-15 11:00:00', e.id, 'إجازة صيفية 2024'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'ANNUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000040', e.id, lt.id, '2024-09-05', '2024-09-05', 1, '2024-09-06 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000041', e.id, lt.id, '2024-10-15', '2024-10-17', 3, '2024-10-18 10:00:00', e.id, 'مرضية'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'SICK';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000042', e.id, lt.id, '2024-11-20', '2024-11-21', 2, '2024-11-22 09:00:00', e.id, NULL
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'CASUAL';

INSERT INTO leave_records (uid, employee_id, leave_type_id, start_date, end_date, days, recorded_at, recorded_by, notes)
SELECT 'lrec_00000000000000000000000000000043', e.id, lt.id, '2024-12-22', '2024-12-26', 5, '2024-12-10 14:00:00', e.id, 'إجازة نهاية السنة'
FROM employees e, leave_types lt WHERE e.uid = 'emp_00000000000000000000000000000001' AND lt.code = 'ANNUAL';
