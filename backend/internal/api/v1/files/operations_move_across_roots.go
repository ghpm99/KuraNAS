package files

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"nas-go/api/internal/roots"
	"nas-go/api/pkg/utils"
)

const directoryManifestSize int64 = -1

func isMoveAcrossRoots(sourcePath string, destinationPath string) bool {
	sourceOwner, isSourceOwned := roots.OwnerOf(sourcePath)
	destinationOwner, isDestinationOwned := roots.OwnerOf(destinationPath)
	return isSourceOwned && isDestinationOwned && sourceOwner.Path != destinationOwner.Path
}

func isStorageRootPath(path string) bool {
	owner, isOwned := roots.OwnerOf(path)
	return isOwned && owner.Path == filepath.Clean(path)
}

func (s *Service) moveAcrossRoots(sourceFile FileDto, resolvedSourcePath string, resolvedDestinationPath string, destinationFolderID *int, destinationPath string) (string, error) {
	if isStorageRootPath(resolvedSourcePath) {
		return "", newFileOperationError(
			http.StatusBadRequest,
			"ERROR_MOVE_ROOT_FORBIDDEN",
			fmt.Errorf("cannot move a storage root"),
		)
	}

	coldDescendants, err := s.listColdDescendantsOfDirectory(sourceFile, resolvedSourcePath)
	if err != nil {
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_MOVE_FAILED", err)
	}

	copiedPath, err := s.CopyFile(sourceFile.ID, destinationFolderID, destinationPath, "")
	if err != nil {
		if isCopyFailureAfterWriting(err) {
			s.discardCopiedDestination(resolvedDestinationPath)
		}
		return "", err
	}

	if err := s.verifyCopiedContent(sourceFile, resolvedSourcePath, copiedPath, coldDescendants); err != nil {
		s.discardCopiedDestination(copiedPath)
		return "", newFileOperationError(http.StatusInternalServerError, "ERROR_MOVE_VERIFY_FAILED", err)
	}

	if err := s.DeleteFileFromDisk(sourceFile.ID, true); err != nil {
		s.discardCopiedDestinationWhenSourceIsIntact(sourceFile, resolvedSourcePath, copiedPath, coldDescendants)
		return "", err
	}

	s.removeColdDescendantBytes(coldDescendants)
	return copiedPath, nil
}

func isCopyFailureAfterWriting(err error) bool {
	var operationErr *FileOperationError
	if errors.As(err, &operationErr) {
		return operationErr.StatusCode >= http.StatusInternalServerError
	}
	return true
}

func (s *Service) listColdDescendantsOfDirectory(sourceFile FileDto, resolvedSourcePath string) ([]FileModel, error) {
	if sourceFile.Type != Directory {
		return nil, nil
	}
	return s.listActiveColdDescendants(resolvedSourcePath)
}

func (s *Service) discardCopiedDestination(destinationPath string) {
	if err := os.RemoveAll(destinationPath); err != nil {
		s.logSyncFailure("MoveAcrossRoots", destinationPath, err)
	}
	if err := s.syncDeletedRows(destinationPath); err != nil {
		s.logSyncFailure("MoveAcrossRoots", destinationPath, err)
	}
	s.ScanDirTask(filepath.Dir(destinationPath))
}

func (s *Service) discardCopiedDestinationWhenSourceIsIntact(sourceFile FileDto, resolvedSourcePath string, destinationPath string, coldDescendants []FileModel) {
	if err := s.verifyCopiedContent(sourceFile, resolvedSourcePath, destinationPath, coldDescendants); err != nil {
		s.logSyncFailure("MoveAcrossRoots", resolvedSourcePath, fmt.Errorf("source no longer matches destination, destination kept: %w", err))
		return
	}
	s.discardCopiedDestination(destinationPath)
}

func (s *Service) removeColdDescendantBytes(coldDescendants []FileModel) {
	for _, coldDescendant := range coldDescendants {
		if err := os.Remove(coldDescendant.PhysicalPath.String); err != nil && !errors.Is(err, fs.ErrNotExist) {
			s.logSyncFailure("MoveAcrossRoots", coldDescendant.PhysicalPath.String, err)
		}
	}
}

func (s *Service) verifyCopiedContent(sourceFile FileDto, resolvedSourcePath string, destinationPath string, coldDescendants []FileModel) error {
	sourceContentPath := resolvedSourcePath
	if sourceFile.PhysicalPath != "" {
		sourceContentPath = sourceFile.ResolveContentPath()
	}

	sourceManifest, err := buildContentManifest(sourceContentPath)
	if err != nil {
		return err
	}
	coldManifest, err := coldDescendantManifest(resolvedSourcePath, coldDescendants)
	if err != nil {
		return err
	}
	maps.Copy(sourceManifest, coldManifest)

	destinationManifest, err := buildContentManifest(destinationPath)
	if err != nil {
		return err
	}

	if !maps.Equal(sourceManifest, destinationManifest) {
		return fmt.Errorf("destination content does not match source")
	}

	if sourceFile.Type != File || sourceFile.CheckSum == "" {
		return nil
	}
	return verifyChecksumsMatch(sourceContentPath, destinationPath)
}

func coldDescendantManifest(sourceDirectoryPath string, coldDescendants []FileModel) (map[string]int64, error) {
	directoryPrefix := sourceDirectoryPath + string(filepath.Separator)
	manifest := make(map[string]int64, len(coldDescendants))
	for _, coldDescendant := range coldDescendants {
		coldInfo, err := os.Stat(coldDescendant.PhysicalPath.String)
		if err != nil {
			return nil, err
		}
		manifest[strings.TrimPrefix(coldDescendant.Path, directoryPrefix)] = coldInfo.Size()
	}
	return manifest, nil
}

func buildContentManifest(rootPath string) (map[string]int64, error) {
	manifest := make(map[string]int64)
	err := filepath.WalkDir(rootPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relativePath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			manifest[relativePath] = directoryManifestSize
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		manifest[relativePath] = info.Size()
		return nil
	})
	return manifest, err
}

func verifyChecksumsMatch(sourceContentPath string, destinationPath string) error {
	sourceChecksum, err := utils.GetFileChecksum(sourceContentPath)
	if err != nil {
		return err
	}
	destinationChecksum, err := utils.GetFileChecksum(destinationPath)
	if err != nil {
		return err
	}
	if sourceChecksum != destinationChecksum {
		return fmt.Errorf("checksum mismatch between source and destination")
	}
	return nil
}
