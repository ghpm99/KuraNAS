// Package tiering implements the copy/verify/swap engine behind the
// tier_migration job (task 13). It moves a file's bytes between the hot tier
// (the SSD, where the logical path lives) and the cold tier (a directory on the
// cold volume) without ever changing the file's logical path — navigation,
// search and the media tabs keep showing it in the same place. The only thing
// that moves is home_file.physical_path.
//
// The whole point is crash safety: every operation copies-and-verifies the new
// copy first, records the new location in the DB second, and only deletes the
// old copy last. A process kill at any point leaves the file with *two* copies
// and a DB row that still points at a real one — never zero copies.
package tiering

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const tmpPrefix = ".kuranas-tier-tmp-"

// SetPhysicalPath records a file's new physical location. An empty path means
// "the bytes are back at the logical path" (promotion clears physical_path).
// The implementation persists the change in a single transaction.
type SetPhysicalPath func(fileID int, physicalPath string) error

// Demotion describes one hot→cold move: the bytes currently at HotPath are
// copied to ColdPath (precomputed by the caller, mirroring the root-relative
// structure under the cold directory).
type Demotion struct {
	FileID   int
	HotPath  string
	ColdPath string
}

// Promotion describes one cold→hot move: the bytes at ColdPath are copied back
// to HotPath (the logical path) and the cold copy is removed.
type Promotion struct {
	FileID   int
	HotPath  string
	ColdPath string
}

type Stats struct {
	Demoted  int `json:"demoted"`
	Promoted int `json:"promoted"`
	Failures int `json:"failures"`
}

// Run executes one migration pass: promote files that came back into use, then
// demote idle ones. Promotions run first so a freshly-used file never gets
// demoted again in the very same pass. Per-file errors are counted and skipped
// — one unreadable file never aborts the whole job.
func Run(coldDir string, promotions []Promotion, demotions []Demotion, setPhysical SetPhysicalPath) Stats {
	stats := Stats{}

	if coldDir != "" {
		removeLeftoverTempFiles(coldDir)
	}

	for _, item := range promotions {
		if err := promoteOne(item, setPhysical); err != nil {
			log.Printf("[tiering] failed to promote %q: %v\n", item.HotPath, err)
			stats.Failures++
			continue
		}
		stats.Promoted++
	}

	for _, item := range demotions {
		if err := demoteOne(item, setPhysical); err != nil {
			log.Printf("[tiering] failed to demote %q: %v\n", item.HotPath, err)
			stats.Failures++
			continue
		}
		stats.Demoted++
	}

	return stats
}

// demoteOne moves a hot file to the cold tier in the crash-safe order:
// copy+verify to cold, record physical_path, then delete the hot copy. A crash
// before the DB update leaves an extra cold copy that the next pass overwrites;
// a crash after it but before the delete leaves a stray hot copy that the diff
// blindagem ignores and the next pass cleans (physical_path is already set).
func demoteOne(item Demotion, setPhysical SetPhysicalPath) error {
	if err := copyVerified(item.HotPath, item.ColdPath); err != nil {
		return err
	}
	if err := setPhysical(item.FileID, item.ColdPath); err != nil {
		return fmt.Errorf("could not record physical_path: %w", err)
	}
	if err := os.Remove(item.HotPath); err != nil && !os.IsNotExist(err) {
		// The DB already points at the cold copy, so this is not data loss —
		// only a stray hot copy. Log and move on; the next pass removes it.
		log.Printf("[tiering] demoted %q but could not remove the hot copy: %v\n", item.HotPath, err)
	}
	return nil
}

// promoteOne is the inverse: copy+verify back to the hot path, clear
// physical_path, then delete the cold copy. Same crash guarantee — the bytes
// always survive on at least one tier.
func promoteOne(item Promotion, setPhysical SetPhysicalPath) error {
	if err := copyVerified(item.ColdPath, item.HotPath); err != nil {
		return err
	}
	if err := setPhysical(item.FileID, ""); err != nil {
		return fmt.Errorf("could not clear physical_path: %w", err)
	}
	if err := os.Remove(item.ColdPath); err != nil && !os.IsNotExist(err) {
		log.Printf("[tiering] promoted %q but could not remove the cold copy: %v\n", item.HotPath, err)
	}
	return nil
}

// copyVerified streams source → a temp file beside the destination, checks the
// written bytes against the source checksum, preserves the modification time
// (so the scanner's size+mtime diff still sees the copy as up to date), then
// atomically renames it into place. A crash mid-copy leaves only a temp file.
func copyVerified(source string, destination string) error {
	if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
		return err
	}

	sourceInfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	sourceSum, tmpPath, err := copyToTemp(source, filepath.Dir(destination))
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

	if err := os.Chtimes(tmpPath, sourceInfo.ModTime(), sourceInfo.ModTime()); err != nil {
		return err
	}
	return os.Rename(tmpPath, destination)
}

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

// removeLeftoverTempFiles clears temp files orphaned by an interrupted pass.
// The cold directory is the only place demotion writes temps; the hot side
// writes them beside each logical file, swept by the scan as needed.
func removeLeftoverTempFiles(coldDir string) {
	_ = filepath.WalkDir(coldDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), tmpPrefix) {
			_ = os.Remove(path)
		}
		return nil
	})
}
