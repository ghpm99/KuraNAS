package video

type PlaylistSeasonSummaryModel struct {
	Season       int
	EpisodeCount int
	WatchedCount int
}

type PlaylistResumeModel struct {
	VideoID         int
	Season          *int
	Episode         *int
	PositionSeconds float64
	Mode            string
}
