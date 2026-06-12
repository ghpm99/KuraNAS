// Package dav embeds a WebDAV server so any standard client (Windows
// Explorer, davfs2/GNOME/KDE, Android file managers, Kodi/VLC) can mount the
// NAS as a network drive. Access control is the global IP whitelist — no
// credentials, per the project's no-auth decision; the handler must be
// registered behind that middleware and OUTSIDE the gzip group (compression
// corrupts PUT/PROPFIND bodies for native clients).
//
// The exposed tree starts at the storage roots: /dav/<label>/... maps into
// the root registered with that label. Internal directories (the per-root
// trash) are hidden and unreachable.
package dav

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"nas-go/api/internal/roots"

	"golang.org/x/net/webdav"
)

// trashDirName mirrors trash.DirName. Declared locally so dav does not import
// the trash domain for one constant.
const trashDirName = ".kuranas-trash"

// Prefix is the URL prefix the WebDAV tree is served under.
const Prefix = "/dav"

// NewHandler builds the WebDAV handler serving the enabled storage roots.
func NewHandler() http.Handler {
	return &webdav.Handler{
		Prefix:     Prefix,
		FileSystem: &rootsFS{},
		LockSystem: webdav.NewMemLS(),
	}
}

// rootsFS is a webdav.FileSystem whose level zero is the list of enabled
// storage roots; everything below dispatches into the owning root's disk
// directory. The registry is re-read per call, so root changes apply live.
type rootsFS struct{}

var errCrossRoot = fmt.Errorf("dav: rename across storage roots is not supported")

// splitPath normalizes a WebDAV path and returns its first segment (the root
// label) and the remainder inside that root ("" addresses the root itself).
// ok=false means the path is the virtual level zero.
func splitPath(name string) (label string, rest string, ok bool) {
	cleaned := path.Clean("/" + strings.ReplaceAll(name, "\\", "/"))
	if cleaned == "/" {
		return "", "", false
	}
	trimmed := strings.TrimPrefix(cleaned, "/")
	if index := strings.Index(trimmed, "/"); index >= 0 {
		return trimmed[:index], trimmed[index+1:], true
	}
	return trimmed, "", true
}

// hidesTrash reports whether any path segment is the internal trash dir.
func hidesTrash(rest string) bool {
	for _, segment := range strings.Split(rest, "/") {
		if segment == trashDirName {
			return true
		}
	}
	return false
}

// resolve maps a WebDAV path to the root-backed filesystem and the path
// inside it. A label that matches no enabled root, or a path touching the
// trash dir, resolves to not-exist.
func resolve(name string) (webdav.Dir, string, error) {
	label, rest, ok := splitPath(name)
	if !ok {
		return "", "", os.ErrInvalid
	}
	if rest != "" && hidesTrash(rest) {
		return "", "", os.ErrNotExist
	}
	for _, root := range roots.Enabled() {
		if root.Label == label {
			return webdav.Dir(root.Path), "/" + rest, nil
		}
	}
	return "", "", os.ErrNotExist
}

func (rootsFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	dir, rest, err := resolve(name)
	if err != nil {
		if err == os.ErrInvalid {
			return os.ErrPermission
		}
		return err
	}
	if rest == "/" {
		// The root labels themselves are managed in Settings, not via DAV.
		return os.ErrPermission
	}
	return dir.Mkdir(ctx, rest, perm)
}

func (rootsFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	dir, rest, err := resolve(name)
	if err != nil {
		if err == os.ErrInvalid {
			if flag&(os.O_WRONLY|os.O_RDWR|os.O_CREATE) != 0 {
				return nil, os.ErrPermission
			}
			return newVirtualRootDir(), nil
		}
		return nil, err
	}
	file, openErr := dir.OpenFile(ctx, rest, flag, perm)
	if openErr != nil {
		return nil, openErr
	}
	return &trashHidingFile{File: file}, nil
}

func (rootsFS) RemoveAll(ctx context.Context, name string) error {
	dir, rest, err := resolve(name)
	if err != nil {
		if err == os.ErrInvalid {
			return os.ErrPermission
		}
		return err
	}
	if rest == "/" {
		return os.ErrPermission
	}
	return dir.RemoveAll(ctx, rest)
}

