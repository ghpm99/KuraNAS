package video

import (
	"os"
	"reflect"
	"testing"
)

func TestParseEmbeddedSubtitleTracksKeepsTextTracksAndCountsImageTracks(t *testing.T) {
	probeJSON, err := os.ReadFile("testdata/ffprobe_subtitle_streams.json")
	if err != nil {
		t.Fatal(err)
	}

	tracks, skippedTrackCount, err := ParseEmbeddedSubtitleTracks(probeJSON)

	expectedTracks := []EmbeddedSubtitleTrack{
		{StreamIndex: 2, Codec: "subrip", Language: "pt", Title: "Portugues Brasil"},
		{StreamIndex: 3, Codec: "ass", Language: "en", Title: "Signs and Songs", IsForced: true},
		{StreamIndex: 5, Codec: "mov_text", Language: "es"},
	}
	if err != nil || skippedTrackCount != 2 || !reflect.DeepEqual(tracks, expectedTracks) {
		t.Fatalf("unexpected %+v skipped=%d err=%v", tracks, skippedTrackCount, err)
	}
}

func TestParseEmbeddedSubtitleTracksHandlesEmptyAndInvalidOutput(t *testing.T) {
	tracks, skippedTrackCount, err := ParseEmbeddedSubtitleTracks([]byte(`{"streams":[]}`))
	if err != nil || skippedTrackCount != 0 || tracks == nil || len(tracks) != 0 {
		t.Fatalf("unexpected %+v %d %v", tracks, skippedTrackCount, err)
	}
	if _, _, err := ParseEmbeddedSubtitleTracks([]byte(`not json`)); err == nil {
		t.Fatal("expected error for invalid json")
	}
}

func TestIsTextSubtitleCodec(t *testing.T) {
	testCases := map[string]bool{
		"subrip": true, "ass": true, "ssa": true, "mov_text": true, "webvtt": true, "text": true, "SubRip": true,
		"hdmv_pgs_subtitle": false, "dvd_subtitle": false, "dvb_subtitle": false, "": false,
	}
	for codecName, expected := range testCases {
		if IsTextSubtitleCodec(codecName) != expected {
			t.Errorf("codec %q expected %v", codecName, expected)
		}
	}
}

func TestBuildEmbeddedSubtitleArguments(t *testing.T) {
	probeArguments := buildEmbeddedSubtitleProbeArguments("/v/a.mkv")
	expectedProbe := []string{"-v", "error", "-select_streams", "s", "-show_streams", "-of", "json", "/v/a.mkv"}
	if !reflect.DeepEqual(probeArguments, expectedProbe) {
		t.Fatalf("unexpected probe arguments %v", probeArguments)
	}
	extractArguments := buildEmbeddedSubtitleExtractArguments("/v/a.mkv", 3, "/tmp/out.vtt")
	expectedExtract := []string{"-hide_banner", "-loglevel", "error", "-y", "-i", "/v/a.mkv", "-map", "0:3", "-f", "webvtt", "/tmp/out.vtt"}
	if !reflect.DeepEqual(extractArguments, expectedExtract) {
		t.Fatalf("unexpected extract arguments %v", extractArguments)
	}
}

func TestBuildEmbeddedSubtitleDto(t *testing.T) {
	subtitle := BuildEmbeddedSubtitleDto(7, EmbeddedSubtitleTrack{StreamIndex: 3, Language: "en", Title: "Full", IsForced: true})

	if subtitle.ID != 0 || subtitle.Source != "embedded" || subtitle.Key != "7:3" || *subtitle.StreamIndex != 3 ||
		subtitle.Label != "Full" || subtitle.Language != "en" || subtitle.Format != "vtt" || !subtitle.IsForced {
		t.Fatalf("unexpected dto %+v", subtitle)
	}
}
