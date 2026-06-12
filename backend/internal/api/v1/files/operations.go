package files

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"mime/multipart"
	"nas-go/api/internal/roots"
	"nas-go/api/pkg/i18n"
	"nas-go/api/pkg/logger"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

	permanent := c.DefaultQuery("permanent", "false") == "true"
	if err := handler.service.DeleteFileFromDisk(req.ID, permanent); err != nil {
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

type FileOperationError struct {
	StatusCode int
	MessageKey string
	Err        error
}

func (e *FileOperationError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return e.MessageKey
}

type UploadFilesResult struct {
	Uploaded []string
	JobID    int
}

func newFileOperationError(statusCode int, messageKey string, err error) *FileOperationError {
	return &FileOperationError{
		StatusCode: statusCode,
		MessageKey: messageKey,
		Err:        err,
	}
}

func normalizePathSeparators(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
}

// resolvePathInRoots validates that a path (absolute, or client-relative)
// lands under some enabled storage root and returns its absolute clean form.
func resolvePathInRoots(inputPath string) (string, error) {
	candidate := normalizePathSeparators(strings.TrimSpace(inputPath))
	return roots.ResolveAbsolute(candidate)
}

// resolveTargetFolder resolves a folder from ID or creates from path.
// If folderID is non-nil and > 0, looks up by ID and validates it's a directory.
// If folderID is nil/0 and path is empty, returns entry point (root).
// If folderID is nil/0 and path is non-empty, creates the folder chain and returns the path.
func (s *Service) resolveTargetFolder(folderID *int, relativePath string) (string, error) {
	if folderID != nil && *folderID > 0 {
		folder, err := s.GetFileById(*folderID)
		if err != nil {
			if err == sql.ErrNoRows {
				return "", newFileOperationError(http.StatusNotFound, "ERROR_FOLDER_NOT_FOUND", err)
			}
			return "", newFileOperationError(http.StatusInternalServerError, "ERROR_FILE_NOT_FOUND", err)
		}
		if folder.Type != Directory {
			return "", newFileOperationError(http.StatusBadRequest, "ERROR_TARGET_NOT_DIRECTORY", fmt.Errorf("target is not a directory"))
		}
		resolved, err := resolvePathInRoots(folder.Path)
		if err != nil {
			return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
		}
		return resolved, nil
	}

	trimmedPath := strings.TrimSpace(relativePath)
	if trimmedPath == "" {
		return resolvePathInRoots("")
	}

	resolved, err := resolvePathInRoots(trimmedPath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	if err := os.MkdirAll(resolved, 0755); err != nil {
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_CREATE_FOLDER_FAILED", err)
	}

	return resolved, nil
}

func (s *Service) UploadFiles(targetFolderID int, files []*multipart.FileHeader) (UploadFilesResult, error) {
	var folderIDPtr *int
	if targetFolderID > 0 {
		folderIDPtr = &targetFolderID
	}

	resolvedTargetPath, err := s.resolveTargetFolder(folderIDPtr, "")
	if err != nil {
		return UploadFilesResult{}, err
	}

	stat, err := os.Stat(resolvedTargetPath)
	if err != nil || !stat.IsDir() {
		return UploadFilesResult{}, newFileOperationError(http.StatusBadRequest, "ERROR_TARGET_NOT_DIRECTORY", err)
	}

	if len(files) == 0 {
		return UploadFilesResult{}, newFileOperationError(http.StatusBadRequest, "ERROR_NO_FILES_UPLOADED", fmt.Errorf("empty upload payload"))
	}

	uploaded := make([]string, 0, len(files))
	for _, fileHeader := range files {
		fileName := filepath.Base(fileHeader.Filename)
		if fileName == "." || fileName == string(filepath.Separator) || strings.TrimSpace(fileName) == "" {
			return UploadFilesResult{}, newFileOperationError(http.StatusBadRequest, "ERROR_FILE_NAME_INVALID", fmt.Errorf("invalid file name"))
		}

		destinationPath := filepath.Join(resolvedTargetPath, fileName)
		destinationPath, err = resolvePathInRoots(destinationPath)
		if err != nil {
			return UploadFilesResult{}, newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
		}

		if _, statErr := os.Stat(destinationPath); statErr == nil {
			return UploadFilesResult{}, newFileOperationError(
				http.StatusConflict,
				"ERROR_TARGET_ALREADY_EXISTS",
				fmt.Errorf("file already exists: %s", destinationPath),
			)
		}

		if saveErr := saveUploadedFile(fileHeader, destinationPath); saveErr != nil {
			return UploadFilesResult{}, newFileOperationError(http.StatusInternalServerError, "ERROR_UPLOAD_FAILED", saveErr)
		}

		if syncErr := s.syncPathRow(destinationPath); syncErr != nil {
			s.logSyncFailure("UploadFiles", destinationPath, syncErr)
		}

		uploaded = append(uploaded, destinationPath)
	}

	jobID, err := s.CreateUploadProcessJob(uploaded)
	if err != nil {
		return UploadFilesResult{}, newFileOperationError(http.StatusInternalServerError, "ERROR_UPLOAD_JOB_CREATE", err)
	}

	return UploadFilesResult{
		Uploaded: uploaded,
		JobID:    jobID,
	}, nil
}

func saveUploadedFile(fileHeader *multipart.FileHeader, destinationPath string) error {
	source, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

func (s *Service) CreateFolder(parentID *int, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_FOLDER_NAME_REQUIRED", fmt.Errorf("empty folder name"))
	}
	if name != filepath.Base(name) {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_FOLDER_NAME_INVALID", fmt.Errorf("invalid folder name"))
	}

	resolvedParentPath, err := s.resolveTargetFolder(parentID, "")
	if err != nil {
		return "", err
	}

	createdPath := filepath.Join(resolvedParentPath, name)
	createdPath, err = resolvePathInRoots(createdPath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	if err := os.Mkdir(createdPath, 0755); err != nil {
		if os.IsExist(err) {
			return "", newFileOperationError(http.StatusConflict, "ERROR_FOLDER_ALREADY_EXISTS", err)
		}
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_CREATE_FOLDER_FAILED", err)
	}

	if err := s.syncPathRow(createdPath); err != nil {
		s.logSyncFailure("CreateFolder", createdPath, err)
	}

	s.ScanDirTask(resolvedParentPath)
	return createdPath, nil
}

func (s *Service) MoveFile(sourceID int, destinationFolderID *int, destinationPath string) (string, error) {
	if sourceID <= 0 {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_ID", fmt.Errorf("invalid source id"))
	}

	sourceFile, err := s.GetFileById(sourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
		}
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_SOURCE_NOT_FOUND", err)
	}

	resolvedSourcePath, err := resolvePathInRoots(sourceFile.Path)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	resolvedDestDir, err := s.resolveTargetFolder(destinationFolderID, destinationPath)
	if err != nil {
		return "", err
	}

	resolvedDestPath := filepath.Join(resolvedDestDir, sourceFile.Name)
	resolvedDestPath, err = resolvePathInRoots(resolvedDestPath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	sourceInfo, err := os.Stat(resolvedSourcePath)
	if err != nil {
		return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
	}

	if _, err := os.Stat(resolvedDestPath); err == nil {
		return "", newFileOperationError(
			http.StatusConflict,
			"ERROR_TARGET_ALREADY_EXISTS",
			fmt.Errorf("destination exists: %s", resolvedDestPath),
		)
	}

	if sourceInfo.IsDir() {
		relPath, relErr := filepath.Rel(resolvedSourcePath, resolvedDestPath)
		if relErr == nil && relPath != "." && relPath != ".." && !strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
			return "", newFileOperationError(
				http.StatusBadRequest,
				"ERROR_CANNOT_MOVE_INTO_ITSELF",
				fmt.Errorf("cannot move directory into itself"),
			)
		}
	}

	if err := os.Rename(resolvedSourcePath, resolvedDestPath); err != nil {
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_MOVE_FAILED", err)
	}

	if err := s.syncMovedRows(sourceFile, resolvedDestPath); err != nil {
		s.logSyncFailure("MoveFile", resolvedDestPath, err)
	}

	s.ScanDirTask(filepath.Dir(resolvedSourcePath))
	if filepath.Dir(resolvedDestPath) != filepath.Dir(resolvedSourcePath) {
		s.ScanDirTask(filepath.Dir(resolvedDestPath))
	}

	return resolvedDestPath, nil
}

// DeleteFileFromDisk removes the entry from the visible tree. By default the
// bytes survive: the path is moved into the trash bin and stays restorable.
// permanent=true keeps the old destructive behavior (os.RemoveAll).
func (s *Service) DeleteFileFromDisk(id int, permanent bool) error {
	if id <= 0 {
		return newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_ID", fmt.Errorf("invalid file id"))
	}

	file, err := s.GetFileById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
		}
		return newFileOperationError(http.StatusInternalServerError, "ERROR_SOURCE_NOT_FOUND", err)
	}

	resolvedPath, err := resolvePathInRoots(file.Path)
	if err != nil {
		return newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	entryPoint, err := resolvePathInRoots("")
	if err != nil {
		return newFileOperationError(http.StatusInternalServerError, "ERROR_DELETE_FAILED", err)
	}
	if resolvedPath == entryPoint {
		return newFileOperationError(
			http.StatusBadRequest,
			"ERROR_DELETE_ENTRYPOINT_FORBIDDEN",
			fmt.Errorf("cannot remove entry point"),
		)
	}

	if _, err := os.Stat(resolvedPath); err != nil {
		return newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
	}

	if permanent {
		if err := os.RemoveAll(resolvedPath); err != nil {
			return newFileOperationError(http.StatusInternalServerError, "ERROR_DELETE_FAILED", err)
		}
	} else {
		if s.TrashBin == nil {
			// Refuse to fall back to destruction: a missing trash bin is a
			// wiring bug, not a license to delete bytes forever.
			return newFileOperationError(
				http.StatusInternalServerError,
				"ERROR_DELETE_FAILED",
				fmt.Errorf("trash bin not configured"),
			)
		}
		if err := s.TrashBin.MoveToTrash(resolvedPath, file.Size); err != nil {
			return newFileOperationError(http.StatusInternalServerError, "ERROR_DELETE_FAILED", err)
		}
	}

	if err := s.syncDeletedRows(file.Path); err != nil {
		s.logSyncFailure("DeleteFileFromDisk", file.Path, err)
	}

	s.ScanDirTask(filepath.Dir(resolvedPath))
	return nil
}

func (s *Service) RenameFile(id int, newName string) (string, error) {
	if id <= 0 {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_ID", fmt.Errorf("invalid file id"))
	}

	trimmedName := strings.TrimSpace(newName)
	if trimmedName == "" {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_RENAME_NAME_REQUIRED", fmt.Errorf("new name required"))
	}
	if trimmedName != filepath.Base(trimmedName) {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_RENAME_NAME_INVALID", fmt.Errorf("invalid new name"))
	}

	file, err := s.GetFileById(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
		}
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_SOURCE_NOT_FOUND", err)
	}

	resolvedSourcePath, err := resolvePathInRoots(file.Path)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	if _, err := os.Stat(resolvedSourcePath); err != nil {
		return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
	}

	destPath, err := resolvePathInRoots(filepath.Join(filepath.Dir(resolvedSourcePath), trimmedName))
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}
	if _, err := os.Stat(destPath); err == nil {
		return "", newFileOperationError(
			http.StatusConflict,
			"ERROR_TARGET_ALREADY_EXISTS",
			fmt.Errorf("destination exists: %s", destPath),
		)
	}

	if err := os.Rename(resolvedSourcePath, destPath); err != nil {
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_RENAME_FAILED", err)
	}

	if err := s.syncMovedRows(file, destPath); err != nil {
		s.logSyncFailure("RenameFile", destPath, err)
	}

	s.ScanDirTask(filepath.Dir(resolvedSourcePath))
	return destPath, nil
}

