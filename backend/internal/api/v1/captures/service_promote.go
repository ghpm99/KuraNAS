package captures

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/api/v1/libraries"
	"nas-go/api/internal/api/v1/notifications"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const (
	posterDownloadTimeout = 8 * time.Second
	maxPosterBytes        = 10 * 1024 * 1024
	remuxTimeout          = 30 * time.Minute
)

var remuxRecording = ffmpegRemuxForPlayback

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

	if markErr := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.MarkCapturePromoting(tx, captureID)
	}); markErr != nil {
		return fmt.Errorf("PromoteCapture: mark capture %d as promoting: %w", captureID, markErr)
	}

	stagingPath := capture.FilePath
	stagingDir := filepath.Dir(stagingPath)

	meta := parseCaptureMetadata(capture.RawMetadata)

	lib, err := s.LibrariesProvider.GetLibraryByCategory(libraries.LibraryCategoryVideos)
	if err != nil {
		return s.failPromotion(capture, stagingPath, promotionFailureReasonKey(err, "CAPTURE_PROMOTION_FAILED_LIBRARY_UNAVAILABLE"),
			fmt.Errorf("PromoteCapture: resolve videos library: %w", err))
	}

	finalPath := collisionAvoidantPath(filepath.Join(lib.Path, buildCaptureRelPath(meta, capture)))
	finalName := filepath.Base(finalPath)
	finalDir := filepath.Dir(finalPath)

	if mkdirErr := os.MkdirAll(finalDir, 0755); mkdirErr != nil {
		return s.failPromotion(capture, stagingPath, promotionFailureReasonKey(mkdirErr, "CAPTURE_PROMOTION_FAILED_LIBRARY_UNAVAILABLE"),
			fmt.Errorf("PromoteCapture: create destination dir: %w", mkdirErr))
	}

	now := time.Now()
	stub := files.FileDto{
		Name:       finalName,
		Path:       finalPath,
		ParentPath: finalDir,
		Type:       files.File,
		Format:     utils.ExtensionOf(finalName),
		Size:       capture.Size,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	created, createErr := s.FilesProvider.CreateFile(stub)
	if createErr != nil {
		return s.failPromotion(capture, stagingPath, promotionFailureReasonKey(createErr, "CAPTURE_PROMOTION_FAILED_REGISTER"),
			fmt.Errorf("PromoteCapture: pre-register home_file: %w", createErr))
	}
	fileID := created.ID

	promotingCapture := capture
	capture.FileID = &fileID
	capture.FileName = finalName
	capture.FilePath = finalPath
	capture.Status = CaptureStatusPromoted
	applyMetadataToCapture(&capture, meta)

	if updateErr := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.UpdateCapturePromotion(tx, capture)
	}); updateErr != nil {
		_ = s.FilesProvider.DeleteFileRecord(fileID)
		return s.failPromotion(promotingCapture, stagingPath, promotionFailureReasonKey(updateErr, "CAPTURE_PROMOTION_FAILED_REGISTER"),
			fmt.Errorf("PromoteCapture: persist capture metadata: %w", updateErr))
	}

	s.downloadPosterBestEffort(meta, fileID)

	if placementErr := placeRecordingInLibrary(stagingPath, finalPath); placementErr != nil {
		s.deletePreRegisteredFile(captureID, fileID)
		return s.failPromotion(promotingCapture, stagingPath, placementErr.reasonKey,
			fmt.Errorf("PromoteCapture: place recording into library: %w", placementErr))
	}

	_ = os.RemoveAll(stagingDir)
	s.emitPromotionCompletedNotification(capture)
	return nil
}

type recordingPlacementError struct {
	reasonKey string
	cause     error
}

func (e *recordingPlacementError) Error() string {
	return e.cause.Error()
}

func (e *recordingPlacementError) Unwrap() error {
	return e.cause
}

func (e *recordingPlacementError) failureReason() string {
	return i18n.GetMessage(e.reasonKey)
}

