package files

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"nas-go/api/pkg/applog"
	"nas-go/api/pkg/i18n"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

type UploadConflictPolicy string

const (
	UploadConflictFail    UploadConflictPolicy = "fail"
	UploadConflictSkip    UploadConflictPolicy = "skip"
	UploadConflictReplace UploadConflictPolicy = "replace"
	UploadConflictRename  UploadConflictPolicy = "rename"
)

const (
	UploadStatusUploaded = "uploaded"
	UploadStatusSkipped  = "skipped"
	UploadStatusReplaced = "replaced"
	UploadStatusRenamed  = "renamed"
	UploadStatusFailed   = "failed"
)

const maxRenameAttempts = 10000

type UploadOptions struct {
	OnConflict    UploadConflictPolicy
	RelativePaths []string
}

type UploadFileResult struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type UploadFilesResult struct {
	Uploaded []string
	JobID    int
	Files    []UploadFileResult
}

type pendingUpload struct {
	header            *multipart.FileHeader
	fileName          string
	directorySegments []string
}

type existingEntry struct {
	exists      bool
	isDirectory bool
	isTiered    bool
}

func (s *Service) UploadFiles(targetFolderID int, files []*multipart.FileHeader) (UploadFilesResult, error) {
	return s.UploadFilesWithOptions(targetFolderID, files, UploadOptions{})
}

func (s *Service) UploadFilesWithOptions(targetFolderID int, files []*multipart.FileHeader, options UploadOptions) (UploadFilesResult, error) {
	policy, err := normalizeConflictPolicy(options.OnConflict)
	if err != nil {
		return UploadFilesResult{}, err
	}

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

	pendingUploads, err := buildPendingUploads(files, options.RelativePaths)
	if err != nil {
		return UploadFilesResult{}, err
	}

	uploaded := make([]string, 0, len(files))
	fileResults := make([]UploadFileResult, 0, len(files))
	for _, pending := range pendingUploads {
		fileResult, storeErr := s.storeUploadedFile(resolvedTargetPath, pending, policy)
		if storeErr != nil && policy == UploadConflictFail {
			return UploadFilesResult{}, storeErr
		}
		if storeErr != nil {
			fileResult = failedUploadResult(pending.fileName, storeErr)
		}
		if isWrittenStatus(fileResult.Status) {
			uploaded = append(uploaded, fileResult.Path)
		}
		fileResults = append(fileResults, fileResult)
	}

	jobID, err := s.createUploadJobWhenNeeded(uploaded)
	if err != nil {
		return UploadFilesResult{}, err
	}

	return UploadFilesResult{Uploaded: uploaded, JobID: jobID, Files: fileResults}, nil
}

func (s *Service) createUploadJobWhenNeeded(uploaded []string) (int, error) {
	if len(uploaded) == 0 {
		return 0, nil
	}
	jobID, err := s.CreateUploadProcessJob(uploaded)
	if err != nil {
		return 0, newFileOperationError(http.StatusInternalServerError, "ERROR_UPLOAD_JOB_CREATE", err)
	}
	return jobID, nil
}

func isWrittenStatus(status string) bool {
	return status == UploadStatusUploaded || status == UploadStatusReplaced || status == UploadStatusRenamed
}

func failedUploadResult(fileName string, storeErr error) UploadFileResult {
	messageKey := "ERROR_UPLOAD_FAILED"
	var operationErr *FileOperationError
	if errors.As(storeErr, &operationErr) {
		messageKey = operationErr.MessageKey
	}
	applog.ErrorWithStack("upload file failed", storeErr, "file", fileName)
	return UploadFileResult{Name: fileName, Status: UploadStatusFailed, Error: i18n.GetMessage(messageKey)}
}

func normalizeConflictPolicy(policy UploadConflictPolicy) (UploadConflictPolicy, error) {
	switch policy {
	case "":
		return UploadConflictFail, nil
	case UploadConflictFail, UploadConflictSkip, UploadConflictReplace, UploadConflictRename:
		return policy, nil
	}
	return "", newFileOperationError(http.StatusBadRequest, "ERROR_UPLOAD_CONFLICT_POLICY_INVALID", fmt.Errorf("invalid conflict policy %q", policy))
}

func buildPendingUploads(files []*multipart.FileHeader, relativePaths []string) ([]pendingUpload, error) {
	if len(relativePaths) != 0 && len(relativePaths) != len(files) {
		return nil, newFileOperationError(
			http.StatusBadRequest,
			"ERROR_UPLOAD_RELATIVE_PATHS_MISMATCH",
			fmt.Errorf("%d relative paths for %d files", len(relativePaths), len(files)),
		)
	}

	pendingUploads := make([]pendingUpload, 0, len(files))
	for index, header := range files {
		fileName := filepath.Base(normalizePathSeparators(header.Filename))
		if fileName == "." || fileName == string(filepath.Separator) || strings.TrimSpace(fileName) == "" {
			return nil, newFileOperationError(http.StatusBadRequest, "ERROR_FILE_NAME_INVALID", fmt.Errorf("invalid file name"))
		}

		var directorySegmentss []string
		if len(relativePaths) != 0 {
			segments, err := parseUploadDirectorySegments(relativePaths[index])
			if err != nil {
				return nil, err
			}
			directorySegmentss = segments
		}
		pendingUploads = append(pendingUploads, pendingUpload{header: header, fileName: fileName, directorySegments: directorySegmentss})
	}
	return pendingUploads, nil
}

