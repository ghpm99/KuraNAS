package video

import (
	"reflect"
	"testing"
)

func TestBuildMasterFrameArguments(t *testing.T) {
	actual := buildMasterFrameArguments("/media/a.mkv", 12.5, "/cache/1_master_f2.jpg")
	expected := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-ss", "12.500",
		"-i", "/media/a.mkv",
		"-frames:v", "1",
		"-vf", `scale=min(1280\,iw):min(720\,ih):force_original_aspect_ratio=decrease`,
		"-q:v", "2",
		"-update", "1",
		"/cache/1_master_f2.jpg",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected arguments:\n%v\nwant:\n%v", actual, expected)
	}
}

func TestBuildPreviewGifArguments(t *testing.T) {
	actual := buildPreviewGifArguments("/media/a.mkv", 0, 320, 180, "/cache/1_320x180_preview_f2.gif")
	expected := []string{
		"-hide_banner", "-loglevel", "error", "-y",
		"-ss", "0.000",
		"-t", "2.500",
		"-i", "/media/a.mkv",
		"-vf", "fps=4,scale=320:180:force_original_aspect_ratio=decrease,pad=320:180:(ow-iw)/2:(oh-ih)/2:black",
		"-loop", "0",
		"/cache/1_320x180_preview_f2.gif",
	}
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("unexpected arguments:\n%v\nwant:\n%v", actual, expected)
	}
}
