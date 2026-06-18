package captures

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/libraries"
)

type librariesProviderMock struct {
	getByCategoryFn func(category libraries.LibraryCategory) (libraries.LibraryDto, error)
}

func (m *librariesProviderMock) GetLibraryByCategory(category libraries.LibraryCategory) (libraries.LibraryDto, error) {
	if m.getByCategoryFn != nil {
		return m.getByCategoryFn(category)
	}
	return libraries.LibraryDto{}, nil
}

type filesProviderMock struct {
	createFileFn func(fileDto files.FileDto) (files.FileDto, error)
	deleteFn     func(id int) error
}

func (m *filesProviderMock) CreateFile(fileDto files.FileDto) (files.FileDto, error) {
	if m.createFileFn != nil {
		return m.createFileFn(fileDto)
	}
	fileDto.ID = 99
	return fileDto, nil
}

func (m *filesProviderMock) DeleteFileRecord(id int) error {
	if m.deleteFn != nil {
		return m.deleteFn(id)
	}
	return nil
}

func intPtr(v int) *int { return &v }

func newPromoteServiceForTest(mock *repoMock, lib LibrariesProviderInterface, fp FilesProviderInterface) *Service {
	return &Service{
		Repository:          mock,
		LibrariesProvider:   lib,
		FilesProvider:       fp,
		NotificationService: &notificationServiceMock{},
	}
}

func TestPromoteCaptureEpisodeMovesAndPersists(t *testing.T) {
	dir := t.TempDir()
	videosDir := filepath.Join(dir, "videos")
	stagingDir := filepath.Join(dir, "capturas", "my_show")
	if err := os.MkdirAll(stagingDir, 0755); err != nil {
		t.Fatal(err)
	}
	stagingFile := filepath.Join(stagingDir, "recording.mp4")
	if err := os.WriteFile(stagingFile, []byte("video-bytes"), 0644); err != nil {
		t.Fatal(err)
	}

	// Stub the ffmpeg remux: copy staging->final so the test exercises the
	// remux-success path without invoking a real ffmpeg.
	swapRemux(t, func(src, dest string) error {
		data, err := os.ReadFile(src)
		if err != nil {
			return err
		}
		return os.WriteFile(dest, data, 0644)
	})

	var promoted CaptureModel
	mock := &repoMock{
		getByIDFn: func(id int) (CaptureModel, error) {
			return CaptureModel{
				ID:          id,
				Name:        "my_show",
				FileName:    "recording.mp4",
				FilePath:    stagingFile,
				Size:        11,
				RawMetadata: json.RawMessage(`{"title":"My Show","season":1,"episode":2,"episode_title":"Pilot","platform":"crunchyroll"}`),
			}, nil
		},
		updatePromotionFn: func(tx *sql.Tx, capture CaptureModel) error {
			promoted = capture
			return nil
		},
	}
	lib := &librariesProviderMock{
		getByCategoryFn: func(category libraries.LibraryCategory) (libraries.LibraryDto, error) {
			if category != libraries.LibraryCategoryVideos {
				t.Fatalf("expected videos category, got %s", category)
			}
			return libraries.LibraryDto{Path: videosDir}, nil
		},
	}
	fp := &filesProviderMock{}
	service := newPromoteServiceForTest(mock, lib, fp)

	if err := service.PromoteCapture(1); err != nil {
		t.Fatalf("PromoteCapture returned error: %v", err)
	}

	expectedPath := filepath.Join(videosDir, "My Show", "Temporada 1", "E2 - Pilot.mp4")
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected recording at %s: %v", expectedPath, err)
	}
	if _, err := os.Stat(stagingDir); !os.IsNotExist(err) {
		t.Fatalf("expected staging dir removed, stat err = %v", err)
	}
	if promoted.Status != CaptureStatusPromoted {
		t.Fatalf("expected promoted status, got %s", promoted.Status)
	}
	if promoted.FileID == nil || *promoted.FileID != 99 {
		t.Fatalf("expected file id 99, got %v", promoted.FileID)
	}
	if promoted.FilePath != expectedPath {
		t.Fatalf("expected capture path %s, got %s", expectedPath, promoted.FilePath)
	}
	if promoted.Episode == nil || *promoted.Episode != 2 {
		t.Fatalf("expected episode 2, got %v", promoted.Episode)
	}
}

