package video

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"os"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

const (
	maxSubtitleBytes              = 5 * 1024 * 1024
	subtitleCacheControl          = "public, max-age=3600"
	webVttContentType             = "text/vtt; charset=utf-8"
	subtitleListCacheControl      = "no-cache"
	unsupportedSubtitleMessageKey = "ERROR_VIDEO_SUBTITLE_UNSUPPORTED"
)

type SubtitleFileSource interface {
	GetFileById(id int) (files.FileDto, error)
}

type SubtitleHandler struct {
	fileSource SubtitleFileSource
	service    SubtitleServiceInterface
	embedded   EmbeddedSubtitleServiceInterface
}

func NewSubtitleHandler(fileSource SubtitleFileSource, service SubtitleServiceInterface) *SubtitleHandler {
	return &SubtitleHandler{fileSource: fileSource, service: service}
}

func (h *SubtitleHandler) UseEmbeddedSubtitles(embedded EmbeddedSubtitleServiceInterface) {
	h.embedded = embedded
}

func (h *SubtitleHandler) ListVideoSubtitlesHandler(c *gin.Context) {
	videoID := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	videoFile, err := h.fileSource.GetFileById(videoID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && videoFile.DeletedAt.HasValue) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_NOT_FOUND")})
		return
	}
	if err != nil {
		respondSubtitleFailure(c, err)
		return
	}
	subtitles, err := h.service.ListVideoSubtitles(videoFile.ParentPath, videoFile.Name)
	if err != nil {
		respondSubtitleFailure(c, err)
		return
	}
	if h.embedded != nil {
		subtitles = append(subtitles, h.embedded.ListEmbeddedSubtitles(videoFile)...)
	}
	c.Header("Cache-Control", subtitleListCacheControl)
	c.JSON(http.StatusOK, subtitles)
}

func (h *SubtitleHandler) GetEmbeddedSubtitleWebVttHandler(c *gin.Context) {
	videoID := utils.ParseInt(c.Param("id"), c)
	streamIndex := utils.ParseInt(c.Param("stream_index"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	if h.embedded == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_SUBTITLE_NOT_FOUND")})
		return
	}
	videoFile, err := h.fileSource.GetFileById(videoID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && videoFile.DeletedAt.HasValue) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_NOT_FOUND")})
		return
	}
	if err != nil {
		respondSubtitleFailure(c, err)
		return
	}

	subtitleETag := buildEmbeddedSubtitleETag(videoFile, streamIndex)
	if c.GetHeader("If-None-Match") == subtitleETag {
		c.Header("ETag", subtitleETag)
		c.Header("Cache-Control", subtitleCacheControl)
		c.Status(http.StatusNotModified)
		return
	}

	webVtt, err := h.embedded.ReadEmbeddedSubtitleWebVtt(videoFile, streamIndex)
	if errors.Is(err, ErrEmbeddedSubtitleNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_SUBTITLE_NOT_FOUND")})
		return
	}
	if err != nil {
		respondSubtitleFailure(c, err)
		return
	}
	c.Header("ETag", subtitleETag)
	c.Header("Cache-Control", subtitleCacheControl)
	c.Data(http.StatusOK, webVttContentType, webVtt)
}

func buildEmbeddedSubtitleETag(videoFile files.FileDto, streamIndex int) string {
	return fmt.Sprintf(`"subtitle-embedded-%d-%d-%d-%d"`, videoFile.ID, streamIndex, videoFile.UpdatedAt.UnixNano(), videoFile.Size)
}

func (h *SubtitleHandler) GetSubtitleWebVttHandler(c *gin.Context) {
	subtitleID := utils.ParseInt(c.Param("id"), c)
	if c.IsAborted() {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}
	subtitleFile, err := h.fileSource.GetFileById(subtitleID)
	if errors.Is(err, sql.ErrNoRows) || (err == nil && subtitleFile.DeletedAt.HasValue) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_SUBTITLE_NOT_FOUND")})
		return
	}
	if err != nil {
		respondSubtitleFailure(c, err)
		return
	}
	subtitleFormat, isSubtitle := SubtitleFormatOfName(subtitleFile.Name)
	if !isSubtitle {
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage(unsupportedSubtitleMessageKey)})
		return
	}

	subtitleETag := buildSubtitleETag(subtitleFile)
	c.Header("ETag", subtitleETag)
	c.Header("Cache-Control", subtitleCacheControl)
	if c.GetHeader("If-None-Match") == subtitleETag {
		c.Status(http.StatusNotModified)
		return
	}

	rawSubtitle, err := readSubtitleFile(subtitleFile.ResolveContentPath())
	if err != nil {
		c.Header("ETag", "")
		c.Header("Cache-Control", "no-store")
		respondSubtitleReadFailure(c, err)
		return
	}
	webVtt := ConvertSubtitleToWebVtt(DecodeSubtitleText(rawSubtitle), subtitleFormat)
	c.Data(http.StatusOK, webVttContentType, []byte(webVtt))
}

var errSubtitleTooLarge = errors.New("subtitle file is too large")

func readSubtitleFile(contentPath string) ([]byte, error) {
	fileInfo, err := os.Stat(contentPath)
	if err != nil {
		return nil, err
	}
	if fileInfo.Size() > maxSubtitleBytes {
		return nil, errSubtitleTooLarge
	}
	return os.ReadFile(contentPath)
}

func buildSubtitleETag(subtitleFile files.FileDto) string {
	return fmt.Sprintf(`"subtitle-%d-%d-%d"`, subtitleFile.ID, subtitleFile.UpdatedAt.UnixNano(), subtitleFile.Size)
}

func respondSubtitleReadFailure(c *gin.Context, err error) {
	if errors.Is(err, errSubtitleTooLarge) {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_SUBTITLE_TOO_LARGE")})
		return
	}
	if errors.Is(err, os.ErrNotExist) {
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_SUBTITLE_NOT_FOUND")})
		return
	}
	respondSubtitleFailure(c, err)
}

func respondSubtitleFailure(c *gin.Context, err error) {
	applog.ErrorWithStack("video: subtitle operation failed", err, "path", c.FullPath(), "ip", c.ClientIP())
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_VIDEO_OPERATION_FAILED")})
}
