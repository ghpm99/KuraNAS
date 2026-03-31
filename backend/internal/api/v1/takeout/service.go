package takeout

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"nas-go/api/internal/api/v1/jobs"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const takeoutUploadChunkSize int64 = 2 * 1024 * 1024

type notificationServiceInterface interface {
	GroupOrCreate(dto notifications.CreateNotificationDto) (notifications.NotificationDto, error)
}

type Service struct {
	UploadJobDispatcher UploadJobDispatcherInterface
	LibraryResolver     LibraryResolverInterface
	NotificationService notificationServiceInterface
}

func NewService(
	uploadJobDispatcher UploadJobDispatcherInterface,
	libraryResolver LibraryResolverInterface,
	notificationService notificationServiceInterface,
) ServiceInterface {
	return &Service{
		UploadJobDispatcher: uploadJobDispatcher,
		LibraryResolver:     libraryResolver,
		NotificationService: notificationService,
	}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	if s.UploadJobDispatcher == nil {
		return fmt.Errorf("upload job dispatcher is not configured")
	}
	return database.ExecOptionalTx(s.UploadJobDispatcher.GetDbContext(), fn)
}

func (s *Service) InitUpload(dto InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error) {
	fileName := sanitizeTakeoutFileName(dto.FileName)
	if fileName == "" || fileName == "unnamed" {
		return InitTakeoutUploadResultDto{}, fmt.Errorf("InitUpload: file_name is required")
	}

	uploadID, err := generateTakeoutUploadID()
	if err != nil {
		return InitTakeoutUploadResultDto{}, fmt.Errorf("InitUpload generate upload id: %w", err)
	}

	sessionDir := s.takeoutUploadSessionDir(uploadID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return InitTakeoutUploadResultDto{}, fmt.Errorf("InitUpload create session dir: %w", err)
	}

	session := TakeoutUploadSession{
		UploadID:      uploadID,
		FileName:      fileName,
		ExpectedSize:  dto.Size,
		ReceivedSize:  0,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
	}
	if err := s.saveTakeoutUploadSession(session); err != nil {
		return InitTakeoutUploadResultDto{}, err
	}

	s.emitImportStartedNotification(fileName, uploadID)
	return InitTakeoutUploadResultDto{
		UploadID:  uploadID,
		ChunkSize: takeoutUploadChunkSize,
	}, nil
}

func (s *Service) UploadChunk(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error {
	session, err := s.loadTakeoutUploadSession(dto.UploadID)
	if err != nil {
		return err
	}

	if dto.Offset != session.ReceivedSize {
		return ErrUploadOffsetMismatch
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("UploadChunk open chunk: %w", err)
	}
	defer src.Close()

	tempPath := s.takeoutUploadTempFilePath(dto.UploadID)
	dst, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("UploadChunk open temp file: %w", err)
	}
	defer dst.Close()

	if _, err := dst.Seek(session.ReceivedSize, io.SeekStart); err != nil {
		return fmt.Errorf("UploadChunk seek temp file: %w", err)
	}

	written, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("UploadChunk write chunk: %w", err)
	}

	session.ReceivedSize += written
	session.LastUpdatedAt = time.Now()
	return s.saveTakeoutUploadSession(session)
}

