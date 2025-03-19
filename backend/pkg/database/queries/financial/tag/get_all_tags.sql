SELECT
    id,
    "name",
    color,
    user_id
FROM
    financial_tag
WHERE
    1 = 1
    AND user_id = $3
    AND "name" LIKE $4
LIMIT
    $1 OFFSET $2;