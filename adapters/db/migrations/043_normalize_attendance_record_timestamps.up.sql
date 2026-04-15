UPDATE attendance_records
SET punched_at = (
    substr(punched_at, 1, 10) || 'T' ||
    substr(punched_at, 12, 8) ||
    substr(punched_at, 21, 3) || ':' || substr(punched_at, 24, 2)
)
WHERE instr(punched_at, 'T') = 0
  AND length(punched_at) >= 25
  AND substr(punched_at, 11, 1) = ' '
  AND substr(punched_at, 20, 1) = ' '
  AND substr(punched_at, 21, 1) IN ('+', '-')
  AND substr(punched_at, 26, 1) = ' ';
