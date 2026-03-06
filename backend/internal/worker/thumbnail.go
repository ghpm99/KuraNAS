package worker

import (
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"time"
)

type ThumbnailStepInput struct {
	FileID int
	File   *files.FileDto
}

type ThumbnailStepOutput struct {
	Skipped bool
}

type ThumbnailStepExecutor struct {
	service files.ServiceInterface
}

func NewThumbnailStepExecutor(service files.ServiceInterface) *ThumbnailStepExecutor {
	return &ThumbnailStepExecutor{service: service}
}

func (e *ThumbnailStepExecutor) Execute(input ThumbnailStepInput) (ThumbnailStepOutput, error) {
	if e == nil || e.service == nil {
		return ThumbnailStepOutput{}, fmt.Errorf("thumbnail step: file service is required")
	}

	fileDto := files.FileDto{}
	if input.File != nil {
		fileDto = *input.File
		if fileDto.ID <= 0 {
			loadedFile, err := e.service.GetFileByNameAndPath(fileDto.Name, fileDto.Path)
			if err != nil {
				return ThumbnailStepOutput{}, fmt.Errorf("thumbnail step: resolve persisted file: %w", err)
			}
			fileDto = loadedFile
		}
	} else {
		if input.FileID <= 0 {
			return ThumbnailStepOutput{}, fmt.Errorf("thumbnail step: invalid file id")
		}

		loadedFile, err := e.service.GetFileById(input.FileID)
		if err != nil {
			return ThumbnailStepOutput{}, err
		}
		fileDto = loadedFile
	}

	if fileDto.Type != files.File {
		return ThumbnailStepOutput{Skipped: true}, newStepSkipped("thumbnail not applicable for directory")
	}

	isVideo := utils.GetFormatTypeByExtension(fileDto.Format).Type == utils.FormatTypeVideo
	if isThumbnailCacheUpToDate(fileDto, isVideo) {
		return ThumbnailStepOutput{Skipped: true}, newStepSkipped("thumbnail up-to-date")
	}

	if isVideo {
		if _, err := e.service.GetVideoThumbnail(fileDto, 320, 180); err != nil {
			return ThumbnailStepOutput{}, err
		}
		if _, err := e.service.GetVideoPreviewGif(fileDto, 320, 180); err != nil {
			return ThumbnailStepOutput{}, err
		}
		return ThumbnailStepOutput{}, nil
	}

	if _, err := e.service.GetFileThumbnail(fileDto, 320, 320); err != nil {
		return ThumbnailStepOutput{}, err
	}
	return ThumbnailStepOutput{}, nil
}

func isThumbnailCacheUpToDate(fileDto files.FileDto, isVideo bool) bool {
	cacheRoot := config.GetBuildConfig("ThumbnailPath")
	if cacheRoot == "" || fileDto.ID <= 0 {
		return false
	}

	paths := []string{}
	if isVideo {
		paths = append(paths,
			filepath.Join(cacheRoot, "video", fmt.Sprintf("%d_%dx%d.png", fileDto.ID, 320, 180)),
			filepath.Join(cacheRoot, "video", fmt.Sprintf("%d_%dx%d_preview.gif", fileDto.ID, 320, 180)),
		)
	} else {
		paths = append(paths, filepath.Join(cacheRoot, fmt.Sprintf("%d_%d.png", fileDto.ID, 320)))
	}

	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return false
		}
		if !isCachedAssetCurrent(info.ModTime(), fileDto.UpdatedAt) {
			return false
		}
	}

	return true
}

func isCachedAssetCurrent(cacheUpdatedAt time.Time, sourceUpdatedAt time.Time) bool {
	return cacheUpdatedAt.Equal(sourceUpdatedAt) || cacheUpdatedAt.After(sourceUpdatedAt)
}

func CreateThumbnailWorker(service files.ServiceInterface, data any, logService logger.LoggerServiceInterface) {
	_ = logService

	fileID, ok := data.(int)
	if !ok || fileID <= 0 {
		log.Println("CreateThumbnailWorker: data inválido, esperado fileID int")
		return
	}

	executor := NewThumbnailStepExecutor(service)
	_, err := executor.Execute(ThumbnailStepInput{FileID: fileID})
	if err != nil && !isStepSkipped(err) {
		log.Printf("CreateThumbnailWorker: erro ao gerar thumbnail padrao fileID=%d: %v\n", fileID, err)
	}
}
