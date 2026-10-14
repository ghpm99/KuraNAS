package video

import (
	"errors"
	"fmt"
	"image"
	"os"
	"path/filepath"
	"strconv"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/img"
)

var ErrVideoMasterFrameUnavailable = errors.New("video master frame unavailable")

func videoThumbnailCacheDir() string {
	return filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video")
}

func masterFramePath(fileID int) string {
	return filepath.Join(videoThumbnailCacheDir(), fmt.Sprintf("%d_master_%s.jpg", fileID, thumbnailFrameStrategyVersion))
}

func sourcePosterPath(fileID int) string {
	return filepath.Join(videoThumbnailCacheDir(), "source", strconv.Itoa(fileID))
}

func isFilePresent(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (s *Service) EnsureVideoMasterFrame(fileDto files.FileDto) error {
	if isFilePresent(sourcePosterPath(fileDto.ID)) || isFilePresent(masterFramePath(fileDto.ID)) {
		return nil
	}

	contentPath := fileDto.ResolveContentPath()
	if !fileExistsOnDisk(contentPath) {
		return fmt.Errorf("%w: %s", files.ErrFileMissingDisk, fileDto.Path)
	}

	failureKey := newThumbnailFallbackKey(fileDto, 0, 0, generatedImageKindMasterFrame)
	if s.FallbackRegistry.isRecentlyFailed(failureKey, fileDto) {
		return ErrVideoMasterFrameUnavailable
	}

	_, err, _ := s.masterFrameFlights.Do(strconv.Itoa(fileDto.ID), func() (any, error) {
		return nil, s.renderMasterFrame(fileDto, contentPath, failureKey)
	})
	return err
}

func (s *Service) renderMasterFrame(fileDto files.FileDto, contentPath string, failureKey thumbnailFallbackKey) error {
	finalPath := masterFramePath(fileDto.ID)
	if isFilePresent(finalPath) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return err
	}

	temporaryPath, err := createTemporaryFileBeside(finalPath, fmt.Sprintf("%d_master_tmp_*.jpg", fileDto.ID))
	if err != nil {
		return err
	}

	frameOffset := thumbnailFrameOffsetSeconds(s.loadVideoDuration(fileDto.ID))
	_, isRendered := s.renderFrameWithRetryAtStart(func(offsetSeconds float64) []string {
		return buildMasterFrameArguments(contentPath, offsetSeconds, temporaryPath)
	}, frameOffset, temporaryPath)
	if !isRendered {
		_ = os.Remove(temporaryPath)
		s.FallbackRegistry.recordFailure(failureKey, fileDto)
		return ErrVideoMasterFrameUnavailable
	}

	if err := publishTemporaryFile(temporaryPath, finalPath); err != nil {
		s.FallbackRegistry.recordFailure(failureKey, fileDto)
		return ErrVideoMasterFrameUnavailable
	}
	s.FallbackRegistry.clear(failureKey)
	return nil
}

func (s *Service) loadMasterFrameImage(fileDto files.FileDto) (image.Image, error) {
	if posterImage, _, err := img.OpenImageFromFile(sourcePosterPath(fileDto.ID)); err == nil {
		return posterImage, nil
	}

	if err := s.EnsureVideoMasterFrame(fileDto); err != nil {
		return nil, err
	}

	masterPath := masterFramePath(fileDto.ID)
	masterImage, _, err := img.OpenImageFromFile(masterPath)
	if err == nil {
		return masterImage, nil
	}

	_ = os.Remove(masterPath)
	s.FallbackRegistry.recordFailure(newThumbnailFallbackKey(fileDto, 0, 0, generatedImageKindMasterFrame), fileDto)
	return nil, ErrVideoMasterFrameUnavailable
}
