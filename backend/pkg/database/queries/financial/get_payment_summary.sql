SELECT
    fp.payments_date,
    fp.user_id,
    fp.total,
    fp.debit,
    fp.credit,
    fp.dif,
    fp.accumulated
FROM
    financial_paymentsummary fp
WHERE
    1 = 1
    AND fp.user_id = $1
    AND fp.payments_date BETWEEN $2
    AND $3
LIMIT
    $4 OFFSET $5;