package files

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
)

type createFolderRequest struct {
	ParentPath string `json:"parent_path"`
	Name       string `json:"name"`
}

type moveFileRequest struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
}

type deleteFileRequest struct {
	Path string `json:"path"`
}

type renamePathRequest struct {
	SourcePath string `json:"source_path"`
	NewName    string `json:"new_name"`
}

type copyPathRequest struct {
	SourcePath      string `json:"source_path"`
	DestinationPath string `json:"destination_path"`
}

func (handler *Handler) respondFileOperationError(c *gin.Context, loggerModel logger.LoggerModel, err error, fallbackKey string) {
	var operationErr *FileOperationError
	if errors.As(err, &operationErr) {
		handler.Logger.CompleteWithErrorLog(loggerModel, operationErr)
		c.JSON(operationErr.StatusCode, gin.H{"error": i18n.GetMessage(operationErr.MessageKey)})
		return
	}

	handler.Logger.CompleteWithErrorLog(loggerModel, err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage(fallbackKey)})
}

func (handler *Handler) UploadFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "UploadFiles",
		Description: "Uploading files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	form, err := c.MultipartForm()
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_NO_FILES_UPLOADED")})
		return
	}

	result, err := handler.service.UploadFiles(c.PostForm("target_path"), form.File["files"])
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_UPLOAD_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusAccepted, gin.H{
		"message":  i18n.GetMessage("ACTION_UPLOAD_SUCCESS"),
		"uploaded": result.Uploaded,
		"job_id":   result.JobID,
	})
}

func (handler *Handler) CreateFolderHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "CreateFolder",
		Description: "Creating folder",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var req createFolderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	createdPath, err := handler.service.CreateFolder(req.ParentPath, req.Name)
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_CREATE_FOLDER_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusCreated, gin.H{"message": i18n.GetMessage("ACTION_CREATE_FOLDER_SUCCESS"), "path": createdPath})
}

func (handler *Handler) MoveFileHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "MoveFile",
		Description: "Moving file or folder",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var req moveFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	destinationPath, err := handler.service.MovePath(req.SourcePath, req.DestinationPath)
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_MOVE_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_MOVE_SUCCESS"), "path": destinationPath})
}

func (handler *Handler) DeletePathHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "DeletePath",
		Description: "Deleting file or folder from disk",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var req deleteFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	if err := handler.service.DeletePath(req.Path); err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_DELETE_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_DELETE_SUCCESS")})
}

func (handler *Handler) RenamePathHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "RenamePath",
		Description: "Renaming file or folder from disk",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var req renamePathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	destinationPath, err := handler.service.RenamePath(req.SourcePath, req.NewName)
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_RENAME_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_RENAME_SUCCESS"), "path": destinationPath})
}

func (handler *Handler) CopyPathHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "CopyPath",
		Description: "Copying file or folder",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var req copyPathRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	destinationPath, err := handler.service.CopyPath(req.SourcePath, req.DestinationPath)
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_COPY_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_COPY_SUCCESS"), "path": destinationPath})
}
