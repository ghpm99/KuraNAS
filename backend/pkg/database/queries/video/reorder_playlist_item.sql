UPDATE video_playlist_item AS vpi
SET order_index = batch.new_order
FROM (
    SELECT unnest($2::int[]) AS vid,
           unnest($3::int[]) AS new_order
) AS batch
WHERE vpi.playlist_id = $1
  AND vpi.video_id = batch.vid;
