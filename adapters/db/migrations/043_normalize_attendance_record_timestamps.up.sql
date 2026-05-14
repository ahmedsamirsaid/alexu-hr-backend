UPDATE attendance_records
SET punched_at = (
    substring(punched_at FROM 1 FOR 10) || 'T' ||
    substring(punched_at FROM 12 FOR 8) ||
    substring(punched_at FROM 21 FOR 3) || ':' || substring(punched_at FROM 24 FOR 2)
)
WHERE position('T' IN punched_at) = 0
  AND length(punched_at) >= 25
  AND substring(punched_at FROM 11 FOR 1) = ' '
  AND substring(punched_at FROM 20 FOR 1) = ' '
  AND substring(punched_at FROM 21 FOR 1) IN ('+', '-')
  AND substring(punched_at FROM 26 FOR 1) = ' ';
