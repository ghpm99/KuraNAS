INSERT INTO video_playlist_item (
    playlist_id,
    video_id,
    order_index,
    source_kind
)
SELECT
    $1,
    UNNEST($2::int[]),
    generate_subscripts($2::int[], 1) - 1,
    $3
ON CONFLICT (playlist_id, video_id)
DO UPDATE SET
    order_index = EXCLUDED.order_index,
    source_kind = EXCLUDED.source_kind;
