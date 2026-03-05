//go:build windows

package analytics

import "golang.org/x/sys/windows"

func getFileSystemStats(path string) (int64, int64, error) {
	pathPtr, err := windows.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, err
	}

	var freeBytesAvailable uint64
	var totalBytes uint64
	var totalFreeBytes uint64
	if err := windows.GetDiskFreeSpaceEx(pathPtr, &freeBytesAvailable, &totalBytes, &totalFreeBytes); err != nil {
		return 0, 0, err
	}

	return int64(totalBytes), int64(freeBytesAvailable), nil
}
