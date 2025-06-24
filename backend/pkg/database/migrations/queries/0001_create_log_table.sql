CREATE TABLE
    IF NOT EXISTS log (
        id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
        name VARCHAR(256) NOT NULL,
        description VARCHAR(256) NULL,
        level VARCHAR(50) NOT NULL CHECK (level IN ('DEBUG', 'INFO', 'WARNING', 'ERROR', 'CRITICAL')),
        ip_address VARCHAR(45) NULL,
        start_time DATETIME NOT NULL,
        end_time DATETIME NULL,
        created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        deleted_at DATETIME NULL,
        status VARCHAR(50) NOT NULL CHECK (status IN ('PENDING', 'COMPLETED', 'FAILED')),
        extra_data JSON NULL
    );