func TestPromoteCaptureMoveFailureRollsBack(t *testing.T) {
	dir := t.TempDir()
	videosDir := filepath.Join(dir, "videos")

	// No staging file on disk -> the move fails, triggering rollback. The remux
	// is stubbed to fail (as a missing ffmpeg would), so the fallback move runs
	// and also fails on the missing source.
	staging := filepath.Join(dir, "capturas", "gone", "missing.mp4")
	swapRemux(t, func(src, dest string) error { return os.ErrNotExist })

	deleted := 0
	var failedStatus CaptureStatus
	mock := &repoMock{
		getByIDFn: func(id int) (CaptureModel, error) {
			return CaptureModel{ID: id, Name: "gone", FileName: "missing.mp4", FilePath: staging, RawMetadata: json.RawMessage(`{"title":"Gone"}`)}, nil
		},
		updateStatusFn: func(tx *sql.Tx, id int, status CaptureStatus, fileID *int) error {
			failedStatus = status
			if fileID != nil {
				t.Fatalf("expected nil file id on rollback, got %v", *fileID)
			}
			return nil
		},
	}
	lib := &librariesProviderMock{getByCategoryFn: func(libraries.LibraryCategory) (libraries.LibraryDto, error) {
		return libraries.LibraryDto{Path: videosDir}, nil
	}}
	fp := &filesProviderMock{deleteFn: func(id int) error { deleted++; return nil }}
	service := newPromoteServiceForTest(mock, lib, fp)

	if err := service.PromoteCapture(1); err == nil {
		t.Fatal("expected move failure error")
	}
	if deleted != 1 {
		t.Fatalf("expected home_file rollback delete, got %d calls", deleted)
	}
	if failedStatus != CaptureStatusFailed {
		t.Fatalf("expected failed status, got %s", failedStatus)
	}
}

// swapRemux replaces the package-level remux function for the duration of a test.
func swapRemux(t *testing.T, fn func(src, dest string) error) {
	t.Helper()
	prev := remuxRecording
	remuxRecording = fn
	t.Cleanup(func() { remuxRecording = prev })
}

func TestIsRemuxableContainer(t *testing.T) {
	cases := map[string]bool{
		"video.mp4":    true,
		"video.MP4":    true,
		"clip.m4v":     true,
		"clip.mov":     true,
		"capture.webm": false,
		"audio.mp3":    false,
		"noext":        false,
	}
	for name, want := range cases {
		if got := isRemuxableContainer(name); got != want {
			t.Fatalf("isRemuxableContainer(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestPlaceRecordingInLibraryRemuxesMP4(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "staging", "rec.mp4")
	if err := os.MkdirAll(filepath.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("raw"), 0644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "lib", "show.mp4")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		t.Fatal(err)
	}

	called := false
	swapRemux(t, func(s, d string) error {
		called = true
		return os.WriteFile(d, []byte("remuxed"), 0644)
	})

	if err := placeRecordingInLibrary(src, dest); err != nil {
		t.Fatalf("placeRecordingInLibrary error: %v", err)
	}
	if !called {
		t.Fatal("expected remux to be used for .mp4")
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("expected remuxed file at dest: %v", err)
	}
	if string(data) != "remuxed" {
		t.Fatalf("expected remuxed content, got %q", string(data))
	}
}

func TestPlaceRecordingInLibraryFallsBackOnRemuxError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "staging", "rec.mp4")
	if err := os.MkdirAll(filepath.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("raw"), 0644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "lib", "show.mp4")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		t.Fatal(err)
	}

	swapRemux(t, func(s, d string) error { return os.ErrNotExist })

	if err := placeRecordingInLibrary(src, dest); err != nil {
		t.Fatalf("expected fallback move to succeed: %v", err)
	}
	if data, err := os.ReadFile(dest); err != nil || string(data) != "raw" {
		t.Fatalf("expected moved file at dest (data=%q, err=%v)", string(data), err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("expected source consumed by move, stat err = %v", err)
	}
}

func TestPlaceRecordingInLibrarySkipsRemuxForNonMP4(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "staging", "rec.webm")
	if err := os.MkdirAll(filepath.Dir(src), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(src, []byte("webm-bytes"), 0644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(dir, "lib", "show.webm")
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		t.Fatal(err)
	}

	swapRemux(t, func(s, d string) error {
		t.Fatal("remux must not be called for non-MP4 containers")
		return nil
	})

	if err := placeRecordingInLibrary(src, dest); err != nil {
		t.Fatalf("placeRecordingInLibrary error: %v", err)
	}
	if data, err := os.ReadFile(dest); err != nil || string(data) != "webm-bytes" {
		t.Fatalf("expected moved webm at dest (data=%q, err=%v)", string(data), err)
	}
}

