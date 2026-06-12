SELECT id,
       provider,
       address,
       display_name,
       token_ciphertext,
       status,
       sync_enabled,
       last_sync_at,
       last_error,
       created_at,
       updated_at
FROM email_account
WHERE id = $1;
