package captures

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const captureUploadChunkSize int64 = 2 * 1024 * 1024

type Service struct {
	Repository          RepositoryInterface
	UploadJobDispatcher UploadJobDispatcherInterface
	NotificationService notificationServiceInterface
}

type notificationServiceInterface interface {
	GroupOrCreate(dto notifications.CreateNotificationDto) (notifications.NotificationDto, error)
}

type captureUploadSession struct {
	UploadID      string    `json:"upload_id"`
	Name          string    `json:"name"`
	MediaType     string    `json:"media_type"`
	MimeType      string    `json:"mime_type"`
	FileName      string    `json:"file_name"`
	ExpectedSize  int64     `json:"expected_size"`
	ReceivedSize  int64     `json:"received_size"`
	CreatedAt     time.Time `json:"created_at"`
	LastUpdatedAt time.Time `json:"last_updated_at"`
}

func NewService(
	repository RepositoryInterface,
	uploadJobDispatcher UploadJobDispatcherInterface,
	notificationService notificationServiceInterface,
) ServiceInterface {
	return &Service{
		Repository:          repository,
		UploadJobDispatcher: uploadJobDispatcher,
		NotificationService: notificationService,
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

	result, err := s.persistCaptureAndDispatch(captureDir, destPath, fileName, dto, file.Size)
	if err != nil {
		_ = os.Remove(destPath)
		return CaptureDto{}, err
	}

	return result, nil
}

func (s *Service) InitCaptureUpload(dto InitCaptureUploadDto) (InitCaptureUploadResultDto, error) {
	if strings.TrimSpace(dto.Name) == "" {
		return InitCaptureUploadResultDto{}, fmt.Errorf("InitCaptureUpload: name is required")
	}

	uploadID, err := generateUploadID()
	if err != nil {
		return InitCaptureUploadResultDto{}, fmt.Errorf("InitCaptureUpload: generate upload id: %w", err)
	}

	sessionDir := s.captureUploadSessionDir(uploadID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return InitCaptureUploadResultDto{}, fmt.Errorf("InitCaptureUpload: create session dir: %w", err)
	}

	fileName := sanitizeFileName(filepath.Base(strings.TrimSpace(dto.FileName)))
	if fileName == "unnamed" {
		fileName = sanitizeFileName(dto.Name)
	}

	session := captureUploadSession{
		UploadID:      uploadID,
		Name:          dto.Name,
		MediaType:     dto.MediaType,
		MimeType:      dto.MimeType,
		FileName:      fileName,
		ExpectedSize:  dto.Size,
		ReceivedSize:  0,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Now(),
	}

	if err := s.saveCaptureUploadSession(session); err != nil {
		return InitCaptureUploadResultDto{}, err
	}

	s.emitUploadStartedNotification(session)

	return InitCaptureUploadResultDto{
		UploadID:  uploadID,
		ChunkSize: captureUploadChunkSize,
	}, nil
}

func (s *Service) UploadCaptureChunk(file *multipart.FileHeader, dto UploadCaptureChunkDto) error {
	session, err := s.loadCaptureUploadSession(dto.UploadID)
	if err != nil {
		return err
	}

	if dto.Offset != session.ReceivedSize {
		return fmt.Errorf("UploadCaptureChunk: offset mismatch, expected %d got %d", session.ReceivedSize, dto.Offset)
	}

	src, err := file.Open()
	if err != nil {
		return fmt.Errorf("UploadCaptureChunk: open chunk: %w", err)
	}
	defer src.Close()

	tempPath := s.captureUploadTempFilePath(dto.UploadID)
	dst, err := os.OpenFile(tempPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("UploadCaptureChunk: open temp file: %w", err)
	}
	defer dst.Close()

	if _, err := dst.Seek(session.ReceivedSize, io.SeekStart); err != nil {
		return fmt.Errorf("UploadCaptureChunk: seek temp file: %w", err)
	}

	written, err := io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("UploadCaptureChunk: write chunk: %w", err)
	}

	session.ReceivedSize += written
	session.LastUpdatedAt = time.Now()

	return s.saveCaptureUploadSession(session)
}

