SELECT
    fc.id,
    fc.name,
    SUM(
        CASE
            fp.type
            WHEN 0 THEN fp.value
            ELSE 0
        END
    ) AS total_value_credit,
    SUM(
        CASE
            fp.type
            WHEN 1 THEN fp.value
            ELSE 0
        END
    ) AS total_value_debit,
    SUM(
        CASE
            fp.status
            WHEN 0 THEN fp.value
            ELSE 0
        END
    ) AS total_value_open,
    SUM(
        CASE
            fp.status
            WHEN 1 THEN fp.value
            ELSE 0
        END
    ) AS total_value_closed,
    COUNT(*) AS total_payments
FROM
    financial_contract AS fc
    INNER JOIN financial_invoice fi ON (fc.id = fi.contract_id)
    INNER JOIN financial_payment fp ON (fi.id = fp.invoice_id)
WHERE
    (
        0 = 0
        AND fc.user_id = %(user_id) s
        AND fp.payment_date BETWEEN %(begin) s
        AND %(end
    ) s
    AND fp.active = true
)
GROUP BY
    fc.id,
    fc.name
ORDER BY
    fc.id;