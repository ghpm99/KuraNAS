package video

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	files "nas-go/api/internal/api/v1/files"
)

type subtitleToolsFake struct {
	probeOutput    []byte
	probeErr       error
	extractErr     error
	extractContent string
	extractDelay   time.Duration
	probeCalls     atomic.Int32
	extractCalls   atomic.Int32
	extractedArgs  []string
	argsMutex      sync.Mutex
}

func (fake *subtitleToolsFake) probe(ctx context.Context, arguments []string) ([]byte, error) {
	fake.probeCalls.Add(1)
	return fake.probeOutput, fake.probeErr
}

func (fake *subtitleToolsFake) extract(ctx context.Context, arguments []string) error {
	fake.extractCalls.Add(1)
	fake.argsMutex.Lock()
	fake.extractedArgs = arguments
	fake.argsMutex.Unlock()
	time.Sleep(fake.extractDelay)
	if fake.extractErr != nil {
		return fake.extractErr
	}
	outputPath := arguments[len(arguments)-1]
	return os.WriteFile(outputPath, []byte(fake.extractContent), 0o644)
}

func newSubtitleToolsFake(t *testing.T) *subtitleToolsFake {
	t.Helper()
	probeOutput, err := os.ReadFile("testdata/ffprobe_subtitle_streams.json")
	if err != nil {
		t.Fatal(err)
	}
	return &subtitleToolsFake{probeOutput: probeOutput, extractContent: "WEBVTT\n\n00:00:01.000 --> 00:00:02.000\nOla\n"}
}

func embeddedSubtitleVideoForTest(t *testing.T, videoID int) files.FileDto {
	t.Helper()
	setProgramFilesForVideoTest(t)
	t.Cleanup(func() {
		cachedFiles, _ := filepath.Glob(filepath.Join(videoThumbnailCacheDir(), strconv.Itoa(videoID)+"_sub_*"))
		for _, cachedFile := range cachedFiles {
			_ = os.Remove(cachedFile)
		}
	})
	return files.FileDto{ID: videoID, Name: "Filme.mkv", Path: "/data/Filme.mkv", Size: 1000, UpdatedAt: time.Unix(1700000000, 0)}
}

func TestListEmbeddedSubtitlesBuildsDtosAndCachesProbe(t *testing.T) {
	tools := newSubtitleToolsFake(t)
	service := NewEmbeddedSubtitleServiceWithRunners(tools.probe, tools.extract)
	video := files.FileDto{ID: 9101, Size: 10, UpdatedAt: time.Unix(1, 0)}

	firstList := service.ListEmbeddedSubtitles(video)
	secondList := service.ListEmbeddedSubtitles(video)

	if len(firstList) != 3 || firstList[0].Key != "9101:2" || firstList[1].IsForced != true || firstList[2].Language != "es" {
		t.Fatalf("unexpected list %+v", firstList)
	}
	if len(secondList) != 3 || tools.probeCalls.Load() != 1 {
		t.Fatalf("expected cached probe, calls=%d", tools.probeCalls.Load())
	}

	changedVideo := video
	changedVideo.Size = 11
	service.ListEmbeddedSubtitles(changedVideo)
	if tools.probeCalls.Load() != 2 {
		t.Fatalf("expected probe rerun after version change, calls=%d", tools.probeCalls.Load())
	}
}

func TestListEmbeddedSubtitlesReturnsEmptyListWhenProbeFailsAndDoesNotCacheFailure(t *testing.T) {
	tools := newSubtitleToolsFake(t)
	tools.probeErr = errors.New("ffprobe missing")
	service := NewEmbeddedSubtitleServiceWithRunners(tools.probe, tools.extract)
	video := files.FileDto{ID: 9102}

	subtitles := service.ListEmbeddedSubtitles(video)
	service.ListEmbeddedSubtitles(video)

	if subtitles == nil || len(subtitles) != 0 || tools.probeCalls.Load() != 2 {
		t.Fatalf("unexpected %+v calls=%d", subtitles, tools.probeCalls.Load())
	}
}

func TestListEmbeddedSubtitlesReturnsEmptyListOnInvalidProbeOutput(t *testing.T) {
	tools := newSubtitleToolsFake(t)
	tools.probeOutput = []byte("garbage")
	service := NewEmbeddedSubtitleServiceWithRunners(tools.probe, tools.extract)

	if subtitles := service.ListEmbeddedSubtitles(files.FileDto{ID: 9103}); len(subtitles) != 0 {
		t.Fatalf("unexpected %+v", subtitles)
	}
}

