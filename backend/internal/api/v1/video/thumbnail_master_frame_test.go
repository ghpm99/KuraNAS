package video

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	files "nas-go/api/internal/api/v1/files"
)

func newMasterFrameTestService(t *testing.T) (files.FileDto, *Service, *countingFrameRunner) {
	t.Helper()
	fileDto, service, _ := newFrameTestVideo(t)
	runner := &countingFrameRunner{}
	service.FrameCommandRunner = runner.run
	return fileDto, service, runner
}

func decodeThumbnailPNG(t *testing.T, encoded []byte) image.Image {
	t.Helper()
	decoded, err := png.Decode(bytes.NewReader(encoded))
	if err != nil {
		t.Fatalf("thumbnail is not a valid png: %v", err)
	}
	return decoded
}

func isOpaqueBlack(pixel color.Color) bool {
	red, green, blue, alpha := pixel.RGBA()
	return red == 0 && green == 0 && blue == 0 && alpha == 0xffff
}

func TestMasterFrameIsGeneratedOnceAndReusedAcrossSizes(t *testing.T) {
	fileDto, service, runner := newMasterFrameTestService(t)

	for _, size := range [][2]int{{240, 135}, {320, 180}, {640, 360}, {1280, 720}, {0, 0}} {
		thumbnail, err := service.GetVideoThumbnail(fileDto, size[0], size[1])
		if err != nil || thumbnail.IsFallback {
			t.Fatalf("size %v: fallback=%v err=%v", size, thumbnail.IsFallback, err)
		}
	}

	if runner.executions.Load() != 1 {
		t.Fatalf("expected a single ffmpeg execution, got %d", runner.executions.Load())
	}
}

func TestDerivedThumbnailHasExactSizeAndBlackPadding(t *testing.T) {
	fileDto, service, _ := newMasterFrameTestService(t)

	thumbnail, err := service.GetVideoThumbnail(fileDto, 320, 180)
	if err != nil {
		t.Fatal(err)
	}

	decoded := decodeThumbnailPNG(t, thumbnail.Data)
	if decoded.Bounds().Dx() != 320 || decoded.Bounds().Dy() != 180 {
		t.Fatalf("unexpected size %v", decoded.Bounds())
	}
	for _, corner := range []image.Point{{0, 0}, {319, 0}, {0, 179}, {319, 179}} {
		if !isOpaqueBlack(decoded.At(corner.X, corner.Y)) {
			t.Fatalf("corner %v is not opaque black: %v", corner, decoded.At(corner.X, corner.Y))
		}
	}
	if isOpaqueBlack(decoded.At(160, 90)) {
		t.Fatalf("center must hold video content")
	}
	contentLeft, contentRight := 160, 160
	for column := 0; column < 320; column++ {
		if isOpaqueBlack(decoded.At(column, 90)) {
			continue
		}
		contentLeft = min(contentLeft, column)
		contentRight = max(contentRight, column)
	}
	contentWidth := contentRight - contentLeft + 1
	if contentWidth < 238 || contentWidth > 242 {
		t.Fatalf("content must keep 4:3 inside 180px height (about 240 wide), got %d", contentWidth)
	}
}

func TestConcurrentRequestsForSameVideoRunFFmpegOnce(t *testing.T) {
	fileDto, service, runner := newMasterFrameTestService(t)
	runner.delay = 50 * time.Millisecond
	sizes := [][2]int{{240, 135}, {320, 180}, {400, 225}, {480, 270}, {640, 360}, {960, 540}, {1280, 720}}

	var group sync.WaitGroup
	for repetition := 0; repetition < 3; repetition++ {
		for _, size := range sizes {
			group.Add(1)
			go func(width, height int) {
				defer group.Done()
				thumbnail, err := service.GetVideoThumbnail(fileDto, width, height)
				if err != nil || thumbnail.IsFallback {
					t.Errorf("fallback=%v err=%v", thumbnail.IsFallback, err)
				}
			}(size[0], size[1])
		}
	}
	group.Wait()

	if runner.executions.Load() != 1 {
		t.Fatalf("expected a single ffmpeg execution, got %d", runner.executions.Load())
	}
}

func TestMasterFrameIsNotVisibleWhileFFmpegIsStillWriting(t *testing.T) {
	fileDto, service, _ := newMasterFrameTestService(t)
	wasMasterVisibleDuringRender := false
	service.FrameCommandRunner = func(ctx context.Context, arguments []string) error {
		wasMasterVisibleDuringRender = isFilePresent(masterFramePath(fileDto.ID))
		return writeFrameFixture(arguments[len(arguments)-1])
	}

	if err := service.EnsureVideoMasterFrame(fileDto); err != nil {
		t.Fatal(err)
	}

	if wasMasterVisibleDuringRender {
		t.Fatal("master must only appear after rename")
	}
	if !isFilePresent(masterFramePath(fileDto.ID)) {
		t.Fatal("master must exist after success")
	}
	assertNoTemporaryCacheFiles(t)
}

func assertNoTemporaryCacheFiles(t *testing.T) {
	t.Helper()
	temporaryFiles, _ := filepath.Glob(filepath.Join(videoThumbnailCacheDir(), "*tmp*"))
	if len(temporaryFiles) != 0 {
		t.Fatalf("temporary files left behind: %v", temporaryFiles)
	}
}

