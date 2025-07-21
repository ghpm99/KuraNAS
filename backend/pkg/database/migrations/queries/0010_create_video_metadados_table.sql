CREATE TABLE video_metadados (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    format TEXT,
    streams TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);