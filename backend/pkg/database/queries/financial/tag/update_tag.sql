UPDATE
    financial_tag
SET
    "name" = $4,
    color = $3
WHERE
    1 = 1
    AND id = $1
    AND user_id = $2