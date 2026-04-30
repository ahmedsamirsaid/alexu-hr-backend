INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000042',
    lt.uid,
    'Leave to Care for a Son/Daughter with Special Needs',
    'أجازة رعاية إبن/إبنه من ذوي الاحتياجات الخاصة.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة رعاية إبن/إبنه من ذوي الاحتياجات الخاصة.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000043',
    lt.uid,
    'Leave to Care for a Sister',
    'أجازة رعاية أخت.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة رعاية أخت.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000044',
    lt.uid,
    'Leave to Care for Children',
    'أجازة رعاية أبناء.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة رعاية أبناء.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000045',
    lt.uid,
    'Leave to Care for a Grandson/Granddaughter',
    'أجازة رعاية حفيد/حفيدة.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة رعاية حفيد/حفيدة.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000046',
    lt.uid,
    'Family Reunification Leave',
    'أجازة لجمع شمل الأسرة.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة لجمع شمل الأسرة.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000047',
    lt.uid,
    'Leave to Care for a Father/Mother',
    'أجازة رعاية والد/ والدة.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة رعاية والد/ والدة.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000048',
    lt.uid,
    'Leave to Attend to Family Affairs',
    'أجازة رعاية مصالح الأسرة.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة رعاية مصالح الأسرة.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000049',
    lt.uid,
    'Leave to Care for a Husband',
    'أجازة رعاية زوج.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة رعاية زوج.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000050',
    lt.uid,
    'Leave for Academic Research',
    'أجازة لجمع مادة علمية.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة لجمع مادة علمية.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000051',
    lt.uid,
    'Leave to Visit Family Abroad',
    'أجازة لزيارة الأسرة بالخارج.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة لزيارة الأسرة بالخارج.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000052',
    lt.uid,
    'Attending a Scientific Conference',
    'حضور مؤتمر علمى.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'حضور مؤتمر علمى.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000053',
    lt.uid,
    'Attending a Workshop',
    'حضور ورشة عمل.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'حضور ورشة عمل.'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000054',
    lt.uid,
    'Study Leave',
    'أجازة دراسية'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أجازة دراسية'
  );

INSERT INTO sub_leave_types (uid, leave_type_uid, name_en, name_ar)
SELECT
    'slt_00000000000000000000000000000055',
    lt.uid,
    'Other',
    'أخرى.'
FROM leave_types lt
WHERE lt.code = 'SPECIAL_UNPAID'
  AND NOT EXISTS (
      SELECT 1
      FROM sub_leave_types slt
      WHERE slt.leave_type_uid = lt.uid
        AND slt.name_ar = 'أخرى.'
  );
