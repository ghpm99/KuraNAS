-- Synced e-mail messages (task 15). Read-only by design: only metadata and a
-- sanitized plain-text body are stored — attachments are NEVER downloaded
-- (attachment_meta holds filename/mime/size only) and URLs found in the body are
-- recorded as bare domains in link_domains and NEVER visited. auth_results holds
-- {spf, dkim, dmarc} parsed from the Authentication-Results header. status moves
-- pending -> prefiltered_spam | analyzed (task 16) | failed. prefilter_rules lists
-- which deterministic rules flagged a message as spam (evidence for task 16).
CREATE TABLE IF NOT EXISTS email_message (
    id SERIAL PRIMARY KEY,
    account_id INTEGER NOT NULL REFERENCES email_account (id) ON DELETE CASCADE,
    provider_message_id TEXT NOT NULL,
    sender_name TEXT NOT NULL DEFAULT '',
    sender_address TEXT NOT NULL DEFAULT '',
    subject TEXT NOT NULL DEFAULT '',
    snippet TEXT NOT NULL DEFAULT '',
    sanitized_body TEXT,
    received_at TIMESTAMPTZ NOT NULL,
    auth_results JSONB NOT NULL DEFAULT '{}',
    attachment_meta JSONB NOT NULL DEFAULT '[]',
    link_domains JSONB NOT NULL DEFAULT '[]',
    prefilter_rules JSONB NOT NULL DEFAULT '[]',
    status VARCHAR(32) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'prefiltered_spam', 'analyzed', 'failed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (account_id, provider_message_id)
);

CREATE INDEX IF NOT EXISTS idx_email_message_account_received
    ON email_message (account_id, received_at DESC);
