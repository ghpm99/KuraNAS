SELECT
    vp.id,
    vp.name,
    vp.type
FROM
    video_playlist vp
    INNER JOIN video_playlist_item vpi ON vpi.playlist_id = vp.id
WHERE
    vpi.video_id = $1
    AND vp.is_hidden = FALSE
ORDER BY
    vp.id;
