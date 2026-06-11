package engine

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"time"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

// persistDirectoryRow ensures the directory at path has an active row in
// home_file. Directories carry no metadata/checksum/thumbnail work, so they are
// upserted directly through the files service instead of going through the job
// pipeline. mtime/size are not reliable change signals for directories, so an
// existing active row is left untouched.
func persistDirectoryRow(service files.ServiceInterface, path string, info os.FileInfo) error {
	dirDto := files.FileDto{
		Path:       path,
		ParentPath: filepath.Dir(path),
	}
	if err := dirDto.ParseFileInfoToFileDto(info); err != nil {
		return err
	}

	existing, err := service.GetFileByNameAndPath(dirDto.Name, dirDto.Path)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, createErr := service.CreateFile(dirDto)
			return createErr
		}
		return err
	}

	if !existing.DeletedAt.HasValue {
		return nil
	}

	// The directory is back on disk but its row is soft-deleted (e.g. the
	// folder was moved away and back) — revive it instead of duplicating.
	existing.DeletedAt = utils.Optional[time.Time]{}
	existing.UpdatedAt = dirDto.UpdatedAt
	_, updateErr := service.UpdateFile(existing)
	return updateErr
}
