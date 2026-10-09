package music

type AudioSummaryDto struct {
	Title         string  `json:"title"`
	Artist        string  `json:"artist"`
	Album         string  `json:"album"`
	Genre         string  `json:"genre"`
	Year          string  `json:"year"`
	TrackNumber   string  `json:"track_number"`
	LengthSeconds float64 `json:"length"`
	Bitrate       int     `json:"bitrate"`
	SampleRate    int     `json:"sample_rate"`
	Channels      int     `json:"channels"`
}
