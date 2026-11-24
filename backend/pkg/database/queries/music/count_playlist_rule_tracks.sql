WITH specs AS (
    SELECT
        spec.playlist_id,
        spec.match_all,
        COALESCE(spec.genre_keys, '{}') AS genre_keys,
        COALESCE(spec.artist_keys, '{}') AS artist_keys,
        COALESCE(spec.album_keys, '{}') AS album_keys,
        COALESCE(spec.folders, '{}') AS folders,
        COALESCE(spec.folder_patterns, '{}') AS folder_patterns,
        spec.year_from,
        spec.year_to,
        spec.favorites_only
    FROM
        jsonb_to_recordset($2::jsonb) AS spec (
            playlist_id integer,
            match_all boolean,
            genre_keys text[],
            artist_keys text[],
            album_keys text[],
            folders text[],
            folder_patterns text[],
            year_from integer,
            year_to integer,
            favorites_only boolean
        )
)
SELECT
    spec.playlist_id,
    totals.track_count,
    totals.total_length_seconds
FROM
    specs spec
    CROSS JOIN LATERAL (
        SELECT
            COUNT(*)::integer AS track_count,
            COALESCE(SUM(am.length), 0)::float8 AS total_length_seconds
        FROM
            home_file hf
            LEFT JOIN audio_metadata_effective am ON am.file_id = hf.id
            CROSS JOIN LATERAL (
                SELECT
                    substring(am.year FROM '^[0-9]{4}')::integer AS release_year
            ) track_year
            CROSS JOIN LATERAL (
                SELECT
                    cardinality(spec.genre_keys) > 0 AS has_genres,
                    cardinality(spec.artist_keys) > 0 AS has_artists,
                    cardinality(spec.album_keys) > 0 AS has_albums,
                    cardinality(spec.folders) > 0 AS has_folders,
                    (spec.year_from IS NOT NULL OR spec.year_to IS NOT NULL) AS has_years,
                    COALESCE(spec.favorites_only, FALSE) AS has_favorites
            ) criteria
        WHERE
            hf.format = ANY ($1)
            AND hf.deleted_at IS NULL
            AND NOT EXISTS (
                SELECT 1
                FROM playlist_rule_exclusion exclusion
                WHERE exclusion.playlist_id = spec.playlist_id
                    AND exclusion.file_id = hf.id
            )
            AND CASE
                WHEN spec.match_all THEN
                    (NOT criteria.has_genres OR COALESCE(am.catalog_genre_keys && spec.genre_keys, FALSE))
                    AND (NOT criteria.has_artists OR COALESCE(am.catalog_artist_key = ANY (spec.artist_keys), FALSE))
                    AND (NOT criteria.has_albums OR COALESCE(am.catalog_album_key = ANY (spec.album_keys), FALSE))
                    AND (NOT criteria.has_folders OR hf.parent_path = ANY (spec.folders) OR hf.parent_path LIKE ANY (spec.folder_patterns))
                    AND (
                        NOT criteria.has_years
                        OR (
                            track_year.release_year IS NOT NULL
                            AND (spec.year_from IS NULL OR track_year.release_year >= spec.year_from)
                            AND (spec.year_to IS NULL OR track_year.release_year <= spec.year_to)
                        )
                    )
                    AND (NOT criteria.has_favorites OR hf.starred)
                ELSE
                    (criteria.has_genres AND COALESCE(am.catalog_genre_keys && spec.genre_keys, FALSE))
                    OR (criteria.has_artists AND COALESCE(am.catalog_artist_key = ANY (spec.artist_keys), FALSE))
                    OR (criteria.has_albums AND COALESCE(am.catalog_album_key = ANY (spec.album_keys), FALSE))
                    OR (criteria.has_folders AND (hf.parent_path = ANY (spec.folders) OR hf.parent_path LIKE ANY (spec.folder_patterns)))
                    OR (
                        criteria.has_years
                        AND track_year.release_year IS NOT NULL
                        AND (spec.year_from IS NULL OR track_year.release_year >= spec.year_from)
                        AND (spec.year_to IS NULL OR track_year.release_year <= spec.year_to)
                    )
                    OR (criteria.has_favorites AND hf.starred)
            END
    ) totals;
