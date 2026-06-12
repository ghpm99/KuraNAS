// Package backup implements the incremental copy engine behind the backup_run
// job (task 12). Backup is not a mirror: every file the source replaced or
// deleted is moved into a timestamped _versions/ area and only expires after
// the configured retention, so ransomware or an accidental delete on the
// source never silently destroys the only other copy.
package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Layout inside the destination directory.
const (
	CurrentDirName  = "current"
	VersionsDirName = "_versions"
	tmpPrefix       = ".kuranas-bak-tmp-"
	versionStampFmt = "20060102-150405"
)

// Root is one source tree to back up; Label keys its subtree under current/.
type Root struct {
	Label string
	Path  string
}

// Stamper marks a source file as backed up (home_file.last_backup). It is only
// called after the copy has been checksum-verified.
type Stamper func(sourcePath string, at time.Time) error

type Options struct {
	Roots         []Root
	Destination   string
	RetentionDays int
	// SkipDirNames are directory basenames never backed up (the trash dir).
	SkipDirNames []string
	Stamp        Stamper
	Now          func() time.Time
}

type Stats struct {
	Scanned   int `json:"scanned"`
	Copied    int `json:"copied"`
	Versioned int `json:"versioned"`
	Purged    int `json:"purged"`
	Failures  int `json:"failures"`
}

// Run executes one incremental backup pass: copy new/changed files (verified
// by checksum before stamping), move replaced/deleted copies into _versions/,
// and purge versions older than the retention. Per-file errors are counted and
// skipped — one bad file never aborts the run.
func Run(opts Options) (Stats, error) {
	stats := Stats{}

	if opts.Now == nil {
		opts.Now = time.Now
	}
	destination := filepath.Clean(strings.TrimSpace(opts.Destination))
	if destination == "" || destination == "." {
		return stats, errors.New("backup: destination not configured")
	}
	if err := ValidateDestination(destination, opts.Roots); err != nil {
		return stats, err
	}

	currentDir := filepath.Join(destination, CurrentDirName)
	versionsDir := filepath.Join(destination, VersionsDirName)
	if err := os.MkdirAll(currentDir, 0o755); err != nil {
		return stats, fmt.Errorf("backup: cannot create destination: %w", err)
	}

	removeLeftoverTempFiles(currentDir)

	runTime := opts.Now()
	versionRunDir := filepath.Join(versionsDir, runTime.UTC().Format(versionStampFmt))

	for _, root := range opts.Roots {
		copyRoot(root, currentDir, versionRunDir, runTime, opts, &stats)
	}

	versionDeletedFiles(opts.Roots, currentDir, versionRunDir, &stats)
	purgeExpiredVersions(versionsDir, opts.RetentionDays, runTime, &stats)

	return stats, nil
}

// ValidateDestination rejects an empty destination and any destination inside
// (or equal to) an indexed root — that is what keeps the backup area invisible
// to the scanner, the watcher, the tree and analytics.
func ValidateDestination(destination string, roots []Root) error {
	cleaned := filepath.Clean(strings.TrimSpace(destination))
	if cleaned == "" || cleaned == "." {
		return errors.New("backup: destination not configured")
	}
	if !filepath.IsAbs(cleaned) {
		return errors.New("backup: destination must be an absolute path")
	}
	for _, root := range roots {
		rootPath := filepath.Clean(root.Path)
		if cleaned == rootPath || strings.HasPrefix(cleaned+string(filepath.Separator), rootPath+string(filepath.Separator)) {
			return fmt.Errorf("backup: destination %q is inside the indexed root %q", cleaned, rootPath)
		}
	}
	return nil
}

func copyRoot(root Root, currentDir string, versionRunDir string, runTime time.Time, opts Options, stats *Stats) {
	rootTarget := filepath.Join(currentDir, root.Label)

	walkErr := filepath.WalkDir(root.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			log.Printf("[backup] skipping inaccessible path %q: %v\n", path, err)
			return nil
		}
		if d.IsDir() {
			for _, skip := range opts.SkipDirNames {
				if d.Name() == skip {
					return filepath.SkipDir
				}
			}
			return nil
		}

		relPath, relErr := filepath.Rel(root.Path, path)
		if relErr != nil {
			stats.Failures++
			return nil
		}

		stats.Scanned++
		target := filepath.Join(rootTarget, relPath)

		sourceInfo, infoErr := d.Info()
		if infoErr != nil {
			stats.Failures++
			return nil
		}

		if upToDate(target, sourceInfo) {
			return nil
		}

		if copyErr := backupOneFile(path, target, sourceInfo, versionRunDir, currentDir, stats); copyErr != nil {
			log.Printf("[backup] failed to back up %q: %v\n", path, copyErr)
			stats.Failures++
			return nil
		}
		stats.Copied++

		if opts.Stamp != nil {
			if stampErr := opts.Stamp(path, runTime); stampErr != nil {
				log.Printf("[backup] failed to stamp last_backup for %q: %v\n", path, stampErr)
			}
		}
		return nil
	})
	if walkErr != nil {
		log.Printf("[backup] walk of root %q aborted: %v\n", root.Path, walkErr)
		stats.Failures++
	}
}

// upToDate reports whether the existing copy already matches the source by
// size + mtime (both truncated to the second, mirroring the scanner's diff).
func upToDate(target string, sourceInfo os.FileInfo) bool {
	targetInfo, err := os.Stat(target)
	if err != nil {
		return false
	}
	return targetInfo.Size() == sourceInfo.Size() &&
		targetInfo.ModTime().Truncate(time.Second).Equal(sourceInfo.ModTime().Truncate(time.Second))
}