func (s *Service) CompleteUpload(dto CompleteTakeoutUploadDto) (TakeoutImportResultDto, error) {
	session, err := s.loadTakeoutUploadSession(dto.UploadID)
	if err != nil {
		return TakeoutImportResultDto{}, err
	}

	if session.ExpectedSize > 0 && session.ReceivedSize != session.ExpectedSize {
		return TakeoutImportResultDto{}, ErrUploadIncomplete
	}

	zipPath := s.takeoutUploadTempFilePath(dto.UploadID)
	if filepath.Ext(session.FileName) != ".zip" {
		return TakeoutImportResultDto{}, ErrInvalidZipFile
	}

	payload, err := json.Marshal(map[string]any{
		"zip_path":  zipPath,
		"upload_id": dto.UploadID,
		"file_name": session.FileName,
	})
	if err != nil {
		return TakeoutImportResultDto{}, fmt.Errorf("CompleteUpload marshal payload: %w", err)
	}

	var createdJob jobs.JobModel
	err = s.withTransaction(func(tx *sql.Tx) error {
		scope, scopeErr := json.Marshal(map[string]any{
			"upload_id": dto.UploadID,
			"file_name": session.FileName,
		})
		if scopeErr != nil {
			return scopeErr
		}

		jobModel, createJobErr := s.UploadJobDispatcher.CreateJob(tx, jobs.JobModel{
			Type:            "takeout_import",
			Priority:        "high",
			Scope:           scope,
			Status:          "queued",
			CancelRequested: false,
		})
		if createJobErr != nil {
			return createJobErr
		}
		createdJob = jobModel

		_, createStepErr := s.UploadJobDispatcher.CreateStep(tx, jobs.StepModel{
			JobID:       createdJob.ID,
			Type:        "takeout_extract",
			Status:      "queued",
			DependsOn:   []byte("[]"),
			Attempts:    0,
			MaxAttempts: 1,
			Progress:    0,
			Payload:     payload,
		})
		return createStepErr
	})
	if err != nil {
		return TakeoutImportResultDto{}, fmt.Errorf("CompleteUpload create job: %w", err)
	}

	return TakeoutImportResultDto{
		JobID:   createdJob.ID,
		Message: i18n.GetMessage("TAKEOUT_IMPORT_STARTED"),
	}, nil
}

func (s *Service) takeoutUploadRootDir() string {
	return filepath.Join(config.AppConfig.EntryPoint, ".takeout_uploads")
}

func (s *Service) takeoutUploadSessionDir(uploadID string) string {
	return filepath.Join(s.takeoutUploadRootDir(), uploadID)
}

func (s *Service) takeoutUploadSessionMetaPath(uploadID string) string {
	return filepath.Join(s.takeoutUploadSessionDir(uploadID), "session.json")
}

func (s *Service) takeoutUploadTempFilePath(uploadID string) string {
	return filepath.Join(s.takeoutUploadSessionDir(uploadID), "takeout.zip")
}

func (s *Service) saveTakeoutUploadSession(session TakeoutUploadSession) error {
	payload, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return fmt.Errorf("saveTakeoutUploadSession marshal session: %w", err)
	}

	if err := os.WriteFile(s.takeoutUploadSessionMetaPath(session.UploadID), payload, 0644); err != nil {
		return fmt.Errorf("saveTakeoutUploadSession write session: %w", err)
	}

	return nil
}

func (s *Service) loadTakeoutUploadSession(uploadID string) (TakeoutUploadSession, error) {
	payload, err := os.ReadFile(s.takeoutUploadSessionMetaPath(uploadID))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return TakeoutUploadSession{}, ErrUploadSessionNotFound
		}
		return TakeoutUploadSession{}, fmt.Errorf("loadTakeoutUploadSession read session: %w", err)
	}

	var session TakeoutUploadSession
	if err := json.Unmarshal(payload, &session); err != nil {
		return TakeoutUploadSession{}, fmt.Errorf("loadTakeoutUploadSession decode session: %w", err)
	}

	return session, nil
}

func (s *Service) emitImportStartedNotification(fileName string, uploadID string) {
	if s.NotificationService == nil {
		return
	}

	_, _ = s.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
		Type:     string(notifications.NotificationTypeInfo),
		Title:    i18n.GetMessage("NOTIFICATION_TAKEOUT_IMPORT_STARTED_TITLE"),
		Message:  i18n.Translate("NOTIFICATION_TAKEOUT_IMPORT_STARTED_MESSAGE", fileName),
		GroupKey: "takeout_import",
		Metadata: map[string]any{
			"event":     "takeout_import_started",
			"upload_id": uploadID,
			"file_name": fileName,
		},
	})
}

func generateTakeoutUploadID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func sanitizeTakeoutFileName(fileName string) string {
	name := strings.TrimSpace(fileName)
	if name == "" {
		return ""
	}
	return strings.Map(func(r rune) rune {
		switch r {
		case '\\', '/', ':', '*', '?', '"', '<', '>', '|':
			return '_'
		default:
			return r
		}
	}, name)
}
