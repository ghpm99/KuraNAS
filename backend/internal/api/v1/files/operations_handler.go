package files

import (
	"errors"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type createFolderRequest struct {
	ParentID *int   `json:"parent_id"`
	Name     string `json:"name"`
}

type moveFileRequest struct {
	SourceID            int    `json:"source_id"`
	DestinationFolderID *int   `json:"destination_folder_id"`
	DestinationPath     string `json:"destination_path"`
}

type deleteFileRequest struct {
	ID int `json:"id"`
}

type renameFileRequest struct {
	ID      int    `json:"id"`
	NewName string `json:"new_name"`
}

type copyFileRequest struct {
	SourceID            int    `json:"source_id"`
	DestinationFolderID *int   `json:"destination_folder_id"`
	DestinationPath     string `json:"destination_path"`
	NewName             string `json:"new_name"`
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

	var targetFolderID int
	if raw := c.PostForm("target_folder_id"); raw != "" {
		parsed, parseErr := strconv.Atoi(raw)
		if parseErr != nil {
			handler.Logger.CompleteWithErrorLog(loggerModel, parseErr)
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_ID")})
			return
		}
		targetFolderID = parsed
	}

	result, err := handler.service.UploadFiles(targetFolderID, form.File["files"])
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

	createdPath, err := handler.service.CreateFolder(req.ParentID, req.Name)
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

	destinationPath, err := handler.service.MoveFile(req.SourceID, req.DestinationFolderID, req.DestinationPath)
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_MOVE_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_MOVE_SUCCESS"), "path": destinationPath})
}

func (handler *Handler) DeleteFileHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "DeleteFile",
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

	if err := handler.service.DeleteFileFromDisk(req.ID); err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_DELETE_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_DELETE_SUCCESS")})
}

func (handler *Handler) RenameFileHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "RenameFile",
		Description: "Renaming file or folder from disk",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var req renameFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	destinationPath, err := handler.service.RenameFile(req.ID, req.NewName)
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_RENAME_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_RENAME_SUCCESS"), "path": destinationPath})
}

func (handler *Handler) CopyFileHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "CopyFile",
		Description: "Copying file or folder",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	var req copyFileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_REQUEST")})
		return
	}

	destinationPath, err := handler.service.CopyFile(req.SourceID, req.DestinationFolderID, req.DestinationPath, req.NewName)
	if err != nil {
		handler.respondFileOperationError(c, loggerModel, err, "ERROR_COPY_FAILED")
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_COPY_SUCCESS"), "path": destinationPath})
}
