package video

import "encoding/json"

type AudioTrack struct {
	StreamIndex int    `json:"stream_index"`
	Language    string `json:"language"`
	Title       string `json:"title"`
	Codec       string `json:"codec"`
	Channels    int    `json:"channels"`
	IsDefault   bool   `json:"is_default"`
}

type AudioTracksDto struct {
	Items []AudioTrack `json:"items"`
}

type audioProbeOutput struct {
	Streams []audioProbeStream `json:"streams"`
}

type audioProbeStream struct {
	Index       int               `json:"index"`
	CodecName   string            `json:"codec_name"`
	Channels    int               `json:"channels"`
	Disposition map[string]int    `json:"disposition"`
	Tags        map[string]string `json:"tags"`
}

func ParseAudioTracks(probeJSON []byte) ([]AudioTrack, error) {
	var probeOutput audioProbeOutput
	if err := json.Unmarshal(probeJSON, &probeOutput); err != nil {
		return nil, err
	}
	audioTracks := []AudioTrack{}
	for _, stream := range probeOutput.Streams {
		audioTracks = append(audioTracks, buildAudioTrack(stream))
	}
	return audioTracks, nil
}

func buildAudioTrack(stream audioProbeStream) AudioTrack {
	language, _ := normalizeLanguageTag(streamTag(stream.Tags, "language"))
	return AudioTrack{
		StreamIndex: stream.Index,
		Language:    language,
		Title:       streamTag(stream.Tags, "title"),
		Codec:       stream.CodecName,
		Channels:    stream.Channels,
		IsDefault:   stream.Disposition["default"] == 1,
	}
}

func containsAudioStreamIndex(audioTracks []AudioTrack, streamIndex int) bool {
	for _, track := range audioTracks {
		if track.StreamIndex == streamIndex {
			return true
		}
	}
	return false
}
