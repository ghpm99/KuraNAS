DELETE FROM music_artist_clusters
WHERE artist_key <> ALL ($1);
