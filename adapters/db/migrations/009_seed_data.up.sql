-- Seed casual leave type
INSERT INTO leave_types (uid, code, name_en, name_ar, default_balance, max_consecutive, recording_deadline_days, advance_notice_days, is_active)
VALUES ('ltype_00000000000000000000000000000001', 'CASUAL', 'Casual Leave', 'عارضة', 7, 2, 2, NULL, 1);

-- Seed weekend days (Friday=5, Saturday=6)
INSERT INTO weekend_config (uid, day_of_week) VALUES ('wknd_00000000000000000000000000000001', 5);
INSERT INTO weekend_config (uid, day_of_week) VALUES ('wknd_00000000000000000000000000000002', 6);
