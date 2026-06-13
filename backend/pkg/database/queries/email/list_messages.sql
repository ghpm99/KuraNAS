-- Lean listing for clients (kiosk on a 2012 tablet): metadata only, NO body.
-- The LEFT JOIN adds the AI verdict/importance/summary when the message has
-- been analyzed (task 16); they are NULL for messages still pending.
SELECT m.id, m.account_id, m.sender_name, m.sender_address, m.subject, m.snippet,
       m.received_at, m.status, m.created_at,
       COALESCE(a.verdict, ''), COALESCE(a.importance, ''), COALESCE(a.summary, '')
FROM email_message m
LEFT JOIN email_analysis a ON a.message_id = m.id
ORDER BY m.received_at DESC
LIMIT $1 OFFSET $2;
