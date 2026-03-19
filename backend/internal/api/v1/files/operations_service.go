package files

import (
	"database/sql"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"nas-go/api/internal/config"
)

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

	candidate := normalizePathSeparators(strings.TrimSpace(inputPath))
	if candidate == "" {
		candidate = entryPointClean
	}

	if !filepath.IsAbs(candidate) || !strings.HasPrefix(filepath.Clean(candidate), entryPointClean) {
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
		resolved, err := resolvePathInEntryPoint(folder.Path)
		if err != nil {
			return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
		}
		return resolved, nil
	}

	trimmedPath := strings.TrimSpace(relativePath)
	if trimmedPath == "" {
		return resolvePathInEntryPoint("")
	}

	resolved, err := resolvePathInEntryPoint(trimmedPath)
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
		destinationPath, err = resolvePathInEntryPoint(destinationPath)
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
	createdPath, err = resolvePathInEntryPoint(createdPath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	if err := os.Mkdir(createdPath, 0755); err != nil {
		if os.IsExist(err) {
			return "", newFileOperationError(http.StatusConflict, "ERROR_FOLDER_ALREADY_EXISTS", err)
		}
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_CREATE_FOLDER_FAILED", err)
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

	resolvedSourcePath, err := resolvePathInEntryPoint(sourceFile.Path)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	resolvedDestDir, err := s.resolveTargetFolder(destinationFolderID, destinationPath)
	if err != nil {
		return "", err
	}

	resolvedDestPath := filepath.Join(resolvedDestDir, sourceFile.Name)
	resolvedDestPath, err = resolvePathInEntryPoint(resolvedDestPath)
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

	s.ScanDirTask(filepath.Dir(resolvedSourcePath))
	if filepath.Dir(resolvedDestPath) != filepath.Dir(resolvedSourcePath) {
		s.ScanDirTask(filepath.Dir(resolvedDestPath))
	}

	return resolvedDestPath, nil
}

func (s *Service) DeleteFileFromDisk(id int) error {
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

	resolvedPath, err := resolvePathInEntryPoint(file.Path)
	if err != nil {
		return newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	entryPoint, err := resolvePathInEntryPoint("")
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

	if err := os.RemoveAll(resolvedPath); err != nil {
		return newFileOperationError(http.StatusInternalServerError, "ERROR_DELETE_FAILED", err)
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

	resolvedSourcePath, err := resolvePathInEntryPoint(file.Path)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	if _, err := os.Stat(resolvedSourcePath); err != nil {
		return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
	}

	destPath, err := resolvePathInEntryPoint(filepath.Join(filepath.Dir(resolvedSourcePath), trimmedName))
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

	resolvedSourcePath, err := resolvePathInEntryPoint(sourceFile.Path)
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
	resolvedDestPath, err = resolvePathInEntryPoint(resolvedDestPath)
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
