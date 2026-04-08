-- Seed casual leave type
INSERT INTO leave_types (uid, code, name_en, name_ar, default_balance, max_consecutive, recording_deadline_days, advance_notice_days, is_active)
VALUES ('ltype_00000000000000000000000000000001', 'CASUAL', 'Casual Leave', 'عارضة', 7, 2, 2, NULL, 1);

-- Seed weekend days (Friday=5, Saturday=6)
INSERT INTO weekend_config (uid, day_of_week) VALUES ('wknd_00000000000000000000000000000001', 5);
INSERT INTO weekend_config (uid, day_of_week) VALUES ('wknd_00000000000000000000000000000002', 6);

-- Seed common Egyptian holiday definitions
INSERT INTO holiday_definitions (uid, code, name_en, name_ar, default_month, default_day)
VALUES
    ('hdef_00000000000000000000000000000001', 'REVOLUTION_JAN_25', 'January 25 Revolution', 'ثورة 25 يناير', 1, 25),
    ('hdef_00000000000000000000000000000002', 'SINAI_LIBERATION', 'Sinai Liberation Day', 'عيد تحرير سيناء', 4, 25),
    ('hdef_00000000000000000000000000000003', 'LABOUR_DAY', 'Labour Day', 'عيد العمال', 5, 1),
    ('hdef_00000000000000000000000000000004', 'REVOLUTION_JUL_23', 'July 23 Revolution', 'ثورة 23 يوليو', 7, 23),
    ('hdef_00000000000000000000000000000005', 'ARMED_FORCES_DAY', 'Armed Forces Day', 'عيد القوات المسلحة', 10, 6),
    ('hdef_00000000000000000000000000000006', 'EID_AL_FITR', 'Eid Al-Fitr', 'عيد الفطر', NULL, NULL),
    ('hdef_00000000000000000000000000000007', 'EID_AL_ADHA', 'Eid Al-Adha', 'عيد الأضحى', NULL, NULL),
    ('hdef_00000000000000000000000000000008', 'ISLAMIC_NEW_YEAR', 'Islamic New Year', 'رأس السنة الهجرية', NULL, NULL),
    ('hdef_00000000000000000000000000000009', 'MAWLID', 'Prophet Muhammad Birthday', 'المولد النبوي', NULL, NULL),
    ('hdef_00000000000000000000000000000010', 'COPTIC_CHRISTMAS', 'Coptic Christmas', 'عيد الميلاد المجيد', 1, 7),
    ('hdef_00000000000000000000000000000011', 'SHAM_EL_NESSIM', 'Sham El-Nessim', 'شم النسيم', NULL, NULL);
