-- Feeds the deterministic pre-filter step: just the fields the rules read.
SELECT id, sender_address, subject, auth_results, attachment_meta, link_domains
FROM email_message
WHERE status = 'pending'
ORDER BY received_at DESC
LIMIT $1;