func placeRecordingInLibrary(stagingPath, finalPath string) *recordingPlacementError {
	if isRemuxableContainer(finalPath) {
		if err := remuxRecording(stagingPath, finalPath); err != nil {
			_ = os.Remove(finalPath)
			applog.Error("captures: remux failed", "src", stagingPath, "dest", finalPath, "error", err)
			return &recordingPlacementError{reasonKey: "CAPTURE_PROMOTION_FAILED_REMUX", cause: err}
		}
		return nil
	}
	if err := moveFile(stagingPath, finalPath); err != nil {
		return &recordingPlacementError{reasonKey: "CAPTURE_PROMOTION_FAILED_MOVE", cause: err}
	}
	return nil
}

func isRemuxableContainer(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".m4v", ".mov":
		return true
	default:
		return false
	}
}

func ffmpegRemuxForPlayback(srcPath, destPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), remuxTimeout)
	defer cancel()

	out, err := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", srcPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-movflags", "+faststart",
		destPath,
	).CombinedOutput()
	if err != nil {
		_ = os.Remove(destPath)
		return fmt.Errorf("ffmpegRemuxForPlayback: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func promotionFailureReasonKey(cause error, defaultReasonKey string) string {
	if errors.Is(cause, syscall.ENOSPC) {
		return "CAPTURE_PROMOTION_FAILED_NO_SPACE"
	}
	return defaultReasonKey
}

func (s *Service) deletePreRegisteredFile(captureID int, fileID int) {
	if delErr := s.FilesProvider.DeleteFileRecord(fileID); delErr != nil {
		applog.Error("captures: rollback delete home_file failed", "capture_id", captureID, "file_id", fileID, "error", delErr)
	}
	removeOrphanPoster(captureID, fileID)
}

func removeOrphanPoster(captureID int, fileID int) {
	if removeErr := os.Remove(posterSourcePath(fileID)); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		applog.Error("captures: rollback remove poster failed", "capture_id", captureID, "file_id", fileID, "error", removeErr)
	}
}

func (s *Service) failPromotion(capture CaptureModel, receivedFilePath string, reasonKey string, promoteErr error) error {
	failureReason := i18n.GetMessage(reasonKey)
	if statusErr := s.withTransaction(func(tx *sql.Tx) error {
		return s.Repository.MarkCapturePromotionFailed(tx, capture.ID, capture.FileName, receivedFilePath, failureReason)
	}); statusErr != nil {
		applog.Error("captures: mark capture promotion failed", "capture_id", capture.ID, "error", statusErr)
	}
	applog.Error("captures: promotion failed", "capture_id", capture.ID, "reason_key", reasonKey, "error", promoteErr)
	s.emitPromotionFailedNotification(capture, failureReason)
	return promoteErr
}

var thumbnailRootPath = func() string {
	return config.GetBuildConfig("ThumbnailPath")
}

func posterSourcePath(fileID int) string {
	return filepath.Join(thumbnailRootPath(), "video", "source", strconv.Itoa(fileID))
}

func (s *Service) downloadPosterBestEffort(meta captureMetadata, fileID int) {
	url := strings.TrimSpace(meta.ThumbnailURL)
	if url == "" {
		url = strings.TrimSpace(meta.PosterURL)
	}
	if url == "" || !strings.HasPrefix(strings.ToLower(url), "https://") {
		return
	}

	data, err := fetchPosterImage(url)
	if err != nil {
		applog.Error("captures: poster download failed", "file_id", fileID, "url", url, "error", err)
		return
	}
	if err := writePosterSource(fileID, data); err != nil {
		applog.Error("captures: poster write failed", "file_id", fileID, "error", err)
	}
}

func fetchPosterImage(url string) ([]byte, error) {
	client := &http.Client{Timeout: posterDownloadTimeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetchPosterImage: bad status %d", resp.StatusCode)
	}
	if !strings.HasPrefix(strings.ToLower(resp.Header.Get("Content-Type")), "image/") {
		return nil, fmt.Errorf("fetchPosterImage: non-image content type %q", resp.Header.Get("Content-Type"))
	}

	data, err := io.ReadAll(io.LimitReader(resp.Body, maxPosterBytes))
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, fmt.Errorf("fetchPosterImage: empty body")
	}
	return data, nil
}

func writePosterSource(fileID int, data []byte) error {
	destPath := posterSourcePath(fileID)
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(destPath, data, 0644)
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

func (s *Service) emitPromotionFailedNotification(capture CaptureModel, failureReason string) {
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
			"error":      failureReason,
		},
	})
}
