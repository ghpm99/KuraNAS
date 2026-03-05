INSERT INTO video_playlist_item (
    playlist_id,
    video_id,
    order_index,
    source_kind
)
VALUES (
    $1,
    $2,
    COALESCE((SELECT MAX(order_index) + 1 FROM video_playlist_item WHERE playlist_id = $1), 0),
    'manual'
)
ON CONFLICT (playlist_id, video_id)
DO UPDATE SET
    source_kind = 'manual';
