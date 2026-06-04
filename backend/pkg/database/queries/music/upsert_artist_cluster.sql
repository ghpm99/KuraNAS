INSERT INTO music_artist_clusters (artist_key, artist, cluster_name, updated_at)
VALUES ($1, $2, $3, CURRENT_TIMESTAMP)
ON CONFLICT (artist_key) DO UPDATE
SET artist = EXCLUDED.artist,
    cluster_name = EXCLUDED.cluster_name,
    updated_at = CURRENT_TIMESTAMP;
