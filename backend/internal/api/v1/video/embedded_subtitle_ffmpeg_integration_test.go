package video

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEmbeddedSubtitleRoundTripWithRealFFmpeg(t *testing.T) {
	for _, executableName := range []string{ffmpegExecutableName, ffprobeExecutableName} {
		if _, err := exec.LookPath(executableName); err != nil {
			t.Skipf("%s not installed", executableName)
		}
	}
	workDir := t.TempDir()
	sidecarPath := filepath.Join(workDir, "legenda.srt")
	if err := os.WriteFile(sidecarPath, []byte("1\n00:00:00,000 --> 00:00:01,000\nOla mundo\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	videoPath := filepath.Join(workDir, "filme.mkv")
	buildCommand := exec.Command(ffmpegExecutableName, "-v", "error", "-y",
		"-f", "lavfi", "-i", "color=c=black:s=64x64:d=2", "-i", sidecarPath,
		"-map", "0:v", "-map", "1", "-c:v", "mpeg4", "-c:s", "srt",
		"-metadata:s:s:0", "language=por", "-metadata:s:s:0", "title=Portugues", videoPath)
	if output, err := buildCommand.CombinedOutput(); err != nil {
		t.Skipf("ffmpeg could not build the fixture: %v %s", err, output)
	}
	service := NewEmbeddedSubtitleService()
	video := embeddedSubtitleVideoForTest(t, 9199)
	video.Path = videoPath
	video.Size = 1
	video.UpdatedAt = time.Now()

	subtitles := service.ListEmbeddedSubtitles(video)

	if len(subtitles) != 1 || subtitles[0].Language != "pt" || subtitles[0].Label != "Portugues" || subtitles[0].Source != "embedded" {
		t.Fatalf("unexpected subtitles %+v", subtitles)
	}
	webVtt, err := service.ReadEmbeddedSubtitleWebVtt(video, *subtitles[0].StreamIndex)
	if err != nil || !strings.HasPrefix(string(webVtt), "WEBVTT") || !strings.Contains(string(webVtt), "Ola mundo") {
		t.Fatalf("unexpected webvtt %q %v", webVtt, err)
	}
}
