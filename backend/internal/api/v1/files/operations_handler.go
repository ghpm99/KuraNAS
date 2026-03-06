package files

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"

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

func (handler *Handler) UploadFilesHandler(c *gin.Context) {
	loggerModel, _ := handler.Logger.CreateLog(logger.LoggerModel{
		Name:        "UploadFiles",
		Description: "Uploading files",
		Level:       logger.LogLevelInfo,
		Status:      logger.LogStatusPending,
		IPAddress:   c.ClientIP(),
	}, nil)

	targetPath, err := resolvePathInEntryPoint(c.PostForm("target_path"))
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}

	stat, err := os.Stat(targetPath)
	if err != nil || !stat.IsDir() {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_TARGET_NOT_DIRECTORY")})
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_NO_FILES_UPLOADED")})
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("empty upload payload"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_NO_FILES_UPLOADED")})
		return
	}

	uploaded := make([]string, 0, len(files))
	for _, fileHeader := range files {
		fileName := filepath.Base(fileHeader.Filename)
		if fileName == "." || fileName == string(filepath.Separator) || strings.TrimSpace(fileName) == "" {
			handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("invalid file name"))
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_FILE_NAME_INVALID")})
			return
		}

		destinationPath := filepath.Join(targetPath, fileName)
		destinationPath, err = resolvePathInEntryPoint(destinationPath)
		if err != nil {
			handler.Logger.CompleteWithErrorLog(loggerModel, err)
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
			return
		}

		if _, statErr := os.Stat(destinationPath); statErr == nil {
			handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("file already exists: %s", destinationPath))
			c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("ERROR_TARGET_ALREADY_EXISTS")})
			return
		}

		if saveErr := c.SaveUploadedFile(fileHeader, destinationPath); saveErr != nil {
			handler.Logger.CompleteWithErrorLog(loggerModel, saveErr)
			c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_UPLOAD_FAILED")})
			return
		}

		uploaded = append(uploaded, destinationPath)
	}

	jobID, err := handler.service.CreateUploadProcessJob(uploaded)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_UPLOAD_JOB_CREATE")})
		return
	}

	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusAccepted, gin.H{
		"message":  i18n.GetMessage("ACTION_UPLOAD_SUCCESS"),
		"uploaded": uploaded,
		"job_id":   jobID,
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

	if strings.TrimSpace(req.Name) == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("empty folder name"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_FOLDER_NAME_REQUIRED")})
		return
	}
	if req.Name != filepath.Base(req.Name) {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("invalid folder name"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_FOLDER_NAME_INVALID")})
		return
	}

	parentPath, err := resolvePathInEntryPoint(req.ParentPath)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}

	createdPath := filepath.Join(parentPath, req.Name)
	createdPath, err = resolvePathInEntryPoint(createdPath)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}

	if err := os.Mkdir(createdPath, 0755); err != nil {
		if os.IsExist(err) {
			handler.Logger.CompleteWithErrorLog(loggerModel, err)
			c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("ERROR_FOLDER_ALREADY_EXISTS")})
			return
		}
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_CREATE_FOLDER_FAILED")})
		return
	}

	handler.service.ScanDirTask(parentPath)
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

	if strings.TrimSpace(req.SourcePath) == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("source path required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_MOVE_SOURCE_REQUIRED")})
		return
	}
	if strings.TrimSpace(req.DestinationPath) == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("destination path required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_MOVE_TARGET_REQUIRED")})
		return
	}

	sourcePath, err := resolvePathInEntryPoint(req.SourcePath)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}
	destinationPath, err := resolvePathInEntryPoint(req.DestinationPath)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_SOURCE_NOT_FOUND")})
		return
	}

	if _, err := os.Stat(destinationPath); err == nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("destination exists: %s", destinationPath))
		c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("ERROR_TARGET_ALREADY_EXISTS")})
		return
	}

	if sourceInfo.IsDir() {
		relPath, relErr := filepath.Rel(sourcePath, destinationPath)
		if relErr == nil && relPath != "." && relPath != ".." && !strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
			handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("cannot move directory into itself"))
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_CANNOT_MOVE_INTO_ITSELF")})
			return
		}
	}

	if err := os.Rename(sourcePath, destinationPath); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_MOVE_FAILED")})
		return
	}

	handler.service.ScanDirTask(filepath.Dir(sourcePath))
	if filepath.Dir(destinationPath) != filepath.Dir(sourcePath) {
		handler.service.ScanDirTask(filepath.Dir(destinationPath))
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

	if strings.TrimSpace(req.Path) == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("path required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DELETE_PATH_REQUIRED")})
		return
	}

	resolvedPath, err := resolvePathInEntryPoint(req.Path)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}

	entryPoint, _ := resolvePathInEntryPoint("")
	if resolvedPath == entryPoint {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("cannot remove entry point"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_DELETE_ENTRYPOINT_FORBIDDEN")})
		return
	}

	if _, err := os.Stat(resolvedPath); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_SOURCE_NOT_FOUND")})
		return
	}

	if err := os.RemoveAll(resolvedPath); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_DELETE_FAILED")})
		return
	}

	handler.service.ScanDirTask(filepath.Dir(resolvedPath))
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

	if strings.TrimSpace(req.SourcePath) == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("source path required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_RENAME_SOURCE_REQUIRED")})
		return
	}
	newName := strings.TrimSpace(req.NewName)
	if newName == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("new name required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_RENAME_NAME_REQUIRED")})
		return
	}
	if newName != filepath.Base(newName) {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("invalid new name"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_RENAME_NAME_INVALID")})
		return
	}

	sourcePath, err := resolvePathInEntryPoint(req.SourcePath)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}
	if _, err := os.Stat(sourcePath); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_SOURCE_NOT_FOUND")})
		return
	}

	destinationPath, err := resolvePathInEntryPoint(filepath.Join(filepath.Dir(sourcePath), newName))
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}
	if _, err := os.Stat(destinationPath); err == nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("destination exists: %s", destinationPath))
		c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("ERROR_TARGET_ALREADY_EXISTS")})
		return
	}

	if err := os.Rename(sourcePath, destinationPath); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_RENAME_FAILED")})
		return
	}

	handler.service.ScanDirTask(filepath.Dir(sourcePath))
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
	if strings.TrimSpace(req.SourcePath) == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("source path required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_COPY_SOURCE_REQUIRED")})
		return
	}
	if strings.TrimSpace(req.DestinationPath) == "" {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("destination path required"))
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_COPY_TARGET_REQUIRED")})
		return
	}

	sourcePath, err := resolvePathInEntryPoint(req.SourcePath)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}
	destinationPath, err := resolvePathInEntryPoint(req.DestinationPath)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_INVALID_PATH")})
		return
	}

	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusNotFound, gin.H{"error": i18n.GetMessage("ERROR_SOURCE_NOT_FOUND")})
		return
	}
	if _, err := os.Stat(destinationPath); err == nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("destination exists: %s", destinationPath))
		c.JSON(http.StatusConflict, gin.H{"error": i18n.GetMessage("ERROR_TARGET_ALREADY_EXISTS")})
		return
	}

	if sourceInfo.IsDir() {
		relPath, relErr := filepath.Rel(sourcePath, destinationPath)
		if relErr == nil && relPath != "." && relPath != ".." && !strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
			handler.Logger.CompleteWithErrorLog(loggerModel, fmt.Errorf("cannot copy directory into itself"))
			c.JSON(http.StatusBadRequest, gin.H{"error": i18n.GetMessage("ERROR_CANNOT_COPY_INTO_ITSELF")})
			return
		}
	}

	if err := copyPathRecursive(sourcePath, destinationPath); err != nil {
		handler.Logger.CompleteWithErrorLog(loggerModel, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": i18n.GetMessage("ERROR_COPY_FAILED")})
		return
	}

	handler.service.ScanDirTask(filepath.Dir(destinationPath))
	handler.Logger.CompleteWithSuccessLog(loggerModel)
	c.JSON(http.StatusOK, gin.H{"message": i18n.GetMessage("ACTION_COPY_SUCCESS"), "path": destinationPath})
}

