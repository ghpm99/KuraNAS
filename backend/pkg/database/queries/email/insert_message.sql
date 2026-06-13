-- Idempotent insert of one synced message. ON CONFLICT keeps the sync safe to
-- re-run: a message already stored (same account + provider id) returns no row.
INSERT INTO email_message (
    account_id, provider_message_id, sender_name, sender_address, subject,
    snippet, sanitized_body, received_at, auth_results, attachment_meta,
    link_domains, prefilter_rules, status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (account_id, provider_message_id) DO NOTHING
RETURNING id;
