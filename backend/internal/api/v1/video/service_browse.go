package video

import (
	"bytes"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
	"path/filepath"
	"time"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/icons"
	"nas-go/api/pkg/img"
	"nas-go/api/pkg/utils"
)

func (s *Service) GetVideos(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	filesModel, err := s.Repository.GetVideos(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}
	return files.ParsePaginationToDto(&filesModel)
}

func ffmpegTimeout() time.Duration {
	return config.StepTimeout()
}

func fileExistsOnDisk(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func (s *Service) GetVideoThumbnail(fileDto files.FileDto, width, height int) (GeneratedImage, error) {
	if width <= 0 {
		width = 320
	}
	if height <= 0 {
		height = 180
	}
	if width > 2048 {
		width = 2048
	}
	if height > 2048 {
		height = 2048
	}

	cacheDir := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video")
	_ = os.MkdirAll(cacheDir, 0755)
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%d_%dx%d_%s.png", fileDto.ID, width, height, thumbnailFrameStrategyVersion))

	if data, err := os.ReadFile(cachePath); err == nil {
		return GeneratedImage{Data: data}, nil
	}

	sourcePoster := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video", "source", fmt.Sprintf("%d", fileDto.ID))
	if posterImg, _, posterErr := img.OpenImageFromFile(sourcePoster); posterErr == nil {
		thumb := img.Thumbnail(posterImg, uint(width), uint(height))
		if encoded, encErr := img.EncodePNG(thumb); encErr == nil {
			_ = os.WriteFile(cachePath, encoded, 0644)
			return GeneratedImage{Data: encoded}, nil
		}
	}

	contentPath := fileDto.ResolveContentPath()
	if !fileExistsOnDisk(contentPath) {
		return GeneratedImage{}, fmt.Errorf("%w: %s", files.ErrFileMissingDisk, fileDto.Path)
	}

	failureKey := newThumbnailFallbackKey(fileDto, width, height, generatedImageKindThumbnail)
	if s.FallbackRegistry.isRecentlyFailed(failureKey, fileDto) {
		return buildThumbnailFallback(width, height)
	}

	frameOffset := thumbnailFrameOffsetSeconds(s.loadVideoDuration(fileDto.ID))
	if renderedFrame, isRendered := s.renderFrameWithRetryAtStart(func(offsetSeconds float64) []string {
		return buildThumbnailFrameArguments(contentPath, offsetSeconds, width, height, cachePath)
	}, frameOffset, cachePath); isRendered {
		s.FallbackRegistry.clear(failureKey)
		return GeneratedImage{Data: renderedFrame}, nil
	}

	s.FallbackRegistry.recordFailure(failureKey, fileDto)
	_ = os.Remove(cachePath)
	return buildThumbnailFallback(width, height)
}

func buildThumbnailFallback(width, height int) (GeneratedImage, error) {
	iconImg, _ := icons.Mp4Icon()
	thumb := img.Thumbnail(iconImg, uint(width), uint(height))
	fallback, err := img.EncodePNG(thumb)
	if err != nil {
		return GeneratedImage{}, err
	}
	return GeneratedImage{Data: fallback, IsFallback: true}, nil
}

func (s *Service) GetVideoPreviewGif(fileDto files.FileDto, width, height int) (GeneratedImage, error) {
	if width <= 0 {
		width = 320
	}
	if height <= 0 {
		height = 180
	}
	if width > 1024 {
		width = 1024
	}
	if height > 1024 {
		height = 1024
	}

	cacheDir := filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video")
	_ = os.MkdirAll(cacheDir, 0755)
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%d_%dx%d_preview_%s.gif", fileDto.ID, width, height, thumbnailFrameStrategyVersion))

	if data, err := os.ReadFile(cachePath); err == nil {
		return GeneratedImage{Data: data}, nil
	}

	contentPath := fileDto.ResolveContentPath()
	if !fileExistsOnDisk(contentPath) {
		return GeneratedImage{}, fmt.Errorf("%w: %s", files.ErrFileMissingDisk, fileDto.Path)
	}

	failureKey := newThumbnailFallbackKey(fileDto, width, height, generatedImageKindPreview)
	if s.FallbackRegistry.isRecentlyFailed(failureKey, fileDto) {
		return buildPreviewFallback(width, height)
	}

	frameOffset := previewFrameOffsetSeconds(s.loadVideoDuration(fileDto.ID))
	if renderedPreview, isRendered := s.renderFrameWithRetryAtStart(func(offsetSeconds float64) []string {
		return buildPreviewGifArguments(contentPath, offsetSeconds, width, height, cachePath)
	}, frameOffset, cachePath); isRendered {
		s.FallbackRegistry.clear(failureKey)
		return GeneratedImage{Data: renderedPreview}, nil
	}

	s.FallbackRegistry.recordFailure(failureKey, fileDto)
	_ = os.Remove(cachePath)
	return buildPreviewFallback(width, height)
}

func buildPreviewFallback(width, height int) (GeneratedImage, error) {
	iconImg, _ := icons.Mp4Icon()
	thumb := img.Thumbnail(iconImg, uint(width), uint(height))

	paletted := image.NewPaletted(thumb.Bounds(), palette.Plan9)
	draw.FloydSteinberg.Draw(paletted, thumb.Bounds(), thumb, image.Point{})

	fallbackGif := &gif.GIF{
		Image:     []*image.Paletted{paletted},
		Delay:     []int{120},
		LoopCount: 0,
	}
	var buffer bytes.Buffer
	if err := gif.EncodeAll(&buffer, fallbackGif); err != nil {
		return GeneratedImage{}, err
	}
	return GeneratedImage{Data: buffer.Bytes(), IsFallback: true}, nil
}
