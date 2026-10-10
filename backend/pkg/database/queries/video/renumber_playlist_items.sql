UPDATE video_playlist_item AS item
SET order_index = ranked.new_order_index
FROM (
    SELECT
        id,
        ROW_NUMBER() OVER (
            ORDER BY (source_kind <> 'auto'), order_index, id
        ) - 1 AS new_order_index
    FROM video_playlist_item
    WHERE playlist_id = $1
) AS ranked
WHERE item.id = ranked.id
  AND item.order_index <> ranked.new_order_index;
