INSERT INTO email_account (provider, address, display_name, token_ciphertext, status, last_error, updated_at)
VALUES ($1, $2, $3, $4, 'linked', '', now())
ON CONFLICT (provider, address) DO UPDATE
    SET display_name     = EXCLUDED.display_name,
        token_ciphertext = EXCLUDED.token_ciphertext,
        status           = 'linked',
        last_error       = '',
        updated_at       = now()
RETURNING id;