func (rootsFS) Rename(ctx context.Context, oldName string, newName string) error {
	oldDir, oldRest, err := resolve(oldName)
	if err != nil {
		if err == os.ErrInvalid {
			return os.ErrPermission
		}
		return err
	}
	newDir, newRest, err := resolve(newName)
	if err != nil {
		if err == os.ErrInvalid {
			return os.ErrPermission
		}
		return err
	}
	if oldRest == "/" || newRest == "/" {
		return os.ErrPermission
	}
	if oldDir != newDir {
		// Different roots usually mean different volumes (EXDEV).
		return errCrossRoot
	}
	return oldDir.Rename(ctx, oldRest, newRest)
}

func (rootsFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	dir, rest, err := resolve(name)
	if err != nil {
		if err == os.ErrInvalid {
			return virtualDirInfo{name: "/"}, nil
		}
		return nil, err
	}
	info, statErr := dir.Stat(ctx, rest)
	if statErr != nil {
		return nil, statErr
	}
	if rest == "/" {
		// Surface the registered label, not the on-disk base name.
		label, _, _ := splitPath(name)
		return renamedDirInfo{FileInfo: info, label: label}, nil
	}
	return info, nil
}

// trashHidingFile filters the internal trash dir out of directory listings.
type trashHidingFile struct {
	webdav.File
}

func (f *trashHidingFile) Readdir(count int) ([]fs.FileInfo, error) {
	entries, err := f.File.Readdir(count)
	if err != nil {
		return entries, err
	}
	filtered := entries[:0]
	for _, entry := range entries {
		if entry.Name() == trashDirName {
			continue
		}
		filtered = append(filtered, entry)
	}
	return filtered, nil
}

// virtualDirInfo is the FileInfo of the synthetic level-zero directory and of
// each root entry inside it.
type virtualDirInfo struct {
	name    string
	modTime time.Time
}

func (i virtualDirInfo) Name() string       { return i.name }
func (i virtualDirInfo) Size() int64        { return 0 }
func (i virtualDirInfo) Mode() os.FileMode  { return os.ModeDir | 0755 }
func (i virtualDirInfo) ModTime() time.Time { return i.modTime }
func (i virtualDirInfo) IsDir() bool        { return true }
func (i virtualDirInfo) Sys() any           { return nil }

// renamedDirInfo decorates a real directory's FileInfo with the root label.
type renamedDirInfo struct {
	os.FileInfo
	label string
}

func (i renamedDirInfo) Name() string { return i.label }

// virtualRootDir is the read-only level-zero directory listing the enabled
// roots, one entry per registered label.
type virtualRootDir struct {
	entries []fs.FileInfo
	offset  int
}

func newVirtualRootDir() *virtualRootDir {
	enabled := roots.Enabled()
	entries := make([]fs.FileInfo, 0, len(enabled))
	for _, root := range enabled {
		modTime := time.Now()
		if info, err := os.Stat(root.Path); err == nil {
			modTime = info.ModTime()
		}
		entries = append(entries, virtualDirInfo{name: root.Label, modTime: modTime})
	}
	return &virtualRootDir{entries: entries}
}

func (d *virtualRootDir) Close() error              { return nil }
func (d *virtualRootDir) Read([]byte) (int, error)  { return 0, os.ErrInvalid }
func (d *virtualRootDir) Write([]byte) (int, error) { return 0, os.ErrPermission }
func (d *virtualRootDir) Seek(int64, int) (int64, error) {
	return 0, os.ErrInvalid
}

func (d *virtualRootDir) Stat() (fs.FileInfo, error) {
	return virtualDirInfo{name: "/"}, nil
}

func (d *virtualRootDir) Readdir(count int) ([]fs.FileInfo, error) {
	if count <= 0 {
		remaining := d.entries[d.offset:]
		d.offset = len(d.entries)
		return remaining, nil
	}
	if d.offset >= len(d.entries) {
		return nil, io.EOF
	}
	end := min(d.offset+count, len(d.entries))
	page := d.entries[d.offset:end]
	d.offset = end
	return page, nil
}
