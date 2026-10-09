SELECT EXISTS (
    SELECT 1
    FROM pg_extension
    WHERE extname = 'pg_trgm'
);
