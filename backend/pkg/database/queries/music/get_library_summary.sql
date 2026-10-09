SELECT
    (
        SELECT
            COUNT(*)
        FROM
            home_file hf
        WHERE
            hf.format = ANY ($1)
            AND hf.deleted_at IS NULL
    ) AS total_tracks,
    (
        SELECT
            COUNT(DISTINCT am.catalog_artist_key)
        FROM
            audio_metadata am
            JOIN home_file hf ON hf.id = am.file_id
        WHERE
            hf.format = ANY ($1)
            AND hf.deleted_at IS NULL
            AND am.catalog_artist_key <> ''
    ) AS total_artists,
    (
        SELECT
            COUNT(DISTINCT am.catalog_album_key)
        FROM
            audio_metadata am
            JOIN home_file hf ON hf.id = am.file_id
        WHERE
            hf.format = ANY ($1)
            AND hf.deleted_at IS NULL
            AND am.catalog_album_key <> ''
    ) AS total_albums,
    (
        SELECT
            COUNT(DISTINCT genre.genre_key)
        FROM
            audio_metadata am
            JOIN home_file hf ON hf.id = am.file_id
            CROSS JOIN LATERAL unnest(am.catalog_genre_keys) AS genre (genre_key)
        WHERE
            hf.format = ANY ($1)
            AND hf.deleted_at IS NULL
    ) AS total_genres,
    (
        SELECT
            COUNT(DISTINCT COALESCE(NULLIF(TRIM(hf.parent_path), ''), '/'))
        FROM
            home_file hf
        WHERE
            hf.format = ANY ($1)
            AND hf.deleted_at IS NULL
    ) AS total_folders;
