// Package roots is the in-memory registry of storage roots — the N directories
// KuraNAS indexes, watches and serves. It is the neutral, dependency-free
// package every layer may import (config-style global): the storageroots
// domain pushes the registered roots in at boot and on every CRUD change,
// and hot paths (DTO conversion, request path resolution, workers) read it
// without touching the database.
//
// Compatibility contract: while nothing was pushed (tests, boot before the DB,
// installs without the storage_root table) the registry behaves exactly like
// the single ENTRY_POINT world — one synthetic primary root.
package roots

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"nas-go/api/internal/config"
)

type Root struct {
	ID      int
	Path    string // absolute, cleaned
	Label   string
	Enabled bool
}

var (
	mu         sync.RWMutex
	registered []Root
)

// Set replaces the registry content. Paths are cleaned; order is preserved
// (the first enabled root is the primary root — the seeded ENTRY_POINT).
func Set(list []Root) {
	cleaned := make([]Root, 0, len(list))
	for _, root := range list {
		trimmedPath := strings.TrimSpace(root.Path)
		if trimmedPath == "" {
			continue
		}
		root.Path = filepath.Clean(trimmedPath)
		root.Label = strings.TrimSpace(root.Label)
		cleaned = append(cleaned, root)
	}

	mu.Lock()
	registered = cleaned
	mu.Unlock()
}

// Reset clears the registry (tests).
func Reset() {
	mu.Lock()
	registered = nil
	mu.Unlock()
}

func fallbackRoot() (Root, bool) {
	entryPoint := strings.TrimSpace(config.AppConfig.EntryPoint)
	if entryPoint == "" {
		return Root{}, false
	}
	cleanPath := filepath.Clean(entryPoint)
	return Root{Path: cleanPath, Label: filepath.Base(cleanPath), Enabled: true}, true
}

// Enabled returns the enabled roots in registration order. With an empty
// registry it falls back to the single ENTRY_POINT root.
func Enabled() []Root {
	mu.RLock()
	defer mu.RUnlock()

	enabled := make([]Root, 0, len(registered))
	for _, root := range registered {
		if root.Enabled {
			enabled = append(enabled, root)
		}
	}
	if len(enabled) > 0 {
		return enabled
	}
	if fallback, ok := fallbackRoot(); ok {
		return []Root{fallback}
	}
	return nil
}

// Primary returns the first enabled root — the root that keeps the legacy
// single-ENTRY_POINT semantics (bare relative paths, level-zero listing).
func Primary() (Root, bool) {
	enabled := Enabled()
	if len(enabled) == 0 {
		return Root{}, false
	}
	return enabled[0], true
}

func isUnder(root string, path string) bool {
	return path == root || strings.HasPrefix(path, root+string(filepath.Separator))
}

// OwnerOf returns the enabled root that contains path (longest-prefix match).
func OwnerOf(path string) (Root, bool) {
	cleanPath := filepath.Clean(strings.TrimSpace(path))
	var owner Root
	found := false
	for _, root := range Enabled() {
		if !isUnder(root.Path, cleanPath) {
			continue
		}
		if !found || len(root.Path) > len(owner.Path) {
			owner = root
			found = true
		}
	}
	return owner, found
}

// ToRelativePath converts an absolute path to its client-visible form: paths
// under the primary root keep the legacy bare shape ("/fotos"); paths under an
// additional root are prefixed with the root label ("/Midia/fotos"). Paths
// outside every root are returned via the legacy primary-prefix strip, so the
// behavior with a single root is byte-for-byte the old one.
func ToRelativePath(absolutePath string) string {
	cleanPath := filepath.Clean(absolutePath)

	primary, hasPrimary := Primary()
	owner, owned := OwnerOf(cleanPath)
	if owned && (!hasPrimary || owner.Path != primary.Path) {
		rel := strings.TrimPrefix(cleanPath, owner.Path)
		rel = strings.TrimPrefix(rel, string(filepath.Separator))
		if rel == "" {
			return "/" + owner.Label
		}
		return "/" + owner.Label + "/" + filepath.ToSlash(rel)
	}

	if !hasPrimary {
		return cleanPath
	}
	rel := strings.TrimPrefix(cleanPath, primary.Path)
	if rel == "" || rel == "." {
		return "/"
	}
	if !strings.HasPrefix(rel, "/") {
		rel = "/" + rel
	}
	return rel
}

// ToAbsolutePath resolves a client-visible relative path back to disk. A first
// segment matching an additional root's label addresses that root (labels win
// over same-named folders of the primary root — validated at registration);
// anything else resolves under the primary root, exactly like the old
// ENTRY_POINT join.
func ToAbsolutePath(relativePath string) string {
	primary, hasPrimary := Primary()
	if !hasPrimary {
		return filepath.Clean(relativePath)
	}

	trimmed := strings.TrimPrefix(strings.TrimSpace(relativePath), "/")
	if trimmed == "" {
		return primary.Path
	}

	firstSegment := trimmed
	rest := ""
	if index := strings.IndexAny(trimmed, "/\\"); index >= 0 {
		firstSegment = trimmed[:index]
		rest = trimmed[index+1:]
	}

	for _, root := range Enabled() {
		if root.Path == primary.Path || root.Label == "" {
			continue
		}
		if root.Label == firstSegment {
			if rest == "" {
				return root.Path
			}
			return filepath.Join(root.Path, rest)
		}
	}

	return filepath.Join(primary.Path, trimmed)
}

// ResolveAbsolute validates that an input path (absolute, or relative to a
// root) lives under some enabled root and returns its cleaned absolute form.
// It is the multi-root successor of the entry-point containment check.
func ResolveAbsolute(inputPath string) (string, error) {
	enabled := Enabled()
	if len(enabled) == 0 {
		return "", fmt.Errorf("no storage root configured")
	}

	candidate := strings.TrimSpace(inputPath)
	if candidate == "" {
		primary, _ := Primary()
		return primary.Path, nil
	}

	if filepath.IsAbs(candidate) {
		cleanPath := filepath.Clean(candidate)
		if _, ok := OwnerOf(cleanPath); ok {
			return cleanPath, nil
		}
		return "", fmt.Errorf("path %q is outside every storage root", inputPath)
	}

	cleanPath := filepath.Clean(ToAbsolutePath(candidate))
	if _, ok := OwnerOf(cleanPath); ok {
		return cleanPath, nil
	}
	return "", fmt.Errorf("path %q is outside every storage root", inputPath)
}
