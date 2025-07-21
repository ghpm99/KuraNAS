CREATE TABLE image_metadados (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    file_path TEXT NOT NULL,
    format TEXT,
    mode TEXT,
    width INTEGER,
    height INTEGER,
    info TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);