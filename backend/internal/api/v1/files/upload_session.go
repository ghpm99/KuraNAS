package files

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/applog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

const (
	uploadChunkSize              = 2 << 20
	uploadSessionsDirName        = ".files-uploads"
	uploadSessionFallbackSubdir  = "capturas"
	uploadSessionMetaFileName    = "meta.json"
	uploadSessionPartialFileName = "payload.part"
)

var uploadIDPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

var uploadSessionLocks sync.Map

// UploadSession is the on-disk state of one resumable upload: where the file
// will be published, what the client declared, and how much was received.
type UploadSession struct {
	UploadID         string               `json:"upload_id"`
	FileName         string               `json:"file_name"`
	Size             int64                `json:"size"`
	Checksum         string               `json:"checksum"`
	ModifiedAt       *time.Time           `json:"modified_at"`
	OnConflict       UploadConflictPolicy `json:"on_conflict"`
	TargetFolderPath string               `json:"target_folder_path"`
	RelativePath     string               `json:"relative_path"`
	ReceivedSize     int64                `json:"received_size"`
	CreatedAt        time.Time            `json:"created_at"`
	LastActivityAt   time.Time            `json:"last_activity_at"`
}

func (session UploadSession) destinationDirectory() string {
	if session.RelativePath == "" {
		return session.TargetFolderPath
	}
	return filepath.Join(session.TargetFolderPath, filepath.FromSlash(session.RelativePath))
}

func (session UploadSession) isEquivalentTo(other UploadSession) bool {
	if session.TargetFolderPath != other.TargetFolderPath || session.RelativePath != other.RelativePath {
		return false
	}
	if session.FileName != other.FileName || session.Size != other.Size {
		return false
	}
	return other.Checksum == "" || session.Checksum == other.Checksum
}

func uploadSessionsRootDir() string {
	capturesPath := strings.TrimSpace(config.AppConfig.CapturesPath)
	if capturesPath == "" {
		capturesPath = filepath.Join(config.AppConfig.EntryPoint, uploadSessionFallbackSubdir)
	}
	return filepath.Join(capturesPath, uploadSessionsDirName)
}

func uploadSessionDir(uploadID string) string {
	return filepath.Join(uploadSessionsRootDir(), uploadID)
}

func uploadSessionMetaPath(uploadID string) string {
	return filepath.Join(uploadSessionDir(uploadID), uploadSessionMetaFileName)
}

func uploadSessionPartialPath(uploadID string) string {
	return filepath.Join(uploadSessionDir(uploadID), uploadSessionPartialFileName)
}

func isValidUploadID(uploadID string) bool {
	return uploadIDPattern.MatchString(uploadID)
}

func generateUploadID() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}

func requireValidUploadID(uploadID string) error {
	if isValidUploadID(uploadID) {
		return nil
	}
	return newFileOperationError(http.StatusBadRequest, "ERROR_UPLOAD_ID_INVALID", fmt.Errorf("invalid upload id %q", uploadID))
}

func lockUploadSession(uploadID string) func() {
	lock, _ := uploadSessionLocks.LoadOrStore(uploadID, &sync.Mutex{})
	sessionLock := lock.(*sync.Mutex)
	sessionLock.Lock()
	return sessionLock.Unlock
}

func uploadInternalError(operation string, err error) *FileOperationError {
	applog.ErrorWithStack(operation, err)
	return newFileOperationError(http.StatusInternalServerError, "ERROR_UPLOAD_FAILED", err)
}

func uploadSessionNotFoundError(uploadID string) *FileOperationError {
	return newFileOperationError(http.StatusNotFound, "ERROR_UPLOAD_SESSION_NOT_FOUND", fmt.Errorf("upload session %q not found", uploadID))
}

func createUploadSession(session UploadSession) error {
	if !isValidUploadID(session.UploadID) {
		return fmt.Errorf("createUploadSession: invalid upload id")
	}
	if err := os.MkdirAll(uploadSessionDir(session.UploadID), 0755); err != nil {
		return fmt.Errorf("createUploadSession: create session dir: %w", err)
	}
	partialFile, err := os.OpenFile(uploadSessionPartialPath(session.UploadID), os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("createUploadSession: create partial file: %w", err)
	}
	if err := partialFile.Close(); err != nil {
		return fmt.Errorf("createUploadSession: close partial file: %w", err)
	}
	return saveUploadSession(session)
}

func saveUploadSession(session UploadSession) error {
	if !isValidUploadID(session.UploadID) {
		return fmt.Errorf("saveUploadSession: invalid upload id")
	}
	encodedSession, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("saveUploadSession: encode session: %w", err)
	}
	metaPath := uploadSessionMetaPath(session.UploadID)
	temporaryPath := metaPath + ".tmp"
	if err := os.WriteFile(temporaryPath, encodedSession, 0644); err != nil {
		return fmt.Errorf("saveUploadSession: write session: %w", err)
	}
	if err := os.Rename(temporaryPath, metaPath); err != nil {
		return fmt.Errorf("saveUploadSession: publish session: %w", err)
	}
	return nil
}

func loadUploadSession(uploadID string) (UploadSession, error) {
	if err := requireValidUploadID(uploadID); err != nil {
		return UploadSession{}, err
	}
	encodedSession, err := os.ReadFile(uploadSessionMetaPath(uploadID))
	if errors.Is(err, os.ErrNotExist) {
		return UploadSession{}, uploadSessionNotFoundError(uploadID)
	}
	if err != nil {
		return UploadSession{}, uploadInternalError("read upload session", err)
	}
	var session UploadSession
	if err := json.Unmarshal(encodedSession, &session); err != nil {
		return UploadSession{}, uploadInternalError("parse upload session", err)
	}
	return session, nil
}

func deleteUploadSession(uploadID string) error {
	if err := requireValidUploadID(uploadID); err != nil {
		return err
	}
	if err := os.RemoveAll(uploadSessionDir(uploadID)); err != nil {
		return uploadInternalError("delete upload session", err)
	}
	return nil
}

func findOpenUploadSession(requested UploadSession) (UploadSession, bool) {
	entries, err := os.ReadDir(uploadSessionsRootDir())
	if err != nil {
		return UploadSession{}, false
	}
	for _, entry := range entries {
		if !entry.IsDir() || !isValidUploadID(entry.Name()) {
			continue
		}
		candidate, loadErr := loadUploadSession(entry.Name())
		if loadErr != nil {
			continue
		}
		if candidate.isEquivalentTo(requested) {
			return candidate, true
		}
	}
	return UploadSession{}, false
}
