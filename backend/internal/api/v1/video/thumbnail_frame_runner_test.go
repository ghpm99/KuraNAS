package video

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"testing"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
)

type recordedFrameCommands struct {
	startOffsets []string
	failFirst    int
}

func (recorded *recordedFrameCommands) run(_ context.Context, arguments []string) error {
	for index, argument := range arguments {
		if argument == "-ss" {
			recorded.startOffsets = append(recorded.startOffsets, arguments[index+1])
		}
	}
	if len(recorded.startOffsets) <= recorded.failFirst {
		return errors.New("ffmpeg failed")
	}
	return os.WriteFile(arguments[len(arguments)-1], []byte("frame"), 0644)
}

func newFrameTestVideo(t *testing.T) (files.FileDto, *Service, *recordedFrameCommands) {
	t.Helper()
	ensureVideoTestIcons(t)
	videoPath := filepath.Join(t.TempDir(), "clip.mp4")
	if err := os.WriteFile(videoPath, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cacheDir := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video")
		_ = os.Remove(filepath.Join(cacheDir, "901_320x180_f2.png"))
		_ = os.Remove(filepath.Join(cacheDir, "901_320x180_preview_f2.gif"))
	})
	recorded := &recordedFrameCommands{}
	repository := &videoRepoMock{getVideoDurationSecondsFn: func(int) (sql.NullFloat64, error) {
		return sql.NullFloat64{Float64: 600, Valid: true}, nil
	}}
	service := newVideoBrowseServiceForTest(t, repository)
	service.FrameCommandRunner = recorded.run
	return files.FileDto{ID: 901, Path: videoPath, Type: files.File}, service, recorded
}

func TestGetVideoThumbnailUsesProportionalOffset(t *testing.T) {
	fileDto, service, recorded := newFrameTestVideo(t)

	thumbnail, err := service.GetVideoThumbnail(fileDto, 320, 180)
	if err != nil || string(thumbnail) != "frame" {
		t.Fatalf("unexpected thumbnail %q err %v", thumbnail, err)
	}
	if len(recorded.startOffsets) != 1 || recorded.startOffsets[0] != "60.000" {
		t.Fatalf("expected a single ffmpeg call at 60.000, got %v", recorded.startOffsets)
	}
}

func TestGetVideoThumbnailRetriesAtZeroWhenFirstAttemptFails(t *testing.T) {
	fileDto, service, recorded := newFrameTestVideo(t)
	recorded.failFirst = 1

	thumbnail, err := service.GetVideoThumbnail(fileDto, 320, 180)
	if err != nil || string(thumbnail) != "frame" {
		t.Fatalf("unexpected thumbnail %q err %v", thumbnail, err)
	}
	if len(recorded.startOffsets) != 2 || recorded.startOffsets[1] != "0.000" {
		t.Fatalf("expected retry at 0.000, got %v", recorded.startOffsets)
	}
}

func TestGetVideoThumbnailFallsBackToIconAfterTwoFailures(t *testing.T) {
	fileDto, service, recorded := newFrameTestVideo(t)
	recorded.failFirst = 10

	thumbnail, err := service.GetVideoThumbnail(fileDto, 320, 180)
	if err != nil || len(thumbnail) == 0 || string(thumbnail) == "frame" {
		t.Fatalf("expected icon fallback, got %d bytes err %v", len(thumbnail), err)
	}
	if len(recorded.startOffsets) != 2 {
		t.Fatalf("expected exactly two attempts, got %v", recorded.startOffsets)
	}
}

func TestGetVideoThumbnailDoesNotRetryWhenOffsetIsAlreadyZero(t *testing.T) {
	fileDto, service, recorded := newFrameTestVideo(t)
	service.Repository.(*videoRepoMock).getVideoDurationSecondsFn = func(int) (sql.NullFloat64, error) {
		return sql.NullFloat64{Float64: 0.5, Valid: true}, nil
	}
	recorded.failFirst = 10

	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}
	if len(recorded.startOffsets) != 1 {
		t.Fatalf("expected a single attempt, got %v", recorded.startOffsets)
	}
}

func TestGetVideoThumbnailTreatsDurationLookupErrorAsUnknown(t *testing.T) {
	fileDto, service, recorded := newFrameTestVideo(t)
	service.Repository.(*videoRepoMock).getVideoDurationSecondsFn = func(int) (sql.NullFloat64, error) {
		return sql.NullFloat64{}, errors.New("db down")
	}

	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}
	if len(recorded.startOffsets) != 1 || recorded.startOffsets[0] != "3.000" {
		t.Fatalf("expected legacy offset 3.000, got %v", recorded.startOffsets)
	}
}

func TestGetVideoPreviewGifRetriesAtZeroAndUsesVersionedCacheName(t *testing.T) {
	fileDto, service, recorded := newFrameTestVideo(t)
	recorded.failFirst = 1

	preview, err := service.GetVideoPreviewGif(fileDto, 320, 180)
	if err != nil || string(preview) != "frame" {
		t.Fatalf("unexpected preview %q err %v", preview, err)
	}
	if len(recorded.startOffsets) != 2 || recorded.startOffsets[0] != "60.000" || recorded.startOffsets[1] != "0.000" {
		t.Fatalf("unexpected offsets %v", recorded.startOffsets)
	}
}