func (s *Service) CompleteCaptureUpload(dto CompleteCaptureUploadDto) (CaptureDto, error) {
	session, err := s.loadCaptureUploadSession(dto.UploadID)
	if err != nil {
		s.emitUploadFailedNotification(dto.UploadID, "", err)
		return CaptureDto{}, err
	}

	if session.ExpectedSize > 0 && session.ReceivedSize != session.ExpectedSize {
		uploadErr := fmt.Errorf(
			"CompleteCaptureUpload: incomplete upload, expected %d got %d",
			session.ExpectedSize,
			session.ReceivedSize,
		)
		s.emitUploadFailedNotification(dto.UploadID, session.Name, uploadErr)
		return CaptureDto{}, uploadErr
	}

	captureDir := buildCaptureDir(session.Name)
	if err := os.MkdirAll(captureDir, 0755); err != nil {
		uploadErr := fmt.Errorf("CompleteCaptureUpload: create directory: %w", err)
		s.emitUploadFailedNotification(dto.UploadID, session.Name, uploadErr)
		return CaptureDto{}, uploadErr
	}

	tempPath := s.captureUploadTempFilePath(dto.UploadID)
	destPath := filepath.Join(captureDir, sanitizeFileName(filepath.Base(session.FileName)))
	if err := moveFile(tempPath, destPath); err != nil {
		uploadErr := fmt.Errorf("CompleteCaptureUpload: finalize upload file: %w", err)
		s.emitUploadFailedNotification(dto.UploadID, session.Name, uploadErr)
		return CaptureDto{}, uploadErr
	}

	createDto := CreateCaptureDto{
		Name:      session.Name,
		MediaType: session.MediaType,
		MimeType:  session.MimeType,
		Size:      session.ReceivedSize,
	}

	capture, persistErr := s.persistCaptureAndDispatch(captureDir, destPath, filepath.Base(destPath), createDto, session.ReceivedSize)
	if persistErr != nil {
		_ = os.Remove(destPath)
		s.emitUploadFailedNotification(dto.UploadID, session.Name, persistErr)
		return CaptureDto{}, persistErr
	}

	_ = os.RemoveAll(s.captureUploadSessionDir(dto.UploadID))
	s.emitUploadCompletedNotification(dto.UploadID, session, capture)
	return capture, nil
}

func (s *Service) emitUploadStartedNotification(session captureUploadSession) {
	if s.NotificationService == nil {
		return
	}

	_, _ = s.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
		Type:     string(notifications.NotificationTypeInfo),
		Title:    i18n.GetMessage("NOTIFICATION_CAPTURE_UPLOAD_STARTED_TITLE"),
		Message:  i18n.Translate("NOTIFICATION_CAPTURE_UPLOAD_STARTED_MESSAGE", session.Name),
		GroupKey: "capture_upload_progress",
		Metadata: map[string]any{
			"event":         "capture_upload_started",
			"upload_id":     session.UploadID,
			"capture_name":  session.Name,
			"media_type":    session.MediaType,
			"mime_type":     session.MimeType,
			"expected_size": session.ExpectedSize,
		},
	})
}

func (s *Service) emitUploadCompletedNotification(uploadID string, session captureUploadSession, capture CaptureDto) {
	if s.NotificationService == nil {
		return
	}

	_, _ = s.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
		Type:     string(notifications.NotificationTypeSuccess),
		Title:    i18n.GetMessage("NOTIFICATION_CAPTURE_UPLOAD_COMPLETED_TITLE"),
		Message:  i18n.Translate("NOTIFICATION_CAPTURE_UPLOAD_COMPLETED_MESSAGE", session.Name),
		GroupKey: "capture_upload_result",
		Metadata: map[string]any{
			"event":         "capture_upload_completed",
			"upload_id":     uploadID,
			"capture_id":    capture.ID,
			"capture_name":  capture.Name,
			"file_name":     capture.FileName,
			"media_type":    capture.MediaType,
			"mime_type":     capture.MimeType,
			"received_size": session.ReceivedSize,
		},
	})
}

