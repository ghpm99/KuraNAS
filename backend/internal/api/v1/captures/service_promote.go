package captures

import (
	"database/sql"
	"fmt"
	"io"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	posterDownloadTimeout = 8 * time.Second
	maxPosterBytes        = 10 * 1024 * 1024
)

// PromoteCapture turns an uploaded capture into an organized video in the
// library. Ordering matters: the home_file stub and the captures row are
// written first (so the DB is consistent), the poster is fetched best-effort,
// and the recording is moved into the watched library last. If the move fails,
// the stub and the promotion are rolled back and the recording stays in the
// staging folder for a later retry. Once the file lands in the library, the
// regular fsnotify pipeline indexes it and merges into the pre-registered row
// (idempotent by name+path) — the pipeline never needs to know about captures.
func (s *Service) PromoteCapture(captureID int) error {
	if s.LibrariesProvider == nil || s.FilesProvider == nil {
		return fmt.Errorf("PromoteCapture: promotion dependencies are not configured")
	}

	capture, err := s.Repository.GetCaptureByID(captureID)
	if err != nil {
		return fmt.Errorf("PromoteCapture: load capture %d: %w", captureID, err)
	}

	stagingPath := capture.FilePath
	stagingDir := filepath.Dir(stagingPath)

	meta := parseCaptureMetadata(capture.RawMetadata)

	lib, err := s.LibrariesProvider.GetLibraryByCategory(libraries.LibraryCategoryVideos)
	if err != nil {
		return fmt.Errorf("PromoteCapture: resolve videos library: %w", err)
	}

	finalPath := collisionAvoidantPath(filepath.Join(lib.Path, buildCaptureRelPath(meta, capture)))
	finalName := filepath.Base(finalPath)
	finalDir := filepath.Dir(finalPath)

	if mkdirErr := os.MkdirAll(finalDir, 0755); mkdirErr != nil {
		return fmt.Errorf("PromoteCapture: create destination dir: %w", mkdirErr)
	}

	now := time.Now()
	stub := files.FileDto{
		Name:       finalName,
		Path:       finalPath,
		ParentPath: finalDir,
		Type:       files.File,
		Format:     filepath.Ext(finalName),
		Size:       capture.Size,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	created, createErr := s.FilesProvider.CreateFile(stub)
	if createErr != nil {
		return fmt.Errorf("PromoteCapture: pre-register home_file: %w", createErr)
	}
	fileID := created.ID

	capture.FileID = &fileID
	capture.FileName = finalName
	capture.FilePath = finalPath
	capture.Status = CaptureStatusPromoted
	applyMetadataToCapture(&capture, meta)

	if updateErr := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.UpdateCapturePromotion(tx, capture)
	}); updateErr != nil {
		_ = s.FilesProvider.DeleteFileRecord(fileID)
		return fmt.Errorf("PromoteCapture: persist capture metadata: %w", updateErr)
	}

	// Poster is fetched before the move and never gates promotion: a failed
	// download just means the thumbnail step later falls back to an ffmpeg frame.
	s.downloadPoster(meta, fileID)

	if moveErr := moveFile(stagingPath, finalPath); moveErr != nil {
		s.rollbackPromotion(captureID, fileID)
		promoteErr := fmt.Errorf("PromoteCapture: move recording into library: %w", moveErr)
		s.emitPromotionFailedNotification(capture, promoteErr)
		return promoteErr
	}

	_ = os.RemoveAll(stagingDir)
	s.emitPromotionCompletedNotification(capture)
	return nil
}

// rollbackPromotion undoes the DB side of a promotion whose move failed: the
// pre-registered home_file stub is removed and the capture is flipped to failed
// with its file_id detached, leaving the recording in staging for a retry.
func (s *Service) rollbackPromotion(captureID int, fileID int) {
	if delErr := s.FilesProvider.DeleteFileRecord(fileID); delErr != nil {
		applog.Error("captures: rollback delete home_file failed", "capture_id", captureID, "file_id", fileID, "error", delErr)
	}
	if statusErr := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.UpdateCaptureStatus(tx, captureID, CaptureStatusFailed, nil)
	}); statusErr != nil {
		applog.Error("captures: rollback capture status failed", "capture_id", captureID, "error", statusErr)
	}
}