func copyPathRecursive(sourcePath string, destinationPath string) error {
	info, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if info.IsDir() {
		if err := os.MkdirAll(destinationPath, info.Mode()); err != nil {
			return err
		}
		entries, err := os.ReadDir(sourcePath)
		if err != nil {
			return err
		}
		for _, entry := range entries {
			nextSource := filepath.Join(sourcePath, entry.Name())
			nextDestination := filepath.Join(destinationPath, entry.Name())
			if err := copyPathRecursive(nextSource, nextDestination); err != nil {
				return err
			}
		}
		return nil
	}

	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	if _, err := io.Copy(destinationFile, sourceFile); err != nil {
		return err
	}
	return nil
}

func resolvePathInEntryPoint(inputPath string) (string, error) {
	entryPoint := strings.TrimSpace(config.AppConfig.EntryPoint)
	if entryPoint == "" {
		return "", fmt.Errorf("entry point not configured")
	}

	entryPointAbs, err := filepath.Abs(entryPoint)
	if err != nil {
		return "", err
	}
	entryPointClean := filepath.Clean(entryPointAbs)

	candidate := strings.TrimSpace(inputPath)
	if candidate == "" {
		candidate = entryPointClean
	}

	if !filepath.IsAbs(candidate) {
		candidate = filepath.Join(entryPointClean, candidate)
	}

	candidateAbs, err := filepath.Abs(candidate)
	if err != nil {
		return "", err
	}
	candidateClean := filepath.Clean(candidateAbs)

	relPath, err := filepath.Rel(entryPointClean, candidateClean)
	if err != nil {
		return "", err
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("path outside entrypoint")
	}

	return candidateClean, nil
}
