ALTER TABLE video_playlist_item
    DROP CONSTRAINT IF EXISTS video_playlist_item_playlist_id_order_index_key;

ALTER TABLE video_playlist_item
    ADD CONSTRAINT video_playlist_item_playlist_id_order_index_key
    UNIQUE (playlist_id, order_index)
    DEFERRABLE INITIALLY DEFERRED;
