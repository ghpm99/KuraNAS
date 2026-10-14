package video

import (
	"encoding/json"
	"strings"
)

type EmbeddedSubtitleTrack struct {
	StreamIndex int
	Codec       string
	Language    string
	Title       string
	IsForced    bool
}

type embeddedSubtitleProbeOutput struct {
	Streams []embeddedSubtitleProbeStream `json:"streams"`
}

type embeddedSubtitleProbeStream struct {
	Index       int               `json:"index"`
	CodecName   string            `json:"codec_name"`
	Disposition map[string]int    `json:"disposition"`
	Tags        map[string]string `json:"tags"`
}

func ParseEmbeddedSubtitleTracks(probeJSON []byte) (textTracks []EmbeddedSubtitleTrack, skippedTrackCount int, err error) {
	var probeOutput embeddedSubtitleProbeOutput
	if err := json.Unmarshal(probeJSON, &probeOutput); err != nil {
		return nil, 0, err
	}
	textTracks = []EmbeddedSubtitleTrack{}
	for _, stream := range probeOutput.Streams {
		if !IsTextSubtitleCodec(stream.CodecName) {
			skippedTrackCount++
			continue
		}
		textTracks = append(textTracks, buildEmbeddedSubtitleTrack(stream))
	}
	return textTracks, skippedTrackCount, nil
}

func buildEmbeddedSubtitleTrack(stream embeddedSubtitleProbeStream) EmbeddedSubtitleTrack {
	language, _ := normalizeLanguageTag(streamTag(stream.Tags, "language"))
	return EmbeddedSubtitleTrack{
		StreamIndex: stream.Index,
		Codec:       stream.CodecName,
		Language:    language,
		Title:       streamTag(stream.Tags, "title"),
		IsForced:    stream.Disposition["forced"] == 1,
	}
}

func streamTag(tags map[string]string, tagName string) string {
	for name, tagValue := range tags {
		if strings.EqualFold(name, tagName) {
			return strings.TrimSpace(tagValue)
		}
	}
	return ""
}
