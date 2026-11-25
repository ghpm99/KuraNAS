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
	"nas-go/api/pkg/database"
	"nas-go/api/pkg/diskspace"
	"nas-go/api/pkg/i18n"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const takeoutUploadChunkSize int64 = 2 * 1024 * 1024

const takeoutUploadIDLength = 32

type notificationServiceInterface interface {
	GroupOrCreate(dto notifications.CreateNotificationDto) (notifications.NotificationDto, error)
}

type Service struct {
	UploadJobDispatcher UploadJobDispatcherInterface
	LibraryResolver     LibraryResolverInterface
	NotificationService notificationServiceInterface
	ImportCreator       ImportCreatorInterface
	AvailableBytesAt    func(path string) (int64, error)
}

func NewService(
	uploadJobDispatcher UploadJobDispatcherInterface,
	libraryResolver LibraryResolverInterface,
	notificationService notificationServiceInterface,
	importCreator ImportCreatorInterface,
) ServiceInterface {
	return &Service{
		UploadJobDispatcher: uploadJobDispatcher,
		LibraryResolver:     libraryResolver,
		NotificationService: notificationService,
		ImportCreator:       importCreator,
		AvailableBytesAt:    diskspace.AvailableBytesAt,
	}
}

func (s *Service) withTransaction(fn func(tx *sql.Tx) error) error {
	if s.UploadJobDispatcher == nil {
		return fmt.Errorf("upload job dispatcher is not configured")
	}
	return database.ExecOptionalTx(s.UploadJobDispatcher.GetDbContext(), fn)
}

func hasZipExtension(fileName string) bool {
	return strings.EqualFold(filepath.Ext(fileName), ".zip")
}

func (s *Service) InitUpload(dto InitTakeoutUploadDto) (InitTakeoutUploadResultDto, error) {
	if dto.Size <= 0 {
		return InitTakeoutUploadResultDto{}, ErrInvalidUploadSize
	}
	fileName := sanitizeTakeoutFileName(dto.FileName)
	if fileName == "" || fileName == "unnamed" {
		return InitTakeoutUploadResultDto{}, ErrInvalidUploadFileName
	}
	if !hasZipExtension(fileName) {
		return InitTakeoutUploadResultDto{}, ErrInvalidUploadFileName
	}

	openSession, hasOpenSession := s.findOpenSessionByFile(fileName, dto.Size)
	requiredBytes := dto.Size
	if hasOpenSession {
		requiredBytes = dto.Size - openSession.ReceivedSize
	}
	if err := s.ensureFreeSpace(requiredBytes); err != nil {
		return InitTakeoutUploadResultDto{}, err
	}

	if hasOpenSession {
		return InitTakeoutUploadResultDto{
			UploadID:     openSession.UploadID,
			ChunkSize:    takeoutUploadChunkSize,
			ReceivedSize: openSession.ReceivedSize,
			Resumed:      true,
		}, nil
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

func (s *Service) ensureFreeSpace(requiredBytes int64) error {
	availableBytes, err := s.availableBytesAtUploadRoot()
	if err != nil {
		return fmt.Errorf("InitUpload read free space: %w", err)
	}
	if availableBytes < requiredBytes {
		return ErrInsufficientStorage
	}
	return nil
}

func (s *Service) availableBytesAtUploadRoot() (int64, error) {
	measure := s.AvailableBytesAt
	if measure == nil {
		measure = diskspace.AvailableBytesAt
	}
	existingPath := s.takeoutUploadRootDir()
	for {
		if _, err := os.Stat(existingPath); err == nil {
			return measure(existingPath)
		}
		parentPath := filepath.Dir(existingPath)
		if parentPath == existingPath {
			return measure(existingPath)
		}
		existingPath = parentPath
	}
}

func (s *Service) GetUploadStatus(uploadID string) (TakeoutUploadStatusDto, error) {
	if !isGeneratedUploadID(uploadID) {
		return TakeoutUploadStatusDto{}, ErrUploadSessionNotFound
	}
	session, err := s.loadTakeoutUploadSession(uploadID)
	if err != nil {
		return TakeoutUploadStatusDto{}, err
	}
	return TakeoutUploadStatusDto{
		UploadID:     session.UploadID,
		Size:         session.ExpectedSize,
		ReceivedSize: session.ReceivedSize,
	}, nil
}

func (s *Service) CancelUpload(uploadID string) error {
	if !isGeneratedUploadID(uploadID) {
		return ErrUploadSessionNotFound
	}
	if _, err := s.loadTakeoutUploadSession(uploadID); err != nil {
		return err
	}
	sessionDir, isInsideSessionsArea := UploadSessionDirOfZip(s.takeoutUploadTempFilePath(uploadID))
	if !isInsideSessionsArea {
		return ErrUploadSessionNotFound
	}
	if err := os.RemoveAll(sessionDir); err != nil {
		return fmt.Errorf("CancelUpload remove session dir: %w", err)
	}
	return nil
}

func isGeneratedUploadID(uploadID string) bool {
	if len(uploadID) != takeoutUploadIDLength {
		return false
	}
	_, err := hex.DecodeString(uploadID)
	return err == nil
}

func (s *Service) findOpenSessionByFile(fileName string, expectedSize int64) (TakeoutUploadSession, bool) {
	entries, err := os.ReadDir(s.takeoutUploadRootDir())
	if err != nil {
		return TakeoutUploadSession{}, false
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		session, loadErr := s.loadTakeoutUploadSession(entry.Name())
		if loadErr != nil {
			continue
		}
		isSameFile := session.FileName == fileName && session.ExpectedSize == expectedSize
		isOpen := session.ReceivedSize < session.ExpectedSize
		if isSameFile && isOpen {
			return session, true
		}
	}

	return TakeoutUploadSession{}, false
}

func (s *Service) UploadChunk(file *multipart.FileHeader, dto UploadTakeoutChunkDto) error {
	session, err := s.loadTakeoutUploadSession(dto.UploadID)
	if err != nil {
		return err
	}

	if dto.Offset != session.ReceivedSize {
		return ErrUploadOffsetMismatch
	}

	if session.ExpectedSize > 0 && session.ReceivedSize+file.Size > session.ExpectedSize {
		return ErrUploadChunkExceedsSize
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
	if !hasZipExtension(session.FileName) {
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
	var importID int
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
		if createStepErr != nil {
			return createStepErr
		}

		createdImportID, createImportErr := s.ImportCreator.CreateQueuedImport(tx, createdJob.ID, session.FileName, session.ExpectedSize)
		importID = createdImportID
		return createImportErr
	})
	if err != nil {
		return TakeoutImportResultDto{}, fmt.Errorf("CompleteUpload create job: %w", err)
	}

	return TakeoutImportResultDto{
		JobID:    createdJob.ID,
		ImportID: importID,
		Message:  i18n.GetMessage("TAKEOUT_IMPORT_STARTED"),
	}, nil
}

func (s *Service) takeoutUploadRootDir() string {
	return UploadSessionsRootDir()
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
