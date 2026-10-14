package video

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/applog"
)

const maxCachedSubtitleTrackLists = 512

var (
	ErrEmbeddedSubtitleNotFound = errors.New("embedded subtitle track not found")
	ErrEmbeddedSubtitleProbe    = errors.New("embedded subtitle probe failed")
	ErrEmbeddedSubtitleExtract  = errors.New("embedded subtitle extraction failed")
)

type EmbeddedSubtitleServiceInterface interface {
	ListEmbeddedSubtitles(videoFile files.FileDto) []SubtitleDto
	ReadEmbeddedSubtitleWebVtt(videoFile files.FileDto, streamIndex int) ([]byte, error)
}

type embeddedSubtitleVersion struct {
	fileID    int
	updatedAt int64
	size      int64
}

type EmbeddedSubtitleService struct {
	probeRunner   SubtitleProbeRunner
	extractRunner SubtitleExtractRunner
	timeout       func() time.Duration

	trackListsMutex sync.Mutex
	trackLists      map[embeddedSubtitleVersion][]EmbeddedSubtitleTrack

	probeFlights   singleflight.Group
	extractFlights singleflight.Group
}

func NewEmbeddedSubtitleService() *EmbeddedSubtitleService {
	return NewEmbeddedSubtitleServiceWithRunners(RunFFprobeSubtitleProbe, RunFFmpegSubtitleExtract)
}

func NewEmbeddedSubtitleServiceWithRunners(probeRunner SubtitleProbeRunner, extractRunner SubtitleExtractRunner) *EmbeddedSubtitleService {
	return &EmbeddedSubtitleService{
		probeRunner:   probeRunner,
		extractRunner: extractRunner,
		timeout:       ffmpegTimeout,
		trackLists:    map[embeddedSubtitleVersion][]EmbeddedSubtitleTrack{},
	}
}

func versionOfVideo(videoFile files.FileDto) embeddedSubtitleVersion {
	return embeddedSubtitleVersion{fileID: videoFile.ID, updatedAt: videoFile.UpdatedAt.UnixNano(), size: videoFile.Size}
}

func (s *EmbeddedSubtitleService) ListEmbeddedSubtitles(videoFile files.FileDto) []SubtitleDto {
	subtitles := []SubtitleDto{}
	tracks, err := s.loadTracks(videoFile)
	if err != nil {
		applog.Warn("video: embedded subtitle probe failed", "video_id", videoFile.ID, "error", err)
		return subtitles
	}
	for _, track := range tracks {
		subtitles = append(subtitles, BuildEmbeddedSubtitleDto(videoFile.ID, track))
	}
	return subtitles
}

func (s *EmbeddedSubtitleService) ReadEmbeddedSubtitleWebVtt(videoFile files.FileDto, streamIndex int) ([]byte, error) {
	tracks, err := s.loadTracks(videoFile)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEmbeddedSubtitleProbe, err)
	}
	if !containsStreamIndex(tracks, streamIndex) {
		return nil, ErrEmbeddedSubtitleNotFound
	}
	cachePath := embeddedSubtitleCachePath(videoFile, streamIndex)
	if cachedWebVtt, err := os.ReadFile(cachePath); err == nil {
		return cachedWebVtt, nil
	}
	flightKey := cachePath
	extractedWebVtt, err, _ := s.extractFlights.Do(flightKey, func() (any, error) {
		return s.extractToCache(videoFile, streamIndex, cachePath)
	})
	if err != nil {
		return nil, err
	}
	return extractedWebVtt.([]byte), nil
}

func containsStreamIndex(tracks []EmbeddedSubtitleTrack, streamIndex int) bool {
	for _, track := range tracks {
		if track.StreamIndex == streamIndex {
			return true
		}
	}
	return false
}

func embeddedSubtitleCachePath(videoFile files.FileDto, streamIndex int) string {
	cacheVersion := fmt.Sprintf("v%d-%d", videoFile.UpdatedAt.Unix(), videoFile.Size)
	return filepath.Join(videoThumbnailCacheDir(), fmt.Sprintf("%d_sub_%d_%s.vtt", videoFile.ID, streamIndex, cacheVersion))
}