func parseUploadDirectorySegments(relativePath string) ([]string, error) {
	normalized := normalizePathSeparators(strings.TrimSpace(relativePath))
	if normalized == "" {
		return nil, nil
	}
	if strings.HasPrefix(normalized, "/") || hasDriveLetterPrefix(normalized) || strings.ContainsRune(normalized, 0) {
		return nil, invalidRelativePathError(relativePath)
	}

	parts := strings.Split(normalized, "/")
	directoryParts := parts[:len(parts)-1]
	segments := make([]string, 0, len(directoryParts))
	for _, part := range directoryParts {
		if part == ".." {
			return nil, invalidRelativePathError(relativePath)
		}
		if part == "" || part == "." {
			continue
		}
		segments = append(segments, part)
	}
	return segments, nil
}

func hasDriveLetterPrefix(path string) bool {
	return len(path) >= 2 && path[1] == ':' && unicode.IsLetter(rune(path[0]))
}

func invalidRelativePathError(relativePath string) *FileOperationError {
	return newFileOperationError(http.StatusBadRequest, "ERROR_UPLOAD_RELATIVE_PATH_INVALID", fmt.Errorf("invalid relative path %q", relativePath))
}

func (s *Service) storeUploadedFile(targetPath string, pending pendingUpload, policy UploadConflictPolicy) (UploadFileResult, error) {
	directoryPath, err := s.ensureUploadDirectory(targetPath, pending.directorySegments)
	if err != nil {
		return UploadFileResult{}, err
	}

	destinationPath, err := resolvePathInRoots(filepath.Join(directoryPath, pending.fileName))
	if err != nil {
		return UploadFileResult{}, newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
	}

	existing := s.findExistingEntry(destinationPath, policy != UploadConflictFail)
	if !existing.exists {
		return s.writeUpload(pending, destinationPath, UploadStatusUploaded)
	}

	switch policy {
	case UploadConflictSkip:
		return UploadFileResult{Name: pending.fileName, Path: destinationPath, Status: UploadStatusSkipped}, nil
	case UploadConflictReplace:
		return s.replaceExisting(pending, destinationPath, existing)
	case UploadConflictRename:
		return s.writeUploadUnderFreeName(pending, destinationPath)
	}
	return UploadFileResult{}, newFileOperationError(
		http.StatusConflict,
		"ERROR_TARGET_ALREADY_EXISTS",
		fmt.Errorf("file already exists: %s", destinationPath),
	)
}

func (s *Service) replaceExisting(pending pendingUpload, destinationPath string, existing existingEntry) (UploadFileResult, error) {
	if existing.isDirectory {
		return UploadFileResult{}, newFileOperationError(
			http.StatusConflict,
			"ERROR_TARGET_ALREADY_EXISTS",
			fmt.Errorf("cannot replace directory: %s", destinationPath),
		)
	}
	if existing.isTiered {
		return UploadFileResult{}, newFileOperationError(
			http.StatusConflict,
			"ERROR_UPLOAD_REPLACE_COLD_UNSUPPORTED",
			fmt.Errorf("cannot replace tiered file: %s", destinationPath),
		)
	}
	return s.writeUpload(pending, destinationPath, UploadStatusReplaced)
}

func (s *Service) writeUploadUnderFreeName(pending pendingUpload, occupiedPath string) (UploadFileResult, error) {
	directoryPath := filepath.Dir(occupiedPath)
	extension := filepath.Ext(pending.fileName)
	baseName := strings.TrimSuffix(pending.fileName, extension)

	for attempt := 2; attempt <= maxRenameAttempts; attempt++ {
		candidatePath := filepath.Join(directoryPath, fmt.Sprintf("%s (%d)%s", baseName, attempt, extension))
		if s.findExistingEntry(candidatePath, true).exists {
			continue
		}
		return s.writeUpload(pending, candidatePath, UploadStatusRenamed)
	}
	return UploadFileResult{}, newFileOperationError(
		http.StatusConflict,
		"ERROR_TARGET_ALREADY_EXISTS",
		fmt.Errorf("no free name for %s", occupiedPath),
	)
}

func (s *Service) writeUpload(pending pendingUpload, destinationPath string, status string) (UploadFileResult, error) {
	if err := saveUploadedFile(pending.header, destinationPath); err != nil {
		return UploadFileResult{}, newFileOperationError(http.StatusInternalServerError, "ERROR_UPLOAD_FAILED", err)
	}

	if syncErr := s.syncPathRow(destinationPath); syncErr != nil {
		s.logSyncFailure("UploadFiles", destinationPath, syncErr)
	}

	return UploadFileResult{Name: filepath.Base(destinationPath), Path: destinationPath, Status: status}, nil
}

func (s *Service) findExistingEntry(path string, shouldCheckTiered bool) existingEntry {
	if info, err := os.Stat(path); err == nil {
		return existingEntry{exists: true, isDirectory: info.IsDir()}
	}
	if !shouldCheckTiered {
		return existingEntry{}
	}
	tieredFile, err := s.GetFileByNameAndPath(filepath.Base(path), path)
	if err != nil || tieredFile.DeletedAt.HasValue || tieredFile.PhysicalPath == "" {
		return existingEntry{}
	}
	return existingEntry{exists: true, isTiered: true}
}

func (s *Service) ensureUploadDirectory(targetPath string, segments []string) (string, error) {
	currentPath := targetPath
	for _, segment := range segments {
		nextPath, err := resolvePathInRoots(filepath.Join(currentPath, segment))
		if err != nil {
			return "", newFileOperationError(http.StatusBadRequest, "ERROR_INVALID_PATH", err)
		}
		if info, statErr := os.Stat(nextPath); statErr == nil && info.IsDir() {
			currentPath = nextPath
			continue
		}
		createdPath, createErr := s.createFolderAt(currentPath, segment)
		if createErr != nil {
			return "", createErr
		}
		currentPath = createdPath
	}
	return currentPath, nil
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
