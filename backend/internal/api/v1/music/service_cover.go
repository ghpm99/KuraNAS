package music

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/img"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"time"
)

const (
	defaultCoverSize      = 256
	minimumCoverSize      = 32
	maximumCoverSize      = 1024
	albumCoverProbeLimit  = 50
	noCoverMarkerSuffix   = "none"
	renderedCoverSuffix   = "jpg"
	noCoverMarkerContent  = "no-cover"
	noCoverMarkerLifetime = 10 * time.Minute
)

var ErrCoverNotFound = errors.New("cover not found")

type CoverTrackSource interface {
	GetFileById(id int) (files.FileDto, error)
}

type CoverService struct {
	trackSource CoverTrackSource
	albumTracks CoverAlbumTracks
	cacheDir    string
}

type CoverAlbumTracks interface {
	GetLibraryTrackIDsByAlbum(albumKey string, page int, pageSize int) (utils.PaginationResponse[int], error)
}

type Cover struct {
	Data []byte
	ETag string
}

func NewCoverService(trackSource CoverTrackSource, albumTracks CoverAlbumTracks, cacheDir string) *CoverService {
	return &CoverService{trackSource: trackSource, albumTracks: albumTracks, cacheDir: cacheDir}
}

func NormalizeCoverSize(requestedSize int) int {
	if requestedSize <= 0 {
		return defaultCoverSize
	}
	return min(max(requestedSize, minimumCoverSize), maximumCoverSize)
}

func (service *CoverService) GetTrackCover(fileID int, size int) (Cover, error) {
	trackFile, err := service.trackSource.GetFileById(fileID)
	if err != nil {
		return Cover{}, err
	}
	return service.renderTrackCover(trackFile, NormalizeCoverSize(size))
}

func (service *CoverService) GetAlbumCover(albumKey string, size int) (Cover, error) {
	albumTrackPage, err := service.albumTracks.GetLibraryTrackIDsByAlbum(albumKey, 1, albumCoverProbeLimit)
	if err != nil {
		return Cover{}, err
	}
	trackIDs := albumTrackPage.Items
	normalizedSize := NormalizeCoverSize(size)
	for _, trackID := range trackIDs {
		trackFile, err := service.trackSource.GetFileById(trackID)
		if err != nil {
			continue
		}
		cover, err := service.renderTrackCover(trackFile, normalizedSize)
		if err == nil {
			return cover, nil
		}
	}
	return Cover{}, ErrCoverNotFound
}

func (service *CoverService) renderTrackCover(trackFile files.FileDto, size int) (Cover, error) {
	coverETag := coverETagFor(trackFile, size)
	renderedPath := service.cacheFilePath(trackFile, size, renderedCoverSuffix)
	if cachedData, err := os.ReadFile(renderedPath); err == nil {
		return Cover{Data: cachedData, ETag: coverETag}, nil
	}
	if service.hasFreshNoCoverMarker(trackFile) {
		return Cover{}, ErrCoverNotFound
	}

	sourceBytes, isFound := resolveCoverSource(trackFile.ResolveContentPath())
	if !isFound {
		service.writeCacheFile(service.cacheFilePath(trackFile, 0, noCoverMarkerSuffix), []byte(noCoverMarkerContent))
		return Cover{}, ErrCoverNotFound
	}
	renderedData, err := resizeCoverToJPEG(sourceBytes, size)
	if err != nil {
		service.writeCacheFile(service.cacheFilePath(trackFile, 0, noCoverMarkerSuffix), []byte(noCoverMarkerContent))
		return Cover{}, ErrCoverNotFound
	}
	service.writeCacheFile(renderedPath, renderedData)
	return Cover{Data: renderedData, ETag: coverETag}, nil
}

func (service *CoverService) hasFreshNoCoverMarker(trackFile files.FileDto) bool {
	markerInfo, err := os.Stat(service.cacheFilePath(trackFile, 0, noCoverMarkerSuffix))
	if err != nil {
		return false
	}
	return time.Since(markerInfo.ModTime()) < noCoverMarkerLifetime
}

func resolveCoverSource(audioPath string) ([]byte, bool) {
	if embeddedBytes, isEmbedded := extractEmbeddedCover(audioPath); isEmbedded {
		if _, _, err := image.DecodeConfig(bytes.NewReader(embeddedBytes)); err == nil {
			return embeddedBytes, true
		}
	}
	return findFolderCover(audioPath)
}

func resizeCoverToJPEG(sourceBytes []byte, size int) ([]byte, error) {
	decoded, _, err := image.Decode(bytes.NewReader(sourceBytes))
	if err != nil {
		return nil, err
	}
	return img.EncodeJPEG(img.FitWithinBox(decoded, size, size, 1))
}

func coverETagFor(trackFile files.FileDto, size int) string {
	return fmt.Sprintf(`"cover-%d-%d-%d"`, trackFile.ID, size, trackFile.UpdatedAt.UnixNano())
}

func (service *CoverService) cacheFilePath(trackFile files.FileDto, size int, suffix string) string {
	return filepath.Join(service.cacheDir, fmt.Sprintf("cover_%d_%d_%d.%s", trackFile.ID, size, trackFile.UpdatedAt.UnixNano(), suffix))
}

func (service *CoverService) writeCacheFile(path string, content []byte) {
	if err := os.MkdirAll(service.cacheDir, 0755); err != nil {
		return
	}
	_ = os.WriteFile(path, content, 0644)
}
