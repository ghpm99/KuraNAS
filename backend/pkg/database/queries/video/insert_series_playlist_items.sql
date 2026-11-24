INSERT INTO video_playlist_item (
    playlist_id,
    video_id,
    order_index,
    source_kind,
    season,
    episode
)
SELECT
    $1,
    entry.video_id,
    entry.position - 1,
    $3,
    entry.season,
    entry.episode
FROM UNNEST($2::int[], $4::int[], $5::int[]) WITH ORDINALITY AS entry(video_id, season, episode, position)
ON CONFLICT (playlist_id, video_id)
DO UPDATE SET
    order_index = EXCLUDED.order_index,
    source_kind = EXCLUDED.source_kind,
    season = EXCLUDED.season,
    episode = EXCLUDED.episode;
