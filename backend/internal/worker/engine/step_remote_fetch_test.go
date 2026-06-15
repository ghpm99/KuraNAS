package engine

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	ingestapi "nas-go/api/internal/api/v1/ingest"
	jobs "nas-go/api/internal/api/v1/jobs"
)

func TestExecuteRemoteFetchStepEarlyErrors(t *testing.T) {
	step := func(p ingestapi.RemoteFetchStepPayload) jobs.StepModel {
		raw, _ := json.Marshal(p)
		return jobs.StepModel{Payload: raw}
	}

	if err := executeRemoteFetchStep(nil, jobs.StepModel{Payload: []byte("{bad")}); err == nil {
		t.Fatal("expected a decode error for malformed payload")
	}
	if err := executeRemoteFetchStep(nil, step(ingestapi.RemoteFetchStepPayload{URL: "https://x.test/v", Preset: "nope", OutputDir: t.TempDir()})); err == nil {
		t.Fatal("expected an error for an unknown preset")
	}
	if err := executeRemoteFetchStep(nil, step(ingestapi.RemoteFetchStepPayload{URL: "https://x.test/v", Preset: "audio_mp3", OutputDir: ""})); err == nil {
		t.Fatal("expected an error for an empty output dir")
	}
	// A missing binary makes cmd.Start fail — exercises the exec path without
	// needing a real yt-dlp installed.
	err := executeRemoteFetchStep(nil, step(ingestapi.RemoteFetchStepPayload{
		URL:       "https://x.test/v",
		Preset:    "audio_mp3",
		OutputDir: t.TempDir(),
		Binary:    "kuranas-no-such-binary-xyz",
	}))
	if err == nil {
		t.Fatal("expected an error when the yt-dlp binary is missing")
	}
}

func TestBuildYtDlpArgs(t *testing.T) {
	args := buildYtDlpArgs([]string{"-x", "--audio-format", "mp3"}, "/tmp/out", "https://x.test/v")
	joined := strings.Join(args, " ")

	for _, want := range []string{"-x", "--audio-format", "mp3", "--no-playlist", "--restrict-filenames", "--no-overwrites", "--newline", "https://x.test/v"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("args missing %q: %v", want, args)
		}
	}
	// The URL must be last so flags never get parsed as the positional arg.
	if args[len(args)-1] != "https://x.test/v" {
		t.Fatalf("expected url last, got %q", args[len(args)-1])
	}
	if !strings.Contains(joined, filepath.Join("/tmp/out", "%(title)s.%(ext)s")) {
		t.Fatalf("output template not rooted at output dir: %v", args)
	}
}

func TestParseDownloadProgress(t *testing.T) {
	cases := []struct {
		line    string
		wantPct int
		wantOK  bool
	}{
		{"[download]   0.0% of 10MiB", 0, true},
		{"[download]  45.6% of 10MiB at 1MiB/s", 45, true},
		{"[download] 100% of 10MiB", 100, true},
		{"[info] something else", 0, false},
		{"no marker 50%", 0, false},
		{"[download] no percent here", 0, false},
	}
	for _, tc := range cases {
		pct, ok := parseDownloadProgress(tc.line)
		if ok != tc.wantOK || (ok && pct != tc.wantPct) {
			t.Errorf("parseDownloadProgress(%q) = (%d,%v), want (%d,%v)", tc.line, pct, ok, tc.wantPct, tc.wantOK)
		}
	}
}

func TestIsTempArtifact(t *testing.T) {
	for _, name := range []string{"video.mp4.part", "audio.m4a.ytdl", "x.TEMP", "y.tmp"} {
		if !isTempArtifact(name) {
			t.Errorf("expected %q to be a temp artifact", name)
		}
	}
	for _, name := range []string{"song.mp3", "movie.mkv"} {
		if isTempArtifact(name) {
			t.Errorf("expected %q to be a finished file", name)
		}
	}
}

func TestMoveFetchedFilesSkipsTempAndHidden(t *testing.T) {
	src := t.TempDir()
	dst := t.TempDir()

	write := func(dir, name string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(src, "song.mp3")
	write(src, "song.mp3.part")
	write(src, ".hidden")

	moved, err := moveFetchedFiles(src, dst)
	if err != nil {
		t.Fatalf("moveFetchedFiles: %v", err)
	}
	if moved != 1 {
		t.Fatalf("expected 1 moved, got %d", moved)
	}
	if _, err := os.Stat(filepath.Join(dst, "song.mp3")); err != nil {
		t.Fatalf("finished file not moved: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "song.mp3.part")); !os.IsNotExist(err) {
		t.Fatal("temp artifact should not have moved")
	}
}

func TestMoveFileCopyFallback(t *testing.T) {
	src := filepath.Join(t.TempDir(), "a.txt")
	dst := filepath.Join(t.TempDir(), "b.txt")
	if err := os.WriteFile(src, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := moveFile(src, dst); err != nil {
		t.Fatalf("moveFile: %v", err)
	}
	data, err := os.ReadFile(dst)
	if err != nil || string(data) != "hello" {
		t.Fatalf("dst content wrong: %q, %v", data, err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatal("src should be gone after move")
	}
}

func TestLastLine(t *testing.T) {
	if got := lastLine("a\nb\n  c  "); got != "c" {
		t.Fatalf("expected 'c', got %q", got)
	}
	if got := lastLine("solo"); got != "solo" {
		t.Fatalf("expected 'solo', got %q", got)
	}
}

func TestDecodeRemoteFetchPayload(t *testing.T) {
	good, _ := json.Marshal(ingestapi.RemoteFetchStepPayload{URL: "https://x.test/v", Preset: "audio_mp3", OutputDir: "/o"})
	payload, err := decodeRemoteFetchPayload(good)
	if err != nil || payload.URL != "https://x.test/v" {
		t.Fatalf("good payload: %+v, %v", payload, err)
	}

	if _, err := decodeRemoteFetchPayload([]byte(`{"preset":"audio_mp3"}`)); err == nil {
		t.Fatal("expected error for payload without url")
	}
	if _, err := decodeRemoteFetchPayload([]byte(`{bad`)); err == nil {
		t.Fatal("expected error for malformed json")
	}
}