func (s *EmbeddedSubtitleService) extractToCache(videoFile files.FileDto, streamIndex int, cachePath string) ([]byte, error) {
	if cachedWebVtt, err := os.ReadFile(cachePath); err == nil {
		return cachedWebVtt, nil
	}
	if err := os.MkdirAll(filepath.Dir(cachePath), 0755); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEmbeddedSubtitleExtract, err)
	}
	temporaryPath, err := createTemporaryFileBeside(cachePath, videoSubtitleTemporaryPattern(videoFile.ID, streamIndex))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEmbeddedSubtitleExtract, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout())
	defer cancel()
	arguments := buildEmbeddedSubtitleExtractArguments(videoFile.ResolveContentPath(), streamIndex, temporaryPath)
	if err := s.extractRunner(ctx, arguments); err != nil {
		_ = os.Remove(temporaryPath)
		return nil, fmt.Errorf("%w: %v", ErrEmbeddedSubtitleExtract, err)
	}
	extractedWebVtt, err := os.ReadFile(temporaryPath)
	if err != nil || len(extractedWebVtt) == 0 {
		_ = os.Remove(temporaryPath)
		return nil, fmt.Errorf("%w: empty or unreadable output: %v", ErrEmbeddedSubtitleExtract, err)
	}
	if err := publishTemporaryFile(temporaryPath, cachePath); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrEmbeddedSubtitleExtract, err)
	}
	return extractedWebVtt, nil
}

func videoSubtitleTemporaryPattern(videoID int, streamIndex int) string {
	return strconv.Itoa(videoID) + "_sub_" + strconv.Itoa(streamIndex) + "_tmp_*.vtt"
}

func (s *EmbeddedSubtitleService) loadTracks(videoFile files.FileDto) ([]EmbeddedSubtitleTrack, error) {
	version := versionOfVideo(videoFile)
	if tracks, isCached := s.cachedTracks(version); isCached {
		return tracks, nil
	}
	flightKey := fmt.Sprintf("%d:%d:%d", version.fileID, version.updatedAt, version.size)
	probedTracks, err, _ := s.probeFlights.Do(flightKey, func() (any, error) {
		if tracks, isCached := s.cachedTracks(version); isCached {
			return tracks, nil
		}
		tracks, err := s.probeTracks(videoFile)
		if err != nil {
			return nil, err
		}
		s.storeTracks(version, tracks)
		return tracks, nil
	})
	if err != nil {
		return nil, err
	}
	return probedTracks.([]EmbeddedSubtitleTrack), nil
}

func (s *EmbeddedSubtitleService) probeTracks(videoFile files.FileDto) ([]EmbeddedSubtitleTrack, error) {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout())
	defer cancel()
	probeJSON, err := s.probeRunner(ctx, buildEmbeddedSubtitleProbeArguments(videoFile.ResolveContentPath()))
	if err != nil {
		return nil, err
	}
	tracks, skippedTrackCount, err := ParseEmbeddedSubtitleTracks(probeJSON)
	if err != nil {
		return nil, err
	}
	if skippedTrackCount > 0 {
		applog.Debug("video: image-based embedded subtitle tracks skipped", "video_id", videoFile.ID, "skipped", skippedTrackCount)
	}
	return tracks, nil
}

func (s *EmbeddedSubtitleService) cachedTracks(version embeddedSubtitleVersion) ([]EmbeddedSubtitleTrack, bool) {
	s.trackListsMutex.Lock()
	defer s.trackListsMutex.Unlock()
	tracks, isCached := s.trackLists[version]
	return tracks, isCached
}

func (s *EmbeddedSubtitleService) storeTracks(version embeddedSubtitleVersion, tracks []EmbeddedSubtitleTrack) {
	s.trackListsMutex.Lock()
	defer s.trackListsMutex.Unlock()
	if len(s.trackLists) >= maxCachedSubtitleTrackLists {
		s.trackLists = map[embeddedSubtitleVersion][]EmbeddedSubtitleTrack{}
	}
	s.trackLists[version] = tracks
}
