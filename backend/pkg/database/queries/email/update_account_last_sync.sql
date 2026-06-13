-- Advance the per-account sync cursor after a successful fetch; a fetch only
-- succeeds with a valid token, so the account is necessarily linked.
UPDATE email_account
SET last_sync_at = $2, status = 'linked', last_error = '', updated_at = now()
WHERE id = $1;
