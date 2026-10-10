DELETE FROM video_playlist vp
WHERE vp.is_auto = TRUE
  AND NOT (vp.id = ANY($1::int[]))
  AND NOT EXISTS (
      SELECT 1
      FROM video_playlist_item vpi
      WHERE vpi.playlist_id = vp.id
        AND vpi.source_kind = 'manual'
  )
  AND NOT EXISTS (
      SELECT 1
      FROM video_playback_state vps
      WHERE vps.playlist_id = vp.id
  );
