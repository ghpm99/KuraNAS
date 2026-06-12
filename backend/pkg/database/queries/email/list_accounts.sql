SELECT id,
       provider,
       address,
       display_name,
       status,
       sync_enabled,
       last_sync_at,
       last_error,
       created_at,
       updated_at
FROM email_account
ORDER BY provider, address;
