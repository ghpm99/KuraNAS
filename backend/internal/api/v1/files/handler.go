package files

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"strconv"
	"strings"
	"time"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"nas-go/api/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service           ServiceInterface
	recentFileService RecentFileServiceInterface
	Logger            logger.LoggerServiceInterface
}

func NewHandler(
	financialService ServiceInterface,
	recentFileService RecentFileServiceInterface,
	loggerService logger.LoggerServiceInterface,
) *Handler {
	return &Handler{
		service:           financialService,
		Logger:            loggerService,
		recentFileService: recentFileService,
	}
}

func (handler *Handler) GetFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFiles",
		Description: "Fetching files with filter",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	fileParent := utils.ParseInt(c.DefaultQuery("file_parent", "0"), c)

	filter := FileFilter{
		FileParent: utils.Optional[int]{
			HasValue: fileParent != 0,
			Value:    fileParent,
		},
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: filter,
	})

	pagination, err := handler.service.GetFiles(filter, page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetFilesByPathHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFilesByPath",
		Description: "Fetching files by path",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	path := c.DefaultQuery("path", config.AppConfig.EntryPoint)

	filter := FileFilter{
		Path: utils.Optional[string]{
			HasValue: true,
			Value:    path,
		},
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: filter,
	})

	pagination, err := handler.service.GetFiles(filter, page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetChildrenByIdHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetChildrenById",
		Description: "Fetching files by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)
	id := utils.ParseInt(c.Param("id"), c)

	filter := FileFilter{
		ID: utils.Optional[int]{
			HasValue: true,
			Value:    id,
		},
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: filter,
	})

	file, err := handler.service.GetFiles(filter, page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	pagination, err := handler.service.GetFiles(FileFilter{
		Path: utils.Optional[string]{
			HasValue: true,
			Value:    file.Items[0].Path,
		},
	}, page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) UpdateFilesHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "UpdateFiles",
		Description: "Updating files with data",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	data := c.PostForm("data")
	if data == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("data is required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DATA_REQUIRED")})
		return
	}
	loggerModel.SetExtraData(logger.LogExtraData{
		Data: data,
	})
	handler.service.ScanFilesTask(data)
	handler.Logger.CompleteWithSuccessLog(loggerModel)
}

func (handler *Handler) GetFilesTreeHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFilesTree",
		Description: "Fetching files with filter",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	fileParentId := utils.ParseInt(c.DefaultQuery("file_parent", "0"), c)

	fileCategory := c.DefaultQuery("category", string(AllCategory))

	fileFilter := FileFilter{
		DeletedAt: utils.Optional[time.Time]{
			HasValue: false,
		},
		Category: FileCategory(fileCategory),
	}

	if fileParentId != 0 {
		fileParent, err := handler.service.GetFileById(fileParentId)
		if err != nil {
			handler.Logger.CompleteWithErrorLog(loggerModel, err)
			fmt.Println("Error getting file by ID:", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if fileParent.ID != 0 {
			fileFilter.ParentPath = utils.Optional[string]{
				HasValue: true,
				Value:    fileParent.Path,
			}
		}
	} else {
		fileFilter.ParentPath = utils.Optional[string]{
			HasValue: true,
			Value:    config.AppConfig.EntryPoint,
		}
	}

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: fileFilter,
	})

	pagination, err := handler.service.GetFiles(fileFilter, page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetFileThumbnailHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFileThumbnail",
		Description: "Fetching file thumbnail by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)
	width := utils.ParseInt(c.DefaultQuery("width", "320"), c)
	height := utils.ParseInt(c.DefaultQuery("height", "320"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: map[string]int{"id": id, "width": width, "height": height},
	})

	file, err := handler.service.GetFileById(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	thumbnailData, err := handler.service.GetFileThumbnail(file, width, height)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		httpStatus := http.StatusInternalServerError
		if errors.Is(err, ErrFileMissingDisk) {
			httpStatus = http.StatusNotFound
		}
		c.JSON(httpStatus, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Header("Content-Type", "image/png")
	c.Header("Cache-Control", "public, max-age=86400")
	c.Data(http.StatusOK, "image/png", thumbnailData)
}

func (handler *Handler) GetBlobFileHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetBlobFile",
		Description: "Fetching file by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: id,
	})

	fileBlob, err := handler.service.GetFileBlobById(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.recentFileService.RegisterAccess(c.ClientIP(), fileBlob.ID, config.AppConfig.RecentFilesKeep)
	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Data(http.StatusOK, mime.TypeByExtension(strings.ToLower(fileBlob.Format)), fileBlob.Blob)
}

