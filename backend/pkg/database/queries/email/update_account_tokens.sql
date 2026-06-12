UPDATE email_account
SET token_ciphertext = $2,
    status           = $3,
    last_error       = $4,
    updated_at       = now()
WHERE id = $1;
