UPDATE video_playlist_item AS item
SET order_index = positions.new_order_index
FROM UNNEST($2::int[], $3::int[]) AS positions(item_id, new_order_index)
WHERE item.playlist_id = $1
  AND item.id = positions.item_id
  AND item.order_index <> positions.new_order_index;