func TestFailedFFmpegLeavesNoPartialMasterAndFallsBackWithoutCache(t *testing.T) {
	fileDto, service, runner := newMasterFrameTestService(t)
	runner.shouldFail = true
	runner.leavePartial = true

	thumbnail, err := service.GetVideoThumbnail(fileDto, 320, 180)

	if err != nil || !thumbnail.IsFallback || len(thumbnail.Data) == 0 {
		t.Fatalf("expected fallback, fallback=%v err=%v", thumbnail.IsFallback, err)
	}
	cachedFiles, _ := filepath.Glob(filepath.Join(videoThumbnailCacheDir(), "901_*"))
	if len(cachedFiles) != 0 {
		t.Fatalf("nothing may remain in cache after failure: %v", cachedFiles)
	}
	assertNoTemporaryCacheFiles(t)
}

func TestMasterFailureIsSharedAcrossSizesWithoutRunningFFmpegAgain(t *testing.T) {
	fileDto, service, runner := newMasterFrameTestService(t)
	runner.shouldFail = true

	for _, size := range [][2]int{{320, 180}, {640, 360}, {240, 135}} {
		thumbnail, err := service.GetVideoThumbnail(fileDto, size[0], size[1])
		if err != nil || !thumbnail.IsFallback {
			t.Fatalf("size %v expected fallback err=%v", size, err)
		}
	}

	if runner.executions.Load() != 2 {
		t.Fatalf("expected only the first request to try (offset plus retry at zero), got %d executions", runner.executions.Load())
	}
}

func TestSourcePosterActsAsMasterWithoutRunningFFmpeg(t *testing.T) {
	fileDto, service, runner := newMasterFrameTestService(t)
	posterPath := sourcePosterPath(fileDto.ID)
	if err := os.MkdirAll(filepath.Dir(posterPath), 0755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Remove(posterPath) })
	posterFile, err := os.Create(posterPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := png.Encode(posterFile, buildFixtureMasterImage(80, 120)); err != nil {
		t.Fatal(err)
	}
	posterFile.Close()

	first, firstErr := service.GetVideoThumbnail(fileDto, 320, 180)
	second, secondErr := service.GetVideoThumbnail(fileDto, 640, 360)

	if firstErr != nil || secondErr != nil || first.IsFallback || second.IsFallback {
		t.Fatalf("unexpected errors %v %v", firstErr, secondErr)
	}
	if runner.executions.Load() != 0 {
		t.Fatalf("poster must replace ffmpeg, got %d executions", runner.executions.Load())
	}
	if isFilePresent(masterFramePath(fileDto.ID)) {
		t.Fatal("poster must not generate a master file")
	}
	if decoded := decodeThumbnailPNG(t, second.Data); decoded.Bounds().Dx() != 640 || decoded.Bounds().Dy() != 360 {
		t.Fatalf("unexpected size %v", decoded.Bounds())
	}
	if err := service.EnsureVideoMasterFrame(fileDto); err != nil || runner.executions.Load() != 0 {
		t.Fatalf("ensure must be a no-op with poster, err=%v", err)
	}
}

func TestEnsureVideoMasterFrameGeneratesOnceAndReportsFailure(t *testing.T) {
	fileDto, service, runner := newMasterFrameTestService(t)

	for attempt := 0; attempt < 3; attempt++ {
		if err := service.EnsureVideoMasterFrame(fileDto); err != nil {
			t.Fatal(err)
		}
	}
	if runner.executions.Load() != 1 {
		t.Fatalf("expected a single execution, got %d", runner.executions.Load())
	}

	removeCachedFilesOfVideo(fileDto.ID)
	runner.shouldFail = true
	if err := service.EnsureVideoMasterFrame(fileDto); !errors.Is(err, ErrVideoMasterFrameUnavailable) {
		t.Fatalf("expected unavailable error, got %v", err)
	}
}

func TestEnsureVideoMasterFrameReportsMissingFile(t *testing.T) {
	_, service, _ := newMasterFrameTestService(t)

	err := service.EnsureVideoMasterFrame(files.FileDto{ID: 902, Path: "/nonexistent/video.mp4", Type: files.File})

	if !errors.Is(err, files.ErrFileMissingDisk) {
		t.Fatalf("expected missing disk error, got %v", err)
	}
}

func TestCorruptedMasterIsDiscardedAndRegeneratedAfterFailureWindow(t *testing.T) {
	fileDto, service, runner := newMasterFrameTestService(t)
	if err := os.MkdirAll(videoThumbnailCacheDir(), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(masterFramePath(fileDto.ID), []byte("garbage"), 0644); err != nil {
		t.Fatal(err)
	}

	thumbnail, err := service.GetVideoThumbnail(fileDto, 320, 180)

	if err != nil || !thumbnail.IsFallback {
		t.Fatalf("expected fallback for corrupted master, err=%v", err)
	}
	if isFilePresent(masterFramePath(fileDto.ID)) {
		t.Fatal("corrupted master must be removed")
	}
	if runner.executions.Load() != 0 {
		t.Fatalf("unexpected ffmpeg execution")
	}
}