func TestReadEmbeddedSubtitleWebVttExtractsOnceThenServesDiskCache(t *testing.T) {
	tools := newSubtitleToolsFake(t)
	service := NewEmbeddedSubtitleServiceWithRunners(tools.probe, tools.extract)
	video := embeddedSubtitleVideoForTest(t, 9104)

	firstWebVtt, firstErr := service.ReadEmbeddedSubtitleWebVtt(video, 3)
	secondWebVtt, secondErr := service.ReadEmbeddedSubtitleWebVtt(video, 3)

	if firstErr != nil || secondErr != nil || string(firstWebVtt) != tools.extractContent || string(secondWebVtt) != tools.extractContent {
		t.Fatalf("unexpected %q %v / %q %v", firstWebVtt, firstErr, secondWebVtt, secondErr)
	}
	if tools.extractCalls.Load() != 1 {
		t.Fatalf("expected one extraction, got %d", tools.extractCalls.Load())
	}
	if !isFilePresent(embeddedSubtitleCachePath(video, 3)) {
		t.Fatal("expected disk cache file")
	}
	leftovers, _ := filepath.Glob(filepath.Join(videoThumbnailCacheDir(), "9104_sub_*_tmp_*"))
	if len(leftovers) != 0 {
		t.Fatalf("temporary files left: %v", leftovers)
	}
}

func TestReadEmbeddedSubtitleWebVttCoalescesConcurrentExtractions(t *testing.T) {
	tools := newSubtitleToolsFake(t)
	tools.extractDelay = 100 * time.Millisecond
	service := NewEmbeddedSubtitleServiceWithRunners(tools.probe, tools.extract)
	video := embeddedSubtitleVideoForTest(t, 9105)

	var group sync.WaitGroup
	for range 8 {
		group.Add(1)
		go func() {
			defer group.Done()
			webVtt, err := service.ReadEmbeddedSubtitleWebVtt(video, 2)
			if err != nil || string(webVtt) != tools.extractContent {
				t.Errorf("unexpected %q %v", webVtt, err)
			}
		}()
	}
	group.Wait()

	if tools.extractCalls.Load() != 1 {
		t.Fatalf("expected coalesced extraction, got %d", tools.extractCalls.Load())
	}
}

func TestReadEmbeddedSubtitleWebVttRejectsUnknownAndImageTracks(t *testing.T) {
	tools := newSubtitleToolsFake(t)
	service := NewEmbeddedSubtitleServiceWithRunners(tools.probe, tools.extract)
	video := embeddedSubtitleVideoForTest(t, 9106)

	for _, streamIndex := range []int{4, 6, 99, -1} {
		if _, err := service.ReadEmbeddedSubtitleWebVtt(video, streamIndex); !errors.Is(err, ErrEmbeddedSubtitleNotFound) {
			t.Fatalf("stream %d expected not found, got %v", streamIndex, err)
		}
	}
	if tools.extractCalls.Load() != 0 {
		t.Fatal("extraction must not run for rejected tracks")
	}
}

func TestReadEmbeddedSubtitleWebVttReportsProbeFailure(t *testing.T) {
	tools := newSubtitleToolsFake(t)
	tools.probeErr = errors.New("boom")
	service := NewEmbeddedSubtitleServiceWithRunners(tools.probe, tools.extract)

	if _, err := service.ReadEmbeddedSubtitleWebVtt(files.FileDto{ID: 9107}, 2); !errors.Is(err, ErrEmbeddedSubtitleProbe) {
		t.Fatalf("expected probe error, got %v", err)
	}
}

func TestReadEmbeddedSubtitleWebVttFailureLeavesNoCache(t *testing.T) {
	testCases := map[string]func(tools *subtitleToolsFake){
		"ffmpeg error": func(tools *subtitleToolsFake) { tools.extractErr = errors.New("exit status 1") },
		"empty output": func(tools *subtitleToolsFake) { tools.extractContent = "" },
	}
	for name, breakTools := range testCases {
		t.Run(name, func(t *testing.T) {
			tools := newSubtitleToolsFake(t)
			breakTools(tools)
			service := NewEmbeddedSubtitleServiceWithRunners(tools.probe, tools.extract)
			video := embeddedSubtitleVideoForTest(t, 9108)

			_, err := service.ReadEmbeddedSubtitleWebVtt(video, 2)

			if !errors.Is(err, ErrEmbeddedSubtitleExtract) {
				t.Fatalf("expected extract error, got %v", err)
			}
			leftovers, _ := filepath.Glob(filepath.Join(videoThumbnailCacheDir(), "9108_sub_*"))
			if len(leftovers) != 0 {
				t.Fatalf("expected no cache or temporary files, got %v", leftovers)
			}
		})
	}
}
