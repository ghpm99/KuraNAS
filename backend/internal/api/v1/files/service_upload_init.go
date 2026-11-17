package files

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var sha256HexPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

type UploadInitRequest struct {
	FileName       string
	Size           int64
	TargetFolderID int
	RelativePath   string
	Checksum       string
	ModifiedAt     *time.Time
	OnConflict     UploadConflictPolicy
}

type UploadInitResult struct {
	UploadID      string
	ChunkSize     int64
	ReceivedSize  int64
	Resumed       bool
	AlreadyExists bool
}

type validatedUploadInit struct {
	session         UploadSession
	destinationPath string
}

// InitUpload opens a resumable upload session, or resumes the equivalent open
// one. It answers AlreadyExists, without a session, when the destination holds
// a file with the declared checksum.
func (s *Service) InitUpload(request UploadInitRequest) (UploadInitResult, error) {
	validated, err := s.validateUploadInit(request)
	if err != nil {
		return UploadInitResult{}, err
	}

	existing := s.findExistingEntry(validated.destinationPath, true)
	if s.holdsDeclaredContent(validated, existing) {
		return UploadInitResult{ChunkSize: uploadChunkSize, ReceivedSize: validated.session.Size, AlreadyExists: true}, nil
	}
	if existing.exists && validated.session.OnConflict == UploadConflictFail {
		return UploadInitResult{}, newFileOperationError(
			http.StatusConflict,
			"ERROR_TARGET_ALREADY_EXISTS",
			fmt.Errorf("file already exists: %s", validated.destinationPath),
		)
	}

	if openSession, isOpen := findOpenUploadSession(validated.session); isOpen {
		return s.resumeUploadSession(openSession, validated.session)
	}
	return s.openUploadSession(validated.session)
}

func (s *Service) validateUploadInit(request UploadInitRequest) (validatedUploadInit, error) {
	fileName, err := validateItemName(request.FileName)
	if err != nil {
		return validatedUploadInit{}, err
	}
	if request.Size < 0 {
		return validatedUploadInit{}, newFileOperationError(http.StatusBadRequest, "ERROR_UPLOAD_SIZE_INVALID", fmt.Errorf("invalid size %d", request.Size))
	}
	policy, err := normalizeConflictPolicy(request.OnConflict)
	if err != nil {
		return validatedUploadInit{}, err
	}
	checksum, err := normalizeUploadChecksum(request.Checksum)
	if err != nil {
		return validatedUploadInit{}, err
	}
	relativeSegments, err := parseRelativeDirectorySegments(request.RelativePath)
	if err != nil {
		return validatedUploadInit{}, err
	}
	targetFolderPath, err := s.resolveUploadTargetFolder(request.TargetFolderID)
	if err != nil {
		return validatedUploadInit{}, err
	}

	relativePath := strings.Join(relativeSegments, "/")
	session := UploadSession{
		FileName:         fileName,
		Size:             request.Size,
		Checksum:         checksum,
		ModifiedAt:       request.ModifiedAt,
		OnConflict:       policy,
		TargetFolderPath: targetFolderPath,
		RelativePath:     relativePath,
	}
	destinationPath, err := resolveContainedPath(filepath.Join(session.destinationDirectory(), fileName))
	if err != nil {
		return validatedUploadInit{}, invalidPathOperationError(err)
	}
	return validatedUploadInit{session: session, destinationPath: destinationPath}, nil
}

func (s *Service) resolveUploadTargetFolder(targetFolderID int) (string, error) {
	var folderIDPointer *int
	if targetFolderID > 0 {
		folderIDPointer = &targetFolderID
	}
	targetFolderPath, err := s.resolveTargetFolder(folderIDPointer, "")
	if err != nil {
		return "", err
	}
	stat, err := os.Stat(targetFolderPath)
	if err != nil || !stat.IsDir() {
		return "", newFileOperationError(http.StatusBadRequest, "ERROR_TARGET_NOT_DIRECTORY", err)
	}
	return targetFolderPath, nil
}

