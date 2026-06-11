package scan

import (
	"fmt"
	"nas-go/api/internal/api/v1/files"
)

func GetCheckSum(fileDto files.FileDto,
	getFileChecksum func(path string) (string, error),
	getDirectoryChecksum func(dirPath string) (string, error),
) (string, error) {

	switch fileDto.Type {
	case files.File:
		return getFileChecksum(fileDto.Path)
	case files.Directory:
		return getDirectoryChecksum(fileDto.Path)
	default:
		return "", fmt.Errorf("file type not found")
	}

}
