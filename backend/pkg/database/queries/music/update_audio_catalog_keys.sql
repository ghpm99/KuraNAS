UPDATE audio_metadata
SET
    catalog_artist_key = $2,
    catalog_artist_label = $3,
    catalog_album_key = $4,
    catalog_album_label = $5,
    catalog_genre_keys = $6,
    catalog_genre_labels = $7
WHERE
    id = $1;
