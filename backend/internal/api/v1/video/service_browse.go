package video

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color/palette"
	"image/draw"
	"image/gif"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/icons"
	"nas-go/api/pkg/img"
	"nas-go/api/pkg/utils"
)

// Browse/streaming support moved from the files core: listing videos and
// generating video thumbnails/previews are video-specific behavior.

func (s *Service) GetVideos(page int, pageSize int) (utils.PaginationResponse[files.FileDto], error) {
	filesModel, err := s.Repository.GetVideos(page, pageSize)
	if err != nil {
		return utils.PaginationResponse[files.FileDto]{}, err
	}
	return files.ParsePaginationToDto(&filesModel)
}

// ffmpegTimeout bounds ffmpeg invocations so a corrupt or stalled media file
// cannot hang the request/worker indefinitely; on timeout the caller falls back
// to a placeholder icon.
func ffmpegTimeout() time.Duration {
	return config.StepTimeout()
}

func fileExistsOnDisk(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
}

func (s *Service) GetVideoThumbnail(fileDto files.FileDto, width, height int) ([]byte, error) {
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
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%d_%dx%d.png", fileDto.ID, width, height))

	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	if !fileExistsOnDisk(fileDto.Path) {
		return nil, fmt.Errorf("%w: %s", files.ErrFileMissingDisk, fileDto.Path)
	}

	ffmpegCtx, ffmpegCancel := context.WithTimeout(context.Background(), ffmpegTimeout())
	defer ffmpegCancel()
	ffmpegErr := exec.CommandContext(
		ffmpegCtx,
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-ss", "00:00:03",
		"-i", fileDto.Path,
		"-frames:v", "1",
		"-vf", fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black", width, height, width, height),
		cachePath,
	).Run()

	if ffmpegErr == nil {
		if data, err := os.ReadFile(cachePath); err == nil {
			return data, nil
		}
	}

	iconImg, _ := icons.Mp4Icon()
	thumb := img.Thumbnail(iconImg, uint(width), uint(height))
	fallback, err := img.EncodePNG(thumb)
	if err != nil {
		return nil, err
	}
	_ = os.WriteFile(cachePath, fallback, 0644)
	return fallback, nil
}

func (s *Service) GetVideoPreviewGif(fileDto files.FileDto, width, height int) ([]byte, error) {
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
	cachePath := filepath.Join(cacheDir, fmt.Sprintf("%d_%dx%d_preview.gif", fileDto.ID, width, height))

	if data, err := os.ReadFile(cachePath); err == nil {
		return data, nil
	}

	if !fileExistsOnDisk(fileDto.Path) {
		return nil, fmt.Errorf("%w: %s", files.ErrFileMissingDisk, fileDto.Path)
	}

	// Curta prévia animada: ~2.5s, baixa taxa de frames para performance de cache e rede local.
	ffmpegCtx, ffmpegCancel := context.WithTimeout(context.Background(), ffmpegTimeout())
	defer ffmpegCancel()
	ffmpegErr := exec.CommandContext(
		ffmpegCtx,
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-ss", "00:00:03",
		"-t", "2.5",
		"-i", fileDto.Path,
		"-vf", fmt.Sprintf("fps=4,scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black", width, height, width, height),
		"-loop", "0",
		cachePath,
	).Run()

	if ffmpegErr == nil {
		if data, err := os.ReadFile(cachePath); err == nil {
			return data, nil
		}
	}

	iconImg, _ := icons.Mp4Icon()
	thumb := img.Thumbnail(iconImg, uint(width), uint(height))

	paletted := image.NewPaletted(thumb.Bounds(), palette.Plan9)
	draw.FloydSteinberg.Draw(paletted, thumb.Bounds(), thumb, image.Point{})

	g := &gif.GIF{
		Image:     []*image.Paletted{paletted},
		Delay:     []int{120},
		LoopCount: 0,
	}
	var buf bytes.Buffer
	if err := gif.EncodeAll(&buf, g); err != nil {
		return nil, err
	}
	fallback := buf.Bytes()
	_ = os.WriteFile(cachePath, fallback, 0644)
	return fallback, nil
}
