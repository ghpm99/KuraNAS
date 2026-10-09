WITH album_tracks AS (
    SELECT
        am.id,
        am.catalog_album_base_key AS base_key,
        am.catalog_album_label AS album_label,
        am.catalog_artist_key AS artist_key,
        TRIM(COALESCE(am.album_artist, '')) AS album_artist,
        NULLIF(substring(TRIM(COALESCE(am.year, '')) FROM '^\d{4}'), '') AS release_year,
        regexp_replace(COALESCE(hf.parent_path, ''), '[/\\](cd|disc|disco|disk)\s*[0-9]+$', '', 'i') AS album_folder
    FROM
        audio_metadata am
        JOIN home_file hf ON hf.id = am.file_id
    WHERE
        hf.deleted_at IS NULL
        AND am.catalog_album_base_key <> ''
),
compilation_albums AS (
    SELECT
        album_folder,
        LOWER(album_label) AS album_name
    FROM
        album_tracks
    WHERE
        album_artist = ''
    GROUP BY
        album_folder,
        LOWER(album_label)
    HAVING
        COUNT(DISTINCT artist_key) > 1
),
classified_tracks AS (
    SELECT
        track.id,
        track.base_key,
        track.album_label,
        track.album_folder,
        track.release_year,
        (
            track.album_artist = ''
            AND compilation.album_folder IS NOT NULL
        ) AS is_compilation
    FROM
        album_tracks track
        LEFT JOIN compilation_albums compilation ON compilation.album_folder = track.album_folder
        AND compilation.album_name = LOWER(track.album_label)
),
homonymous_albums AS (
    SELECT
        base_key
    FROM
        classified_tracks
    WHERE
        NOT is_compilation
    GROUP BY
        base_key
    HAVING
        COUNT(DISTINCT album_folder) > 1
        AND COUNT(DISTINCT release_year) > 1
),
target_groupings AS (
    SELECT
        track.id,
        CASE
            WHEN track.is_compilation THEN 'various-artists::' || LOWER(track.album_label) || '::' || LEFT(MD5(track.album_folder), 8)
            WHEN homonymous.base_key IS NOT NULL THEN track.base_key || '::' || LEFT(MD5(track.album_folder), 8)
            ELSE track.base_key
        END AS album_key,
        CASE
            WHEN track.is_compilation THEN $1::TEXT
        END AS album_artist_label
    FROM
        classified_tracks track
        LEFT JOIN homonymous_albums homonymous ON homonymous.base_key = track.base_key
)
UPDATE audio_metadata am
SET
    catalog_album_key = target.album_key,
    catalog_album_artist_label = target.album_artist_label
FROM
    target_groupings target
WHERE
    am.id = target.id
    AND (
        am.catalog_album_key IS DISTINCT FROM target.album_key
        OR am.catalog_album_artist_label IS DISTINCT FROM target.album_artist_label
    );