func (s *Service) CopyFile(sourceID int, destinationFolderID *int, destinationPath string, newName string) (string, error) {
	if sourceID <= 0 {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_ID", fmt.Errorf("invalid source id"))
	}

	sourceFile, err := s.GetFileById(sourceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
		}
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_SOURCE_NOT_FOUND", err)
	}

	resolvedSourcePath, err := resolvePathInRoots(sourceFile.Path)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	resolvedDestDir, err := s.resolveTargetFolder(destinationFolderID, destinationPath)
	if err != nil {
		return "", err
	}

	fileName := strings.TrimSpace(newName)
	if fileName == "" {
		fileName = sourceFile.Name
	}

	resolvedDestPath := filepath.Join(resolvedDestDir, fileName)
	resolvedDestPath, err = resolvePathInRoots(resolvedDestPath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	sourceInfo, err := os.Stat(resolvedSourcePath)
	if err != nil {
		return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
	}
	if _, err := os.Stat(resolvedDestPath); err == nil {
		return "", newFileOperationError(
			http.StatusConflict,
			"ERROR_TARGET_ALREADY_EXISTS",
			fmt.Errorf("destination exists: %s", resolvedDestPath),
		)
	}

	if sourceInfo.IsDir() {
		relPath, relErr := filepath.Rel(resolvedSourcePath, resolvedDestPath)
		if relErr == nil && relPath != "." && relPath != ".." && !strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
			return "", newFileOperationError(
				http.StatusBadRequest,
				"ERROR_CANNOT_COPY_INTO_ITSELF",
				fmt.Errorf("cannot copy directory into itself"),
			)
		}
	}

	if err := copyPathRecursive(resolvedSourcePath, resolvedDestPath); err != nil {
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_COPY_FAILED", err)
	}

	// Only the copied root row is inserted synchronously; the contents of a
	// copied directory (and metadata/checksum/thumbnail) come via the pipeline.
	if err := s.syncPathRow(resolvedDestPath); err != nil {
		s.logSyncFailure("CopyFile", resolvedDestPath, err)
	}

	s.ScanDirTask(filepath.Dir(resolvedDestPath))
	return resolvedDestPath, nil
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

	_, err = io.Copy(destinationFile, sourceFile)
	return err
}
