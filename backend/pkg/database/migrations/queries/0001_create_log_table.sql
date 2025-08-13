CREATE TABLE IF NOT EXISTS
    LOG(
        id SERIAL PRIMARY KEY,
        NAME VARCHAR(256) NOT NULL,
        description VARCHAR(256),
        LEVEL VARCHAR(50) NOT NULL CHECK (LEVEL IN ('DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL')),
        ip_address VARCHAR(45),
        start_time TIMESTAMPTZ NOT NULL,
        end_time TIMESTAMPTZ,
        created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
        deleted_at TIMESTAMPTZ,
        status VARCHAR(50) NOT NULL CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED')),
        extra_data JSON
    );