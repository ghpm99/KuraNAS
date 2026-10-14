package video

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
)

type fakeClock struct{ current time.Time }

func (clock *fakeClock) now() time.Time { return clock.current }

func (clock *fakeClock) advance(duration time.Duration) { clock.current = clock.current.Add(duration) }

type togglingFrameRunner struct {
	shouldFail bool
	callCount  int
}

func (runner *togglingFrameRunner) run(_ context.Context, arguments []string) error {
	runner.callCount++
	if runner.shouldFail {
		return errors.New("ffmpeg failed")
	}
	return os.WriteFile(arguments[len(arguments)-1], []byte("frame"), 0644)
}

func newFallbackTestService(t *testing.T) (files.FileDto, *Service, *togglingFrameRunner, *fakeClock) {
	t.Helper()
	fileDto, service, _ := newFrameTestVideo(t)
	runner := &togglingFrameRunner{shouldFail: true}
	clock := &fakeClock{current: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	service.FrameCommandRunner = runner.run
	service.FallbackRegistry.now = clock.now
	return fileDto, service, runner, clock
}

func thumbnailCachePathFor(fileDto files.FileDto) string {
	return filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video", "901_320x180_f2.png")
}

func previewCachePathFor(fileDto files.FileDto) string {
	return filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video", "901_320x180_preview_f2.gif")
}

func TestThumbnailFailureReturnsFallbackWithoutCachingIt(t *testing.T) {
	fileDto, service, _, _ := newFallbackTestService(t)

	thumbnail, err := service.GetVideoThumbnail(fileDto, 320, 180)

	if err != nil || !thumbnail.IsFallback || len(thumbnail.Data) == 0 {
		t.Fatalf("expected fallback, got fallback=%v len=%d err=%v", thumbnail.IsFallback, len(thumbnail.Data), err)
	}
	if _, statErr := os.Stat(thumbnailCachePathFor(fileDto)); !os.IsNotExist(statErr) {
		t.Fatalf("fallback must not be cached, stat err=%v", statErr)
	}
}

func TestPreviewFailureReturnsFallbackWithoutCachingIt(t *testing.T) {
	fileDto, service, _, _ := newFallbackTestService(t)

	preview, err := service.GetVideoPreviewGif(fileDto, 320, 180)

	if err != nil || !preview.IsFallback || len(preview.Data) == 0 {
		t.Fatalf("expected fallback, got fallback=%v len=%d err=%v", preview.IsFallback, len(preview.Data), err)
	}
	if _, statErr := os.Stat(previewCachePathFor(fileDto)); !os.IsNotExist(statErr) {
		t.Fatalf("fallback must not be cached, stat err=%v", statErr)
	}
}

func TestPartialOutputLeftByFailedFFmpegIsRemoved(t *testing.T) {
	fileDto, service, _, _ := newFallbackTestService(t)
	service.FrameCommandRunner = func(_ context.Context, arguments []string) error {
		_ = os.WriteFile(arguments[len(arguments)-1], []byte("partial"), 0644)
		return errors.New("ffmpeg crashed")
	}

	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}

	if _, statErr := os.Stat(thumbnailCachePathFor(fileDto)); !os.IsNotExist(statErr) {
		t.Fatalf("partial output must be removed, stat err=%v", statErr)
	}
}

func TestSecondCallWithinTTLDoesNotRunFFmpeg(t *testing.T) {
	fileDto, service, runner, clock := newFallbackTestService(t)
	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}
	callsAfterFirstRequest := runner.callCount

	clock.advance(thumbnailFallbackTTL - time.Second)
	second, err := service.GetVideoThumbnail(fileDto, 320, 180)

	if err != nil || !second.IsFallback {
		t.Fatalf("expected fallback, err=%v", err)
	}
	if runner.callCount != callsAfterFirstRequest {
		t.Fatalf("expected no new ffmpeg calls, had %d now %d", callsAfterFirstRequest, runner.callCount)
	}
}

func TestCallAfterTTLRunsFFmpegAgain(t *testing.T) {
	fileDto, service, runner, clock := newFallbackTestService(t)
	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}
	callsAfterFirstRequest := runner.callCount

	clock.advance(thumbnailFallbackTTL)
	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}

	if runner.callCount <= callsAfterFirstRequest {
		t.Fatalf("expected ffmpeg to run again after TTL")
	}
}

