INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000037',
    lt.uid,
    'Leave to Perform Hajj',
    'أجازة لأداء فريضة الحج.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_PAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة لأداء فريضة الحج.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000038',
    lt.uid,
    'Leave for Contact with a Sick Person',
    'أجازة مخالط مريض.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_PAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة مخالط مريض.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000039',
    lt.uid,
    'Leave for Work Injury',
    'أجازة اصابة عمل.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_PAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة اصابة عمل.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000040',
    lt.uid,
    'Leave for Taking Examinations',
    'أجازة إداء الامتحانات.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_PAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة إداء الامتحانات.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000041',
    lt.uid,
    'Other',
    'أخرى.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_PAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أخرى.'
  );
