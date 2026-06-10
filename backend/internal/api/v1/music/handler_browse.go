package music

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"

	files "nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/utils"

	"github.com/gin-gonic/gin"
)

func (handler *Handler) GetMusicHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetMusic", "Fetching music files", c), nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	pagination, err := handler.service.GetMusic(page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, files.ParsePaginationToResponse(pagination))
}

func (handler *Handler) GetMusicArtistsHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetMusicArtists", "Fetching music artists", c), nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetMusicArtists(page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicByArtistHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetMusicByArtist", "Fetching music by artist", c), nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	artist := c.Param("name")

	pagination, err := handler.service.GetMusicByArtist(artist, page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, files.ParsePaginationToResponse(pagination))
}

func (handler *Handler) GetMusicAlbumsHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetMusicAlbums", "Fetching music albums", c), nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetMusicAlbums(page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicByAlbumHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetMusicByAlbum", "Fetching music by album", c), nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	album := c.Param("name")

	pagination, err := handler.service.GetMusicByAlbum(album, page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, files.ParsePaginationToResponse(pagination))
}

func (handler *Handler) GetMusicGenresHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetMusicGenres", "Fetching music genres", c), nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetMusicGenres(page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicByGenreHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetMusicByGenre", "Fetching music by genre", c), nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	genre := c.Param("name")

	pagination, err := handler.service.GetMusicByGenre(genre, page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, files.ParsePaginationToResponse(pagination))
}

func (handler *Handler) GetMusicFoldersHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("GetMusicFolders", "Fetching music folders", c), nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetMusicFolders(page, pageSize)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

// StreamAudioHandler streams audio files with HTTP Range support.
// Served at GET /files/stream/:id — path unchanged, now owned by music.
func (handler *Handler) StreamAudioHandler(c *gin.Context) {
	loggerModel, _ := handler.logService.CreateLog(logEntry("StreamAudio", "Streaming audio file", c), nil)

	id := utils.ParseInt(c.Param("id"), c)

	file, err := handler.filesService.GetFileById(id)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	exists := handler.filesService.CheckFileExistsByPath(file.Path)
	if !exists {
		handler.logService.CompleteWithErrorLog(loggerModel, fmt.Errorf("file not found on disk"))
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	audioFile, err := os.Open(file.Path)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}
	defer audioFile.Close()

	fileInfo, err := audioFile.Stat()
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_INTERNAL")})
		return
	}

	contentType := contentTypeByFormat(file.Format, "audio/mpeg")
	c.Header("Content-Type", contentType)
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=3600")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		start, end, ok := parseHTTPRange(rangeHeader, fileInfo.Size())
		if ok {
			c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
			c.Header("Content-Length", fmt.Sprintf("%d", end-start+1))
			c.Status(http.StatusPartialContent)

			audioFile.Seek(start, 0)
			_, err := io.CopyN(c.Writer, audioFile, end-start+1)
			if err != nil {
				handler.logService.CompleteWithErrorLog(loggerModel, err)
				return
			}

			handler.logService.CompleteWithSuccessLog(loggerModel)
			return
		}
	}

	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, audioFile)
	if err != nil {
		handler.logService.CompleteWithErrorLog(loggerModel, err)
		return
	}

	handler.logService.CompleteWithSuccessLog(loggerModel)
}

func parseHTTPRange(rangeHeader string, fileSize int64) (int64, int64, bool) {
	if fileSize <= 0 {
		return 0, 0, false
	}

	parts := strings.SplitN(strings.TrimSpace(rangeHeader), "=", 2)
	if len(parts) != 2 || parts[0] != "bytes" {
		return 0, 0, false
	}

	rangeValue := strings.TrimSpace(parts[1])
	if rangeValue == "" {
		return 0, 0, false
	}

	if commaIndex := strings.Index(rangeValue, ","); commaIndex >= 0 {
		rangeValue = strings.TrimSpace(rangeValue[:commaIndex])
	}

	bounds := strings.SplitN(rangeValue, "-", 2)
	if len(bounds) != 2 {
		return 0, 0, false
	}

	startText := strings.TrimSpace(bounds[0])
	endText := strings.TrimSpace(bounds[1])

	var start int64
	var end int64
	var err error

	if startText == "" {
		suffixLength, parseErr := strconv.ParseInt(endText, 10, 64)
		if parseErr != nil || suffixLength <= 0 {
			return 0, 0, false
		}
		if suffixLength > fileSize {
			suffixLength = fileSize
		}
		start = fileSize - suffixLength
		end = fileSize - 1
	} else {
		start, err = strconv.ParseInt(startText, 10, 64)
		if err != nil || start < 0 || start >= fileSize {
			return 0, 0, false
		}

		if endText == "" {
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(endText, 10, 64)
			if err != nil {
				return 0, 0, false
			}
			if end >= fileSize {
				end = fileSize - 1
			}
		}
	}

	if end < start {
		return 0, 0, false
	}

	return start, end, true
}

func contentTypeByFormat(format string, fallback string) string {
	ext := strings.TrimSpace(format)
	if ext == "" {
		return fallback
	}
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	contentType := mime.TypeByExtension(strings.ToLower(ext))
	if contentType == "" {
		return fallback
	}
	return contentType
}
