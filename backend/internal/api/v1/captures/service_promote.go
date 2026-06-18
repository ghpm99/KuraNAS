package captures

import (
	"context"
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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	posterDownloadTimeout = 8 * time.Second
	maxPosterBytes        = 10 * 1024 * 1024
	// remuxTimeout bounds the ffmpeg remux. Video is stream-copied (no re-encode)
	// so this is I/O-bound; the generous ceiling only guards a stalled process on
	// a very large file. The worker step itself has no context to cancel us.
	remuxTimeout = 30 * time.Minute
)

// remuxRecording rewrites a captured recording into a broadly-compatible MP4.
// It is a package var so tests can swap the real ffmpeg call for a stub.
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

	if moveErr := placeRecordingInLibrary(stagingPath, finalPath); moveErr != nil {
		s.rollbackPromotion(captureID, fileID)
		promoteErr := fmt.Errorf("PromoteCapture: move recording into library: %w", moveErr)
		s.emitPromotionFailedNotification(capture, promoteErr)
		return promoteErr
	}

	_ = os.RemoveAll(stagingDir)
	s.emitPromotionCompletedNotification(capture)
	return nil
}

// placeRecordingInLibrary moves the staged recording into the watched library,
// remuxing MP4 captures into a broadly-compatible file first. The plugin emits
// fragmented MP4 with Opus audio (the browser had no AAC encoder), which the
// Android/ExoPlayer Mp4 demuxer cannot play; the remux copies the video stream
// untouched, transcodes the audio to AAC and writes a faststart (defragmented)
// file. ffmpeg is already a runtime dependency (video thumbnails); if it is
// missing or the remux fails, fall back to a plain move so the recording is
// still promoted — just not necessarily playable everywhere — and log it for
// follow-up. Non-MP4 containers (e.g. WebM) are moved as-is.
func placeRecordingInLibrary(stagingPath, finalPath string) error {
	if isRemuxableContainer(finalPath) {
		if err := remuxRecording(stagingPath, finalPath); err != nil {
			applog.Error("captures: remux failed, falling back to plain move", "src", stagingPath, "dest", finalPath, "error", err)
		} else {
			return nil
		}
	}
	return moveFile(stagingPath, finalPath)
}

// isRemuxableContainer reports whether the destination is an MP4-family container
// where stream-copying the (H.264) video into a faststart MP4 is valid.
func isRemuxableContainer(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".mp4", ".m4v", ".mov":
		return true
	default:
		return false
	}
}

// ffmpegRemuxForPlayback writes destPath from srcPath: video stream-copied, audio
// transcoded to AAC, faststart. A partial output is removed on failure so the
// caller's fallback move can recreate the destination.
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
// only https URLs are accepted, and any failure is logged and swallowed so the
// promotion is never gated on artwork.
func (s *Service) downloadPoster(meta captureMetadata, fileID int) {
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

// fetchPosterImage GETs the URL with a short timeout and returns its bytes,
// rejecting a non-200 status, a non-image content type, and capping the read at
// maxPosterBytes. The scheme guard lives in the caller.
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
