SELECT
    vp.id AS playlist_id,
    vpi.video_id
FROM
    video_playlist vp
    INNER JOIN video_playlist_item vpi ON vpi.playlist_id = vp.id
WHERE
    ($1 = TRUE OR vp.is_hidden = FALSE)
ORDER BY
    vp.id,
    vpi.order_index,
    vpi.id;
