package playlist

const (
	NameKeyContinueWatching = "VIDEO_PLAYLIST_NAME_CONTINUE_WATCHING"
	NameKeyRecommended      = "VIDEO_PLAYLIST_NAME_RECOMMENDED"
	NameKeyMixed            = "VIDEO_PLAYLIST_NAME_MIXED"
	NameKeyUnknownSeries    = "VIDEO_PLAYLIST_NAME_UNKNOWN_SERIES"
	NameKeyVideos           = "VIDEO_PLAYLIST_NAME_VIDEOS"
)

var classificationNameKeys = map[VideoClassification]string{
	ClassSeries:   "VIDEO_PLAYLIST_NAME_SERIES",
	ClassMovie:    "VIDEO_PLAYLIST_NAME_MOVIES",
	ClassAnime:    "VIDEO_PLAYLIST_NAME_ANIMES",
	ClassCourse:   "VIDEO_PLAYLIST_NAME_COURSES",
	ClassPersonal: "VIDEO_PLAYLIST_NAME_PERSONAL",
	ClassClip:     "VIDEO_PLAYLIST_NAME_CLIPS",
	ClassMusic:    "VIDEO_PLAYLIST_NAME_MUSIC_VIDEOS",
	ClassProgram:  "VIDEO_PLAYLIST_NAME_PROGRAMS",
}

func classificationNameKey(class VideoClassification) string {
	if nameKey, ok := classificationNameKeys[class]; ok {
		return nameKey
	}
	return NameKeyVideos
}