// backupOneFile copies source → temp file in the target's directory, verifies
// the written bytes by checksum, moves any previous copy into the versions
// area, and atomically renames the temp file into place. A crash mid-copy
// leaves only a temp file behind, never a corrupt current/ entry.
func backupOneFile(source string, target string, sourceInfo os.FileInfo, versionRunDir string, currentDir string, stats *Stats) error {
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	sourceSum, tmpPath, err := copyToTemp(source, filepath.Dir(target))
	if err != nil {
		return err
	}
	defer os.Remove(tmpPath)

	writtenSum, err := checksumFile(tmpPath)
	if err != nil {
		return err
	}
	if writtenSum != sourceSum {
		return fmt.Errorf("checksum mismatch after copy (source %s, written %s)", sourceSum, writtenSum)
	}

	if _, statErr := os.Stat(target); statErr == nil {
		if versionErr := moveToVersions(target, currentDir, versionRunDir); versionErr != nil {
			return fmt.Errorf("could not preserve previous version: %w", versionErr)
		}
		stats.Versioned++
	}

	if err := os.Chtimes(tmpPath, sourceInfo.ModTime(), sourceInfo.ModTime()); err != nil {
		return err
	}
	return os.Rename(tmpPath, target)
}

// copyToTemp streams source into a fresh temp file beside the target, hashing
// the source bytes on the way through.
func copyToTemp(source string, targetDir string) (sum string, tmpPath string, err error) {
	in, err := os.Open(source)
	if err != nil {
		return "", "", err
	}
	defer in.Close()

	tmp, err := os.CreateTemp(targetDir, tmpPrefix+"*")
	if err != nil {
		return "", "", err
	}
	tmpPath = tmp.Name()

	hasher := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(tmp, hasher), in)
	closeErr := tmp.Close()
	if copyErr != nil {
		os.Remove(tmpPath)
		return "", "", copyErr
	}
	if closeErr != nil {
		os.Remove(tmpPath)
		return "", "", closeErr
	}

	return hex.EncodeToString(hasher.Sum(nil)), tmpPath, nil
}

func checksumFile(path string) (string, error) {
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

// moveToVersions relocates an existing copy under the run's _versions/<stamp>/
// directory, preserving its current/-relative path.
func moveToVersions(target string, currentDir string, versionRunDir string) error {
	relPath, err := filepath.Rel(currentDir, target)
	if err != nil {
		return err
	}
	versionPath := filepath.Join(versionRunDir, relPath)
	if err := os.MkdirAll(filepath.Dir(versionPath), 0o755); err != nil {
		return err
	}
	return os.Rename(target, versionPath)
}

// versionDeletedFiles sweeps current/ for copies whose source file no longer
// exists and moves them into the versions area — deletion on the source must
// stay recoverable for the whole retention window.
func versionDeletedFiles(rootList []Root, currentDir string, versionRunDir string, stats *Stats) {
	rootByLabel := make(map[string]Root, len(rootList))
	for _, root := range rootList {
		rootByLabel[root.Label] = root
	}

	walkErr := filepath.WalkDir(currentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}

		relPath, relErr := filepath.Rel(currentDir, path)
		if relErr != nil {
			return nil
		}
		parts := strings.SplitN(relPath, string(filepath.Separator), 2)
		if len(parts) < 2 {
			return nil
		}
		root, known := rootByLabel[parts[0]]
		if !known {
			// A subtree of a root that is no longer configured: leave it alone;
			// removing roots is an operator decision, not the job's.
			return nil
		}

		sourcePath := filepath.Join(root.Path, parts[1])
		if _, statErr := os.Stat(sourcePath); statErr == nil {
			return nil
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return nil
		}

		if moveErr := moveToVersions(path, currentDir, versionRunDir); moveErr != nil {
			log.Printf("[backup] failed to version deleted file %q: %v\n", path, moveErr)
			stats.Failures++
			return nil
		}
		stats.Versioned++
		return nil
	})
	if walkErr != nil {
		log.Printf("[backup] deletion sweep aborted: %v\n", walkErr)
		stats.Failures++
	}

	removeEmptyDirs(currentDir)
}

// purgeExpiredVersions deletes _versions/<stamp> directories older than the
// retention window. RetentionDays <= 0 falls back to 30 — purge must never
// mean "purge everything" by accident.
func purgeExpiredVersions(versionsDir string, retentionDays int, now time.Time, stats *Stats) {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	cutoff := now.UTC().Add(-time.Duration(retentionDays) * 24 * time.Hour)

	entries, err := os.ReadDir(versionsDir)
	if err != nil {
		return
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		stamp, parseErr := time.Parse(versionStampFmt, entry.Name())
		if parseErr != nil {
			continue
		}
		if stamp.Before(cutoff) {
			if removeErr := os.RemoveAll(filepath.Join(versionsDir, entry.Name())); removeErr != nil {
				log.Printf("[backup] failed to purge versions %q: %v\n", entry.Name(), removeErr)
				stats.Failures++
				continue
			}
			stats.Purged++
		}
	}
}

// removeLeftoverTempFiles clears temp files orphaned by an interrupted run.
func removeLeftoverTempFiles(currentDir string) {
	_ = filepath.WalkDir(currentDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), tmpPrefix) {
			_ = os.Remove(path)
		}
		return nil
	})
}

// removeEmptyDirs prunes directories emptied by the deletion sweep, walking
// bottom-up so nested empty chains collapse in one pass.
func removeEmptyDirs(rootDir string) {
	var dirs []string
	_ = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && d.IsDir() && path != rootDir {
			dirs = append(dirs, path)
		}
		return nil
	})
	for i := len(dirs) - 1; i >= 0; i-- {
		entries, err := os.ReadDir(dirs[i])
		if err == nil && len(entries) == 0 {
			_ = os.Remove(dirs[i])
		}
	}
}