func normalizeUploadChecksum(rawChecksum string) (string, error) {
	checksum := strings.ToLower(strings.TrimSpace(rawChecksum))
	if checksum == "" || sha256HexPattern.MatchString(checksum) {
		return checksum, nil
	}
	return "", newFileOperationError(http.StatusBadRequest, "ERROR_UPLOAD_CHECKSUM_INVALID", fmt.Errorf("invalid checksum %q", rawChecksum))
}

func parseRelativeDirectorySegments(relativePath string) ([]string, error) {
	normalizedPath := normalizePathSeparators(strings.TrimSpace(relativePath))
	if strings.HasPrefix(normalizedPath, "/") || hasDriveLetterPrefix(normalizedPath) || strings.ContainsRune(normalizedPath, 0) {
		return nil, invalidRelativePathError(relativePath)
	}
	var segments []string
	for _, segment := range strings.Split(normalizedPath, "/") {
		if segment == ".." {
			return nil, invalidRelativePathError(relativePath)
		}
		if segment == "" || segment == "." {
			continue
		}
		segments = append(segments, segment)
	}
	return segments, nil
}

func (s *Service) holdsDeclaredContent(validated validatedUploadInit, existing existingEntry) bool {
	if validated.session.Checksum == "" || !existing.exists || existing.isDirectory || existing.isTiered {
		return false
	}
	stat, err := os.Stat(validated.destinationPath)
	if err != nil || stat.Size() != validated.session.Size {
		return false
	}
	checksum, err := sha256HexOfFile(validated.destinationPath)
	return err == nil && checksum == validated.session.Checksum
}

func sha256HexOfFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (s *Service) resumeUploadSession(openSession UploadSession, requested UploadSession) (UploadInitResult, error) {
	unlock := lockUploadSession(openSession.UploadID)
	defer unlock()

	if err := s.ensureFreeSpaceFor(openSession.destinationDirectory(), requested.Size-openSession.ReceivedSize); err != nil {
		return UploadInitResult{}, err
	}
	openSession.OnConflict = requested.OnConflict
	openSession.ModifiedAt = requested.ModifiedAt
	if openSession.Checksum == "" {
		openSession.Checksum = requested.Checksum
	}
	openSession.LastActivityAt = time.Now()
	if err := saveUploadSession(openSession); err != nil {
		return UploadInitResult{}, uploadInternalError("resume upload session", err)
	}
	return UploadInitResult{
		UploadID:     openSession.UploadID,
		ChunkSize:    uploadChunkSize,
		ReceivedSize: openSession.ReceivedSize,
		Resumed:      true,
	}, nil
}

func (s *Service) openUploadSession(session UploadSession) (UploadInitResult, error) {
	if err := s.ensureFreeSpaceFor(session.destinationDirectory(), session.Size); err != nil {
		return UploadInitResult{}, err
	}
	uploadID, err := generateUploadID()
	if err != nil {
		return UploadInitResult{}, uploadInternalError("generate upload id", err)
	}
	now := time.Now()
	session.UploadID = uploadID
	session.CreatedAt = now
	session.LastActivityAt = now
	if err := createUploadSession(session); err != nil {
		return UploadInitResult{}, uploadInternalError("create upload session", err)
	}
	return UploadInitResult{UploadID: uploadID, ChunkSize: uploadChunkSize}, nil
}

func (s *Service) ensureFreeSpaceFor(destinationDirectory string, requiredBytes int64) error {
	availableBytes, err := s.availableBytesAtExistingAncestor(destinationDirectory)
	if err != nil {
		return uploadInternalError("read free space", err)
	}
	if requiredBytes > availableBytes {
		return newFileOperationError(
			http.StatusInsufficientStorage,
			"ERROR_UPLOAD_INSUFFICIENT_SPACE",
			fmt.Errorf("need %d bytes, %d available", requiredBytes, availableBytes),
		)
	}
	return nil
}

func (s *Service) availableBytesAtExistingAncestor(path string) (int64, error) {
	measure := s.AvailableBytesAt
	if measure == nil {
		measure = availableBytesAt
	}
	existingPath := path
	for {
		if _, err := os.Stat(existingPath); err == nil {
			return measure(existingPath)
		}
		parentPath := filepath.Dir(existingPath)
		if parentPath == existingPath {
			return measure(existingPath)
		}
		existingPath = parentPath
	}
}