func TestChangedUpdatedAtInvalidatesNegativeEntry(t *testing.T) {
	fileDto, service, runner, _ := newFallbackTestService(t)
	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}
	callsAfterFirstRequest := runner.callCount

	fileDto.UpdatedAt = fileDto.UpdatedAt.Add(time.Minute)
	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}

	if runner.callCount <= callsAfterFirstRequest {
		t.Fatalf("expected ffmpeg to run again after updated_at change")
	}
}

func TestChangedSizeInvalidatesNegativeEntry(t *testing.T) {
	fileDto, service, runner, _ := newFallbackTestService(t)
	if _, err := service.GetVideoPreviewGif(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}
	callsAfterFirstRequest := runner.callCount

	fileDto.Size += 10
	if _, err := service.GetVideoPreviewGif(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}

	if runner.callCount <= callsAfterFirstRequest {
		t.Fatalf("expected ffmpeg to run again after size change")
	}
}

func TestThumbnailAndPreviewFailuresAreTrackedIndependently(t *testing.T) {
	fileDto, service, runner, _ := newFallbackTestService(t)
	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}
	callsAfterThumbnail := runner.callCount

	if _, err := service.GetVideoPreviewGif(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}

	if runner.callCount <= callsAfterThumbnail {
		t.Fatalf("preview must not be suppressed by thumbnail failure")
	}
}

func TestRecoveredFFmpegProducesCachedRealThumbnail(t *testing.T) {
	fileDto, service, runner, clock := newFallbackTestService(t)
	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
		t.Fatal(err)
	}

	runner.shouldFail = false
	clock.advance(thumbnailFallbackTTL)
	thumbnail, err := service.GetVideoThumbnail(fileDto, 320, 180)

	if err != nil || thumbnail.IsFallback || string(thumbnail.Data) != "frame" {
		t.Fatalf("expected real thumbnail, fallback=%v data=%q err=%v", thumbnail.IsFallback, thumbnail.Data, err)
	}
	cached, readErr := os.ReadFile(thumbnailCachePathFor(fileDto))
	if readErr != nil || string(cached) != "frame" {
		t.Fatalf("expected real thumbnail cached, err=%v", readErr)
	}
	callsAfterRecovery := runner.callCount
	if _, err := service.GetVideoThumbnail(fileDto, 320, 180); err != nil || runner.callCount != callsAfterRecovery {
		t.Fatalf("expected cache hit without ffmpeg, err=%v", err)
	}
}

func TestRegistryEvictsOldestEntryWhenFull(t *testing.T) {
	clock := &fakeClock{current: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	registry := &thumbnailFallbackRegistry{now: clock.now, maxEntries: 2}
	firstFile, secondFile, thirdFile := files.FileDto{ID: 1}, files.FileDto{ID: 2}, files.FileDto{ID: 3}

	registry.recordFailure(newThumbnailFallbackKey(firstFile, 320, 180, generatedImageKindThumbnail), firstFile)
	clock.advance(time.Second)
	registry.recordFailure(newThumbnailFallbackKey(secondFile, 320, 180, generatedImageKindThumbnail), secondFile)
	clock.advance(time.Second)
	registry.recordFailure(newThumbnailFallbackKey(thirdFile, 320, 180, generatedImageKindThumbnail), thirdFile)

	if len(registry.entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(registry.entries))
	}
	if registry.isRecentlyFailed(newThumbnailFallbackKey(firstFile, 320, 180, generatedImageKindThumbnail), firstFile) {
		t.Fatalf("oldest entry should have been evicted")
	}
	if !registry.isRecentlyFailed(newThumbnailFallbackKey(thirdFile, 320, 180, generatedImageKindThumbnail), thirdFile) {
		t.Fatalf("newest entry should be kept")
	}
}

func TestRegistryPrefersEvictingExpiredEntries(t *testing.T) {
	clock := &fakeClock{current: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}
	registry := &thumbnailFallbackRegistry{now: clock.now, maxEntries: 2}
	expiredFile, liveFile, newFile := files.FileDto{ID: 1}, files.FileDto{ID: 2}, files.FileDto{ID: 3}

	registry.recordFailure(newThumbnailFallbackKey(expiredFile, 320, 180, generatedImageKindThumbnail), expiredFile)
	clock.advance(thumbnailFallbackTTL)
	registry.recordFailure(newThumbnailFallbackKey(liveFile, 320, 180, generatedImageKindThumbnail), liveFile)
	registry.recordFailure(newThumbnailFallbackKey(newFile, 320, 180, generatedImageKindThumbnail), newFile)

	if !registry.isRecentlyFailed(newThumbnailFallbackKey(liveFile, 320, 180, generatedImageKindThumbnail), liveFile) {
		t.Fatalf("live entry should survive eviction of the expired one")
	}
}