func TestFfmpegRemuxForPlaybackErrorsOnBadInput(t *testing.T) {
	dir := t.TempDir()
	dest := filepath.Join(dir, "out.mp4")
	// Whether ffmpeg is absent (exec error) or present (it rejects the missing
	// input), the result must be an error with no partial output left behind.
	if err := ffmpegRemuxForPlayback(filepath.Join(dir, "nope.mp4"), dest); err == nil {
		t.Fatal("expected error remuxing a missing source")
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("expected no partial output, stat err = %v", err)
	}
}

func TestPromoteCaptureRequiresDependencies(t *testing.T) {
	service := &Service{Repository: &repoMock{}}
	if err := service.PromoteCapture(1); err == nil {
		t.Fatal("expected error when promotion dependencies are nil")
	}
}

func TestBuildCaptureRelPath(t *testing.T) {
	tests := []struct {
		name    string
		meta    captureMetadata
		capture CaptureModel
		want    string
	}{
		{
			name:    "episode with season",
			meta:    captureMetadata{Title: "Show", Season: intPtr(2), Episode: intPtr(5), EpisodeTitle: "Ep Title"},
			capture: CaptureModel{Name: "rec", FileName: "rec.mp4"},
			want:    filepath.Join("Show", "Temporada 2", "E5 - Ep Title.mp4"),
		},
		{
			name:    "episode without season or title",
			meta:    captureMetadata{Title: "Show", Episode: intPtr(7)},
			capture: CaptureModel{Name: "rec", FileName: "rec.mp4"},
			want:    filepath.Join("Show", "E7.mp4"),
		},
		{
			name:    "movie with year",
			meta:    captureMetadata{Title: "Movie", ReleaseYear: intPtr(1999)},
			capture: CaptureModel{Name: "rec", FileName: "rec.mp4"},
			want:    "Movie (1999).mp4",
		},
		{
			name:    "no title falls back to recording name",
			meta:    captureMetadata{},
			capture: CaptureModel{Name: "my recording", FileName: "rec.mp4"},
			want:    "my recording.mp4",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := buildCaptureRelPath(tc.meta, tc.capture)
			if got != tc.want {
				t.Fatalf("buildCaptureRelPath = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestCollisionAvoidantPath(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "video.mp4")

	if got := collisionAvoidantPath(base); got != base {
		t.Fatalf("expected free path returned as-is, got %s", got)
	}

	if err := os.WriteFile(base, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, "video (2).mp4")
	if got := collisionAvoidantPath(base); got != want {
		t.Fatalf("expected %s, got %s", want, got)
	}

	if err := os.WriteFile(want, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	want3 := filepath.Join(dir, "video (3).mp4")
	if got := collisionAvoidantPath(base); got != want3 {
		t.Fatalf("expected %s, got %s", want3, got)
	}
}

func TestDownloadPosterRejectsNonHTTPS(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write([]byte("\x89PNG fake"))
	}))
	defer server.Close()

	// httptest serves http, not https; downloadPoster only accepts https, so a
	// plain http URL must be a no-op (no file written).
	service := &Service{}
	service.downloadPoster(captureMetadata{ThumbnailURL: server.URL}, 4242)
	if _, err := os.Stat(posterSourcePath(4242)); !os.IsNotExist(err) {
		_ = os.Remove(posterSourcePath(4242))
		t.Fatal("expected http poster url to be rejected")
	}
}

func TestFetchPosterImage(t *testing.T) {
	t.Run("returns image bytes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("\x89PNG fake-bytes"))
		}))
		defer server.Close()

		data, err := fetchPosterImage(server.URL)
		if err != nil {
			t.Fatalf("expected success, got %v", err)
		}
		if len(data) == 0 {
			t.Fatal("expected non-empty poster bytes")
		}
	})

	t.Run("rejects non-image content type", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html>"))
		}))
		defer server.Close()

		if _, err := fetchPosterImage(server.URL); err == nil {
			t.Fatal("expected non-image rejection")
		}
	})

	t.Run("rejects bad status", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		if _, err := fetchPosterImage(server.URL); err == nil {
			t.Fatal("expected bad-status rejection")
		}
	})
}

func TestWritePosterSource(t *testing.T) {
	const fileID = 91234
	t.Cleanup(func() { _ = os.Remove(posterSourcePath(fileID)) })

	if err := writePosterSource(fileID, []byte("poster")); err != nil {
		t.Fatalf("writePosterSource error: %v", err)
	}
	data, err := os.ReadFile(posterSourcePath(fileID))
	if err != nil {
		t.Fatalf("expected poster written: %v", err)
	}
	if string(data) != "poster" {
		t.Fatalf("unexpected poster content: %q", string(data))
	}
}
