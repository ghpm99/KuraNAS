package ingest

import "nas-go/api/pkg/i18n"

// preset is a fixed, closed mapping from a UI-facing key to the yt-dlp format
// arguments it expands into. The map is intentionally small: a download feature
// that exposes arbitrary yt-dlp flags is a footgun. LabelKey is an i18n key
// resolved server-side before the list is returned, so non-i18n clients (the
// browser extension) can show it verbatim.
type preset struct {
	Key      string
	LabelKey string
	Args     []string
}

// presets is the source of truth for the allowed download formats. Order is the
// display order in the client dropdown.
var presets = []preset{
	{
		Key:      "audio_mp3",
		LabelKey: "DOWNLOAD_PRESET_AUDIO_MP3",
		Args:     []string{"-x", "--audio-format", "mp3", "--embed-metadata", "--embed-thumbnail"},
	},
	{
		Key:      "video_1080",
		LabelKey: "DOWNLOAD_PRESET_VIDEO_1080",
		Args:     []string{"-f", "bv*[height<=1080]+ba/b[height<=1080]", "--merge-output-format", "mp4", "--embed-metadata"},
	},
	{
		Key:      "video_best",
		LabelKey: "DOWNLOAD_PRESET_VIDEO_BEST",
		Args:     []string{"-f", "bv*+ba/b", "--merge-output-format", "mp4", "--embed-metadata"},
	},
}

// ResolvePreset returns the format arguments for a preset key. The second result
// is false when the key is unknown, so callers reject it instead of running an
// unconfigured download.
func ResolvePreset(key string) ([]string, bool) {
	for _, p := range presets {
		if p.Key == key {
			// Copy so callers cannot mutate the shared slice.
			args := make([]string, len(p.Args))
			copy(args, p.Args)
			return args, true
		}
	}
	return nil, false
}

// availablePresets returns the selectable presets as transport DTOs, with the
// label already resolved server-side so non-i18n clients show it verbatim.
func availablePresets() []PresetDto {
	out := make([]PresetDto, 0, len(presets))
	for _, p := range presets {
		out = append(out, PresetDto{Key: p.Key, Label: i18n.GetMessage(p.LabelKey)})
	}
	return out
}
