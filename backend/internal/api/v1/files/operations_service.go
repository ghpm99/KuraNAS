package files

import (
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

func (s *Service) UploadFiles(targetPath string, files []*multipart.FileHeader) (UploadFilesResult, error) {
	resolvedTargetPath, err := resolvePathInEntryPoint(targetPath)
	if err != nil {
		return UploadFilesResult{}, newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
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

func (s *Service) CreateFolder(parentPath string, name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_FOLDER_NAME_REQUIRED", fmt.Errorf("empty folder name"))
	}
	if name != filepath.Base(name) {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_FOLDER_NAME_INVALID", fmt.Errorf("invalid folder name"))
	}

	resolvedParentPath, err := resolvePathInEntryPoint(parentPath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
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

func (s *Service) MovePath(sourcePath string, destinationPath string) (string, error) {
	if strings.TrimSpace(sourcePath) == "" {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_MOVE_SOURCE_REQUIRED", fmt.Errorf("source path required"))
	}
	if strings.TrimSpace(destinationPath) == "" {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_MOVE_TARGET_REQUIRED", fmt.Errorf("destination path required"))
	}

	resolvedSourcePath, err := resolvePathInEntryPoint(sourcePath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}
	resolvedDestinationPath, err := resolvePathInEntryPoint(destinationPath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	sourceInfo, err := os.Stat(resolvedSourcePath)
	if err != nil {
		return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
	}

	if _, err := os.Stat(resolvedDestinationPath); err == nil {
		return "", newFileOperationError(
			http.StatusConflict,
			"ERROR_TARGET_ALREADY_EXISTS",
			fmt.Errorf("destination exists: %s", resolvedDestinationPath),
		)
	}

	if sourceInfo.IsDir() {
		relPath, relErr := filepath.Rel(resolvedSourcePath, resolvedDestinationPath)
		if relErr == nil && relPath != "." && relPath != ".." && !strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
			return "", newFileOperationError(
				http.StatusBadRequest,
				"ERROR_CANNOT_MOVE_INTO_ITSELF",
				fmt.Errorf("cannot move directory into itself"),
			)
		}
	}

	if err := os.Rename(resolvedSourcePath, resolvedDestinationPath); err != nil {
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_MOVE_FAILED", err)
	}

	s.ScanDirTask(filepath.Dir(resolvedSourcePath))
	if filepath.Dir(resolvedDestinationPath) != filepath.Dir(resolvedSourcePath) {
		s.ScanDirTask(filepath.Dir(resolvedDestinationPath))
	}

	return resolvedDestinationPath, nil
}

func (s *Service) DeletePath(path string) error {
	if strings.TrimSpace(path) == "" {
		return newFileOperationError(http.StatusBadRequest, "ERROR_DELETE_PATH_REQUIRED", fmt.Errorf("path required"))
	}

	resolvedPath, err := resolvePathInEntryPoint(path)
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

func (s *Service) RenamePath(sourcePath string, newName string) (string, error) {
	if strings.TrimSpace(sourcePath) == "" {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_RENAME_SOURCE_REQUIRED", fmt.Errorf("source path required"))
	}

	trimmedName := strings.TrimSpace(newName)
	if trimmedName == "" {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_RENAME_NAME_REQUIRED", fmt.Errorf("new name required"))
	}
	if trimmedName != filepath.Base(trimmedName) {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_RENAME_NAME_INVALID", fmt.Errorf("invalid new name"))
	}

	resolvedSourcePath, err := resolvePathInEntryPoint(sourcePath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	if _, err := os.Stat(resolvedSourcePath); err != nil {
		return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
	}

	destinationPath, err := resolvePathInEntryPoint(filepath.Join(filepath.Dir(resolvedSourcePath), trimmedName))
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}
	if _, err := os.Stat(destinationPath); err == nil {
		return "", newFileOperationError(
			http.StatusConflict,
			"ERROR_TARGET_ALREADY_EXISTS",
			fmt.Errorf("destination exists: %s", destinationPath),
		)
	}

	if err := os.Rename(resolvedSourcePath, destinationPath); err != nil {
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_RENAME_FAILED", err)
	}

	s.ScanDirTask(filepath.Dir(resolvedSourcePath))
	return destinationPath, nil
}

func (s *Service) CopyPath(sourcePath string, destinationPath string) (string, error) {
	if strings.TrimSpace(sourcePath) == "" {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_COPY_SOURCE_REQUIRED", fmt.Errorf("source path required"))
	}
	if strings.TrimSpace(destinationPath) == "" {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_COPY_TARGET_REQUIRED", fmt.Errorf("destination path required"))
	}

	resolvedSourcePath, err := resolvePathInEntryPoint(sourcePath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}
	resolvedDestinationPath, err := resolvePathInEntryPoint(destinationPath)
	if err != nil {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	sourceInfo, err := os.Stat(resolvedSourcePath)
	if err != nil {
		return "", newFileOperationError(http.StatusNotFound, "ERROR_SOURCE_NOT_FOUND", err)
	}
	if _, err := os.Stat(resolvedDestinationPath); err == nil {
		return "", newFileOperationError(
			http.StatusConflict,
			"ERROR_TARGET_ALREADY_EXISTS",
			fmt.Errorf("destination exists: %s", resolvedDestinationPath),
		)
	}

	if sourceInfo.IsDir() {
		relPath, relErr := filepath.Rel(resolvedSourcePath, resolvedDestinationPath)
		if relErr == nil && relPath != "." && relPath != ".." && !strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
			return "", newFileOperationError(
				http.StatusBadRequest,
				"ERROR_CANNOT_COPY_INTO_ITSELF",
				fmt.Errorf("cannot copy directory into itself"),
			)
		}
	}

	if err := copyPathRecursive(resolvedSourcePath, resolvedDestinationPath); err != nil {
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_COPY_FAILED", err)
	}

	s.ScanDirTask(filepath.Dir(resolvedDestinationPath))
	return resolvedDestinationPath, nil
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
