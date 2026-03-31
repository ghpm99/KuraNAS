CREATE TABLE IF NOT EXISTS libraries (
    id          SERIAL PRIMARY KEY,
    category    VARCHAR(20) NOT NULL UNIQUE,
    path        VARCHAR(500) NOT NULL,
    created_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT libraries_category_check CHECK (
        category IN ('images', 'music', 'videos', 'documents')
    )
);
