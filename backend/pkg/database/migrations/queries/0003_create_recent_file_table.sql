CREATE TABLE
    IF NOT EXISTS recent_file (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        ip_address VARCHAR(45) NOT NULL,
        file_id INTEGER NOT NULL,
        accessed_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
        UNIQUE (ip_address, file_id)
    );