func (s *Service) emitUploadFailedNotification(uploadID string, captureName string, uploadErr error) {
	if s.NotificationService == nil {
		return
	}

	messageName := captureName
	if strings.TrimSpace(messageName) == "" {
		messageName = uploadID
	}

	_, _ = s.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
		Type:     string(notifications.NotificationTypeError),
		Title:    i18n.GetMessage("NOTIFICATION_CAPTURE_UPLOAD_FAILED_TITLE"),
		Message:  i18n.Translate("NOTIFICATION_CAPTURE_UPLOAD_FAILED_MESSAGE", messageName),
		GroupKey: "",
		Metadata: map[string]any{
			"event":        "capture_upload_failed",
			"upload_id":    uploadID,
			"capture_name": captureName,
			"error":        uploadErr.Error(),
		},
	})
}

func (s *Service) persistCaptureAndDispatch(
	captureDir string,
	destPath string,
	fileName string,
	dto CreateCaptureDto,
	size int64,
) (CaptureDto, error) {
	model := CaptureModel{
		Name:      dto.Name,
		FileName:  fileName,
		FilePath:  destPath,
		MediaType: dto.MediaType,
		MimeType:  dto.MimeType,
		Size:      size,
		CreatedAt: time.Now(),
	}

	var result CaptureModel
	err := s.withTransaction(func(tx *sql.Tx) error {
		var createErr error
		result, createErr = s.Repository.CreateCapture(tx, model)
		return createErr
	})

	if err != nil {
		return CaptureDto{}, err
	}

	if s.UploadJobDispatcher != nil {
		_, jobErr := s.UploadJobDispatcher.CreateUploadProcessJob([]string{captureDir, destPath})
		if jobErr != nil {
			cleanupErr := s.withTransaction(func(tx *sql.Tx) error {
				return s.Repository.DeleteCapture(tx, result.ID)
			})
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

func moveFile(source string, destination string) error {
	if err := os.Rename(source, destination); err == nil {
		return nil
	}

	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return os.Remove(source)
}

func generateUploadID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Service) captureUploadRootDir() string {
	return filepath.Join(config.AppConfig.EntryPoint, "capturas", ".uploads")
}

func (s *Service) captureUploadSessionDir(uploadID string) string {
	return filepath.Join(s.captureUploadRootDir(), uploadID)
}

func (s *Service) captureUploadSessionMetaPath(uploadID string) string {
	return filepath.Join(s.captureUploadSessionDir(uploadID), "meta.json")
}

func (s *Service) captureUploadTempFilePath(uploadID string) string {
	return filepath.Join(s.captureUploadSessionDir(uploadID), "payload.bin")
}

func (s *Service) saveCaptureUploadSession(session captureUploadSession) error {
	if session.UploadID == "" {
		return fmt.Errorf("saveCaptureUploadSession: upload id is required")
	}

	sessionDir := s.captureUploadSessionDir(session.UploadID)
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return fmt.Errorf("saveCaptureUploadSession: create session dir: %w", err)
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("saveCaptureUploadSession: marshal session: %w", err)
	}

	if err := os.WriteFile(s.captureUploadSessionMetaPath(session.UploadID), data, 0644); err != nil {
		return fmt.Errorf("saveCaptureUploadSession: write session meta: %w", err)
	}

	return nil
}

func (s *Service) loadCaptureUploadSession(uploadID string) (captureUploadSession, error) {
	if !isValidUploadID(uploadID) {
		return captureUploadSession{}, fmt.Errorf("loadCaptureUploadSession: invalid upload id")
	}

	data, err := os.ReadFile(s.captureUploadSessionMetaPath(uploadID))
	if err != nil {
		return captureUploadSession{}, fmt.Errorf("loadCaptureUploadSession: read session meta: %w", err)
	}

	var session captureUploadSession
	if err := json.Unmarshal(data, &session); err != nil {
		return captureUploadSession{}, fmt.Errorf("loadCaptureUploadSession: parse session meta: %w", err)
	}
	return session, nil
}

func isValidUploadID(uploadID string) bool {
	if len(uploadID) != 32 {
		return false
	}
	_, err := strconv.ParseUint(uploadID[:16], 16, 64)
	if err != nil {
		return false
	}
	_, err = strconv.ParseUint(uploadID[16:], 16, 64)
	return err == nil
}
