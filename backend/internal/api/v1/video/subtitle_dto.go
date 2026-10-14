package video

import "strconv"

const (
	SubtitleSourceExternal = "external"
	SubtitleSourceEmbedded = "embedded"
)

type SubtitleFileModel struct {
	ID   int
	Name string
}

// SubtitleDto describes one selectable subtitle of a video.
//
// External subtitles carry the file id of the sidecar in ID and have no Key.
// Embedded subtitles are not files: ID is 0 and Key identifies the track as
// "<video_id>:<stream_index>", which cannot collide with any external ID.
// StreamIndex is the absolute ffprobe stream index inside the container.
type SubtitleDto struct {
	ID          int    `json:"id"`
	Source      string `json:"source"`
	Key         string `json:"key,omitempty"`
	StreamIndex *int   `json:"stream_index,omitempty"`
	Label       string `json:"label"`
	Language    string `json:"language"`
	Format      string `json:"format"`
	IsForced    bool   `json:"is_forced"`
}

func BuildEmbeddedSubtitleKey(videoID int, streamIndex int) string {
	return strconv.Itoa(videoID) + ":" + strconv.Itoa(streamIndex)
}

func BuildEmbeddedSubtitleDto(videoID int, track EmbeddedSubtitleTrack) SubtitleDto {
	streamIndex := track.StreamIndex
	return SubtitleDto{
		Source:      SubtitleSourceEmbedded,
		Key:         BuildEmbeddedSubtitleKey(videoID, streamIndex),
		StreamIndex: &streamIndex,
		Label:       track.Title,
		Language:    track.Language,
		Format:      SubtitleFormatVtt,
		IsForced:    track.IsForced,
	}
}
