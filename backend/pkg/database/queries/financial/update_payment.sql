UPDATE
    financial_payment
SET
    "type" = $3,
    "name" = $4,
    "date" = $5,
    installments = $6,
    payment_date = $7,
    fixed = $8,
    active = $9,
    value = $10,
    status = $11,
    invoice_id = $12
WHERE
    1 = 1
    AND id = $1
    AND user_id = $2