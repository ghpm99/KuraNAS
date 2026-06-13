-- Lean listing for clients (kiosk on a 2012 tablet): metadata only, NO body.
SELECT id, account_id, sender_name, sender_address, subject, snippet,
       received_at, status, created_at
FROM email_message
ORDER BY received_at DESC
LIMIT $1 OFFSET $2;
