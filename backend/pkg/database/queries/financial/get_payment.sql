SELECT
    fp.id,
    fp.status,
    fp."type",
    fp."name",
    fp."date",
    fp.installments,
    fp.payment_date,
    fp.fixed,
    fp.active,
    fp.value,
    fp.invoice_id
FROM
    financial_payment fp
WHERE
    1 = 1
    AND fp.id = $1
    AND fp.user_id = $2;