// posterSourcePath is the agreed location of a capture's source poster:
// <ThumbnailPath>/video/source/<file_id>. video.GetVideoThumbnail reads from
// here before falling back to an ffmpeg frame — the two sides must stay in sync.
func posterSourcePath(fileID int) string {
	return filepath.Join(config.GetBuildConfig("ThumbnailPath"), "video", "source", strconv.Itoa(fileID))
}

// downloadPoster fetches the capture's poster (thumbnail_url, falling back to
// poster_url) and stores it as the source poster. It is strictly best-effort:
// only https URLs, a short timeout, a ~10MB cap and an image/* content type are
// accepted, and any failure is logged and swallowed.
func (s *Service) downloadPoster(meta captureMetadata, fileID int) {
	url := strings.TrimSpace(meta.ThumbnailURL)
	if url == "" {
		url = strings.TrimSpace(meta.PosterURL)
	}
	if url == "" || !strings.HasPrefix(strings.ToLower(url), "https://") {
		return
	}

	client := &http.Client{Timeout: posterDownloadTimeout}
	resp, err := client.Get(url)
	if err != nil {
		applog.Error("captures: poster download failed", "file_id", fileID, "url", url, "error", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		applog.Error("captures: poster download bad status", "file_id", fileID, "url", url, "status", resp.StatusCode)
		return
	}
	if !strings.HasPrefix(strings.ToLower(resp.Header.Get("Content-Type")), "image/") {
		applog.Error("captures: poster download non-image", "file_id", fileID, "content_type", resp.Header.Get("Content-Type"))
		return
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxPosterBytes))
	if err != nil || len(data) == 0 {
		applog.Error("captures: poster read failed", "file_id", fileID, "error", err)
		return
	}

	destPath := posterSourcePath(fileID)
	if mkErr := os.MkdirAll(filepath.Dir(destPath), 0755); mkErr != nil {
		applog.Error("captures: poster dir create failed", "file_id", fileID, "error", mkErr)
		return
	}
	if writeErr := os.WriteFile(destPath, data, 0644); writeErr != nil {
		applog.Error("captures: poster write failed", "file_id", fileID, "error", writeErr)
	}
}

func (s *Service) emitPromotionCompletedNotification(capture CaptureModel) {
	if s.NotificationService == nil {
		return
	}
	_, _ = s.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
		Type:     string(notifications.NotificationTypeSuccess),
		Title:    i18n.GetMessage("NOTIFICATION_CAPTURE_PROMOTED_TITLE"),
		Message:  i18n.Translate("NOTIFICATION_CAPTURE_PROMOTED_MESSAGE", capture.Title),
		GroupKey: "capture_promotion_result",
		Metadata: map[string]any{
			"event":      "capture_promoted",
			"capture_id": capture.ID,
			"file_id":    capture.FileID,
			"file_path":  capture.FilePath,
		},
	})
}

func (s *Service) emitPromotionFailedNotification(capture CaptureModel, promoteErr error) {
	if s.NotificationService == nil {
		return
	}
	_, _ = s.NotificationService.GroupOrCreate(notifications.CreateNotificationDto{
		Type:     string(notifications.NotificationTypeError),
		Title:    i18n.GetMessage("NOTIFICATION_CAPTURE_PROMOTION_FAILED_TITLE"),
		Message:  i18n.Translate("NOTIFICATION_CAPTURE_PROMOTION_FAILED_MESSAGE", capture.Name),
		GroupKey: "",
		Metadata: map[string]any{
			"event":      "capture_promotion_failed",
			"capture_id": capture.ID,
			"error":      promoteErr.Error(),
		},
	})
}
