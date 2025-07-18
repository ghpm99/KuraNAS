package files

import (
	"bytes"
	"fmt"
	"image/png"
	"mime"
	"strings"
	"time"

	"nas-go/api/internal/config"
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
	fmt.Println("📁 Recebendo dados para processamento:", data)

	if data == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("data is required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "data is required"})
		return
	}
	loggerModel.SetExtraData(logger.LogExtraData{
		Data: data,
	})
	handler.service.ScanFilesTask(data)
	handler.Logger.CompleteWithSuccessLog(loggerModel)
}

func (handler *Handler) GetFilesThreeHandler(c *gin.Context) {

	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "GetFilesThree",
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

	loggerModel.SetExtraData(logger.LogExtraData{
		Data: id,
	})

	file, err := handler.service.GetFileById(id)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error1": err.Error()})
		return
	}

	thumbnail, err := handler.service.GetFileThumbnail(file, 320)

	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error2": err.Error()})
		return
	}

	var buf bytes.Buffer
	err = png.Encode(&buf, thumbnail)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error3": err.Error()})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.Data(http.StatusOK, "image/png", buf.Bytes())
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
		c.JSON(http.StatusInternalServerError, gin.H{"error1": err.Error()})
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