func (handler *Handler) GetRecentFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetRecentFiles",
		Description: "Fetching recent files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	recentFiles, err := handler.recentFileService.GetRecentFiles(page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, recentFiles)
}

func (handler *Handler) GetRecentAccessByFileHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetRecentAccessByFile",
		Description: "Fetching recent access by file ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: id,
	})

	recentFiles, err := handler.recentFileService.GetRecentAccessByFileID(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, recentFiles)
}

func (handler *Handler) StarreFileHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "StarFile",
		Description: "Starring a file by ID",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: id,
	})

	file, err := handler.service.GetFileById(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	file.Starred = !file.Starred

	result, err := handler.service.UpdateFile(file)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"success": result})
}

func (handler *Handler) GetTotalSpaceUsedHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetTotalSpaceUsed",
		Description: "Fetching total space used",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	totalSpaceUsed, err := handler.service.GetTotalSpaceUsed()

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"total_space_used": totalSpaceUsed})
}

func (handler *Handler) GetTotalFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetTotalFiles",
		Description: "Fetching total files count",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	totalFiles, err := handler.service.GetTotalFiles()

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"total_files": totalFiles})
}

func (handler *Handler) GetTotalDirectoryHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetTotalSpaceUsedByPath",
		Description: "Fetching total space used by path",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	totalSpaceUsed, err := handler.service.GetTotalDirectory()

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"total_directory": totalSpaceUsed})
}

func (handler *Handler) GetReportSizeByFormatHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetReportSizeByFormat",
		Description: "Fetching report size by format",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	report, err := handler.service.GetReportSizeByFormat()

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, report)
}

func (handler *Handler) GetTopFilesBySizeHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetTopFilesBySize",
		Description: "Fetching top files by size",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	limit := utils.ParseInt(c.DefaultQuery("limit", "5"), c)

	topFiles, err := handler.service.GetTopFilesBySize(limit)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, topFiles)
}

func (handler *Handler) GetDuplicateFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetDuplicateFiles",
		Description: "Fetching duplicate files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	report, err := handler.service.GetDuplicateFiles(page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, report)
}

func (handler *Handler) GetImagesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFilesTree",
		Description: "Fetching files with filter",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	pagination, err := handler.service.GetImages(page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetMusic",
		Description: "Fetching music files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	pagination, err := handler.service.GetMusic(page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicArtistsHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetMusicArtists",
		Description: "Fetching music artists",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetMusicArtists(page, pageSize)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicByArtistHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetMusicByArtist",
		Description: "Fetching music by artist",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	artist := c.Param("name")

	pagination, err := handler.service.GetMusicByArtist(artist, page, pageSize)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicAlbumsHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetMusicAlbums",
		Description: "Fetching music albums",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetMusicAlbums(page, pageSize)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicByAlbumHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetMusicByAlbum",
		Description: "Fetching music by album",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	album := c.Param("name")

	pagination, err := handler.service.GetMusicByAlbum(album, page, pageSize)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicGenresHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetMusicGenres",
		Description: "Fetching music genres",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetMusicGenres(page, pageSize)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicByGenreHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetMusicByGenre",
		Description: "Fetching music by genre",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)
	genre := c.Param("name")

	pagination, err := handler.service.GetMusicByGenre(genre, page, pageSize)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) GetMusicFoldersHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetMusicFolders",
		Description: "Fetching music folders",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "50"), c)

	pagination, err := handler.service.GetMusicFolders(page, pageSize)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

