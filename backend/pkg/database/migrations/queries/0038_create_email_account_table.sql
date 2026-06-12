-- E-mail accounts linked via read-only OAuth2 (task 14). The whole token set
-- (access/refresh/expiry JSON) is encrypted with AES-256-GCM before storage;
-- token_ciphertext is nonce||ciphertext and is never exposed by the API.
CREATE TABLE IF NOT EXISTS email_account (
    id SERIAL PRIMARY KEY,
    provider VARCHAR(32) NOT NULL CHECK (provider IN ('google', 'microsoft')),
    address TEXT NOT NULL,
    display_name TEXT NOT NULL DEFAULT '',
    token_ciphertext BYTEA NOT NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'linked' CHECK (status IN ('linked', 'error', 'reauth_required')),
    sync_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    last_sync_at TIMESTAMPTZ,
    last_error TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (provider, address)
);
