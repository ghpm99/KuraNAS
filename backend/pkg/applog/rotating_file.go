package applog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// RotateConfig bounds the forensic log on disk so it never grows unbounded.
type RotateConfig struct {
	Dir        string // directory holding the log files
	Prefix     string // file name prefix, e.g. "kuranas-"
	MaxSizeMB  int    // rotate the active file once it exceeds this size
	MaxBackups int    // how many rotated files to keep (0 = keep all)
	MaxAgeDays int    // delete rotated files older than this (0 = no age limit)
}

// RotatingFile is an io.Writer that appends to a single active log file and
// rolls it over to a timestamped name once MaxSizeMB is exceeded, pruning old
// files by count and age. It is safe for concurrent use.
type RotatingFile struct {
	cfg RotateConfig

	mu      sync.Mutex
	file    *os.File
	size    int64
	maxSize int64
}

// NewRotatingFile opens (or creates) the active log file and returns a writer
// that rotates it. The active file is "<Prefix><timestamp>.log".
func NewRotatingFile(cfg RotateConfig) (*RotatingFile, error) {
	if cfg.Dir == "" {
		return nil, fmt.Errorf("applog: rotate dir is required")
	}
	if cfg.Prefix == "" {
		cfg.Prefix = "kuranas-"
	}
	if cfg.MaxSizeMB <= 0 {
		cfg.MaxSizeMB = 50
	}

	if err := os.MkdirAll(cfg.Dir, 0o755); err != nil {
		return nil, err
	}

	rf := &RotatingFile{
		cfg:     cfg,
		maxSize: int64(cfg.MaxSizeMB) * 1024 * 1024,
	}
	if err := rf.openNew(); err != nil {
		return nil, err
	}
	rf.prune()
	return rf, nil
}

// File returns the currently active *os.File, used to redirect the OS stderr
// handle at it. The pointer changes on rotation, but stderr only needs the
// initial one for crash capture.
func (rf *RotatingFile) File() *os.File {
	rf.mu.Lock()
	defer rf.mu.Unlock()
	return rf.file
}

func (rf *RotatingFile) Write(p []byte) (int, error) {
	rf.mu.Lock()
	defer rf.mu.Unlock()

	if rf.file != nil && rf.size+int64(len(p)) > rf.maxSize {
		if err := rf.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := rf.file.Write(p)
	rf.size += int64(n)
	return n, err
}

func (rf *RotatingFile) openNew() error {
	path := rf.nextPath()

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	info, statErr := file.Stat()
	rf.file = file
	if statErr == nil {
		rf.size = info.Size()
	} else {
		rf.size = 0
	}
	return nil
}

// nextPath returns a fresh, non-existing file path for the active log. The base
// name carries a second-resolution timestamp; a numeric suffix disambiguates
// rotations that happen within the same second.
func (rf *RotatingFile) nextPath() string {
	base := rf.cfg.Prefix + time.Now().UTC().Format("2006-01-02_15-04-05")
	path := filepath.Join(rf.cfg.Dir, base+".log")
	for i := 1; ; i++ {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return path
		}
		path = filepath.Join(rf.cfg.Dir, fmt.Sprintf("%s.%d.log", base, i))
	}
}

func (rf *RotatingFile) rotate() error {
	if rf.file != nil {
		_ = rf.file.Close()
	}
	if err := rf.openNew(); err != nil {
		return err
	}
	rf.prune()
	return nil
}

// prune removes rotated files beyond MaxBackups (by count, oldest first) and
// older than MaxAgeDays. Best-effort: removal errors are ignored.
func (rf *RotatingFile) prune() {
	if rf.cfg.MaxBackups <= 0 && rf.cfg.MaxAgeDays <= 0 {
		return
	}

	entries, err := os.ReadDir(rf.cfg.Dir)
	if err != nil {
		return
	}

	type logFile struct {
		path    string
		modTime time.Time
	}
	var files []logFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), rf.cfg.Prefix) || !strings.HasSuffix(entry.Name(), ".log") {
			continue
		}
		info, infoErr := entry.Info()
		if infoErr != nil {
			continue
		}
		files = append(files, logFile{path: filepath.Join(rf.cfg.Dir, entry.Name()), modTime: info.ModTime()})
	}

	// Newest first so the active file and the most recent backups are kept.
	sort.Slice(files, func(i, j int) bool { return files[i].modTime.After(files[j].modTime) })

	activePath := ""
	if rf.file != nil {
		activePath = rf.file.Name()
	}

	cutoff := time.Time{}
	if rf.cfg.MaxAgeDays > 0 {
		cutoff = time.Now().Add(-time.Duration(rf.cfg.MaxAgeDays) * 24 * time.Hour)
	}

	kept := 0
	for _, f := range files {
		if f.path == activePath {
			kept++
			continue
		}
		tooMany := rf.cfg.MaxBackups > 0 && kept >= rf.cfg.MaxBackups
		tooOld := rf.cfg.MaxAgeDays > 0 && f.modTime.Before(cutoff)
		if tooMany || tooOld {
			_ = os.Remove(f.path)
			continue
		}
		kept++
	}
}
