//go:build !windows

package files

import "syscall"

func availableBytesAt(path string) (int64, error) {
	var filesystemStat syscall.Statfs_t
	if err := syscall.Statfs(path, &filesystemStat); err != nil {
		return 0, err
	}
	return int64(filesystemStat.Bavail) * int64(filesystemStat.Bsize), nil
}
