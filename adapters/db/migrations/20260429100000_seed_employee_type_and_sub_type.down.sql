UPDATE employees
SET
    type = 'permanent',
    sub_type = 'normal'
WHERE
    uid LIKE 'emp\_%' ESCAPE '\'
    OR uid LIKE 'emp56\_%' ESCAPE '\';
