UPDATE video_playlist AS playlist
SET name_key = generated.name_key
FROM (
    VALUES
        ('continue:%', 'Continue assistindo', 'VIDEO_PLAYLIST_NAME_CONTINUE_WATCHING'),
        ('favorites:%', 'Recomendados para voce', 'VIDEO_PLAYLIST_NAME_RECOMMENDED'),
        ('random:smart_mix', 'Mix variado', 'VIDEO_PLAYLIST_NAME_MIXED'),
        ('related:series', 'Series', 'VIDEO_PLAYLIST_NAME_SERIES'),
        ('related:movie', 'Filmes', 'VIDEO_PLAYLIST_NAME_MOVIES'),
        ('related:anime', 'Animes', 'VIDEO_PLAYLIST_NAME_ANIMES'),
        ('related:course', 'Cursos', 'VIDEO_PLAYLIST_NAME_COURSES'),
        ('related:personal', 'Videos pessoais', 'VIDEO_PLAYLIST_NAME_PERSONAL'),
        ('related:clip', 'Clips', 'VIDEO_PLAYLIST_NAME_CLIPS'),
        ('related:music_video', 'Videos musicais', 'VIDEO_PLAYLIST_NAME_MUSIC_VIDEOS'),
        ('related:program', 'Programas', 'VIDEO_PLAYLIST_NAME_PROGRAMS')
) AS generated(source_path_pattern, generated_name, name_key)
WHERE playlist.is_auto = TRUE
  AND playlist.name_key IS NULL
  AND playlist.source_path LIKE generated.source_path_pattern
  AND playlist.name = generated.generated_name;
