-- Feeds the AI analysis step: pending messages (those the deterministic
-- pre-filter let through) with the full sanitized body and every evidence
-- column. Oldest first, so a backlog drains in arrival order.
SELECT id, account_id, sender_name, sender_address, subject, sanitized_body,
       auth_results, attachment_meta, link_domains, prefilter_rules
FROM email_message
WHERE status = 'pending'
ORDER BY received_at ASC
LIMIT $1;
