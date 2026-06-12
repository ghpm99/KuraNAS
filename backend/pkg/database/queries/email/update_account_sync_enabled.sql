UPDATE email_account
SET sync_enabled = $2,
    updated_at   = now()
WHERE id = $1;
