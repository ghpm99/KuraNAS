package captures

import (
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Service struct {
	Repository          RepositoryInterface
	UploadJobDispatcher UploadJobDispatcherInterface
}

func NewService(repository RepositoryInterface, uploadJobDispatcher UploadJobDispatcherInterface) ServiceInterface {
	return &Service{
		Repository:          repository,
		UploadJobDispatcher: uploadJobDispatcher,
	}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	return database.ExecOptionalTx(s.Repository.GetDbContext(), fn)
}

func (s *Service) UploadCapture(file *multipart.FileHeader, dto CreateCaptureDto) (CaptureDto, error) {
	captureDir := buildCaptureDir(dto.Name)

	if err := os.MkdirAll(captureDir, 0755); err != nil {
		return CaptureDto{}, fmt.Errorf("UploadCapture: failed to create directory: %w", err)
	}

	fileName := sanitizeFileName(filepath.Base(file.Filename))
	destPath := filepath.Join(captureDir, fileName)

	if err := saveUploadedFile(file, destPath); err != nil {
		return CaptureDto{}, fmt.Errorf("UploadCapture: failed to save file: %w", err)
	}

	model := CaptureModel{
		Name:      dto.Name,
		FileName:  fileName,
		FilePath:  destPath,
		MediaType: dto.MediaType,
		MimeType:  dto.MimeType,
		Size:      file.Size,
		CreatedAt: time.Now(),
	}

	var result CaptureModel
	err := s.withTransaction(func(tx *sql.Tx) error {
		var createErr error
		result, createErr = s.Repository.CreateCapture(tx, model)
		return createErr
	})

	if err != nil {
		os.Remove(destPath)
		return CaptureDto{}, err
	}

	if s.UploadJobDispatcher != nil {
		_, jobErr := s.UploadJobDispatcher.CreateUploadProcessJob([]string{captureDir, destPath})
		if jobErr != nil {
			cleanupErr := s.withTransaction(func(tx *sql.Tx) error {
				return s.Repository.DeleteCapture(tx, result.ID)
			})
			_ = os.Remove(destPath)

			if cleanupErr != nil {
				return CaptureDto{}, fmt.Errorf("UploadCapture: failed to enqueue upload processing job: %w (cleanup failed: %v)", jobErr, cleanupErr)
			}

			return CaptureDto{}, fmt.Errorf("UploadCapture: failed to enqueue upload processing job: %w", jobErr)
		}
	}

	return result.ToDto(), nil
}

func (s *Service) GetCaptures(filter CaptureFilter, page int, pageSize int) (utils.PaginationResponse[CaptureDto], error) {
	pagination, err := s.Repository.GetCaptures(filter, page, pageSize)
	if err != nil {
		return utils.PaginationResponse[CaptureDto]{}, err
	}
	return ParsePaginationToDto(&pagination), nil
}

func (s *Service) GetCaptureByID(id int) (CaptureDto, error) {
	model, err := s.Repository.GetCaptureByID(id)
	if err != nil {
		return CaptureDto{}, err
	}
	return model.ToDto(), nil
}

func (s *Service) DeleteCapture(id int) error {
	model, err := s.Repository.GetCaptureByID(id)
	if err != nil {
		return fmt.Errorf("DeleteCapture: %w", err)
	}

	err = s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.DeleteCapture(tx, id)
	})

	if err != nil {
		return err
	}

	os.Remove(model.FilePath)
	return nil
}

func buildCaptureDir(name string) string {
	safeName := sanitizeFileName(name)
	return filepath.Join(config.AppConfig.EntryPoint, "capturas", safeName)
}

func sanitizeFileName(name string) string {
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	result := replacer.Replace(name)
	result = strings.TrimSpace(result)
	if result == "" {
		result = "unnamed"
	}
	return result
}

func saveUploadedFile(file *multipart.FileHeader, destPath string) error {
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
