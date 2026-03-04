INSERT INTO video_playlist_item (
    playlist_id,
    video_id,
    order_index
)
SELECT
    $1,
    UNNEST($2::int[]),
    generate_subscripts($2::int[], 1) - 1;
