INSERT INTO
    financial_payment (
        "type",
        "name",
        "date",
        installments,
        payment_date,
        fixed,
        active,
        value,
        status,
        invoice_id,
        user_id
    )
VALUES
    (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11
    )