CREATE TABLE audio_metadados (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_id INTEGER NOT NULL,
    "path" TEXT NOT NULL,
    mime TEXT,
    info TEXT,
    tags TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES home_file(id),
    UNIQUE (file_id, "path")
);