// StreamAudioHandler streams audio files with HTTP Range support
func (handler *Handler) StreamAudioHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "StreamAudio",
		Description: "Streaming audio file",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	file, err := handler.service.GetFileById(id)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	exists := handler.service.CheckFileExistsByPath(file.Path)
	if !exists {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("file not found on disk"))
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	audioFile, err := os.Open(file.Path)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer audioFile.Close()

	fileInfo, err := audioFile.Stat()
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "audio/mpeg")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=3600")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		// Parse Range header: "bytes=0-1023"
		ranges := strings.Split(rangeHeader, "=")
		if len(ranges) == 2 && ranges[0] == "bytes" {
			byteRange := strings.Split(ranges[1], "-")
			if len(byteRange) == 2 {
				start, _ := strconv.ParseInt(byteRange[0], 10, 64)
				end, _ := strconv.ParseInt(byteRange[1], 10, 64)

				// Validação do range
				if start >= 0 && end < fileInfo.Size() && start <= end {
					c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
					c.Header("Content-Length", fmt.Sprintf("%d", end-start+1))
					c.Status(http.StatusPartialContent)

					audioFile.Seek(start, 0)
					_, err := io.CopyN(c.Writer, audioFile, end-start+1)
					if err != nil {
						handler.Logger.CompleteWithErrorLog(loggerModel, err)
						return
					}

					handler.Logger.CompleteWithSuccessLog(loggerModel)
					return
				}
			}
		}
	}

	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, audioFile)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
}

func (handler *Handler) GetVideosHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetVideos",
		Description: "Fetching video files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)
	page := utils.ParseInt(c.DefaultQuery("page", "1"), c)
	pageSize := utils.ParseInt(c.DefaultQuery("page_size", "15"), c)

	pagination, err := handler.service.GetVideos(page, pageSize)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, pagination)
}

func (handler *Handler) StreamVideoHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "StreamVideo",
		Description: "Streaming video file",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	id := utils.ParseInt(c.Param("id"), c)

	file, err := handler.service.GetFileById(id)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	exists := handler.service.CheckFileExistsByPath(file.Path)
	if !exists {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("file not found on disk"))
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_FILE_NOT_FOUND")})
		return
	}

	videoFile, err := os.Open(file.Path)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer videoFile.Close()

	fileInfo, err := videoFile.Stat()
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "video/mp4")
	c.Header("Accept-Ranges", "bytes")
	c.Header("Cache-Control", "public, max-age=3600")

	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		// Parse Range header: "bytes=0-1048576"
		ranges := strings.Split(rangeHeader, "=")
		if len(ranges) == 2 && ranges[0] == "bytes" {
			byteRange := strings.Split(ranges[1], "-")
			if len(byteRange) == 2 {
				start, _ := strconv.ParseInt(byteRange[0], 10, 64)
				end, _ := strconv.ParseInt(byteRange[1], 10, 64)

				// Validação do range
				if start >= 0 && end < fileInfo.Size() && start <= end {
					c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileInfo.Size()))
					c.Header("Content-Length", fmt.Sprintf("%d", end-start+1))
					c.Status(http.StatusPartialContent)

					videoFile.Seek(start, 0)
					_, err := io.CopyN(c.Writer, videoFile, end-start+1)
					if err != nil {
						handler.Logger.CompleteWithErrorLog(loggerModel, err)
						return
					}

					handler.Logger.CompleteWithSuccessLog(loggerModel)
					return
				}
			}
		}
	}

	c.Header("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	c.Status(http.StatusOK)

	_, err = io.Copy(c.Writer, videoFile)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
}
