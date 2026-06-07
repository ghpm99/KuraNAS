package distribution

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// ErrArtifactNotFound is returned when an id has no present artifact.
var ErrArtifactNotFound = errors.New("artifact not found")

const manifestFileName = "manifest.json"

// Repository reads the distributable client apps from a downloads directory.
// There is no database: artifacts are pre-built files on disk plus a manifest
// describing them. A missing or unreadable manifest yields an empty catalog
// (e.g. the dev server, which ships no downloads), never an error.
type Repository struct {
	baseDir string
}

func NewRepository(baseDir string) *Repository {
	return &Repository{baseDir: baseDir}
}

func (r *Repository) ListArtifacts() ([]Artifact, error) {
	parsed, err := r.readManifest()
	if err != nil {
		return nil, err
	}

	present := make([]Artifact, 0, len(parsed.Artifacts))
	for _, artifact := range parsed.Artifacts {
		resolved, ok := r.resolve(artifact)
		if !ok {
			continue
		}
		present = append(present, resolved)
	}
	return present, nil
}

func (r *Repository) GetArtifact(id string) (Artifact, error) {
	parsed, err := r.readManifest()
	if err != nil {
		return Artifact{}, err
	}

	for _, artifact := range parsed.Artifacts {
		if artifact.ID != id {
			continue
		}
		if resolved, ok := r.resolve(artifact); ok {
			return resolved, nil
		}
		break
	}
	return Artifact{}, ErrArtifactNotFound
}

// readManifest loads downloads/manifest.json. A missing manifest is treated as
// an empty catalog so the endpoint degrades gracefully when no apps are shipped.
func (r *Repository) readManifest() (manifest, error) {
	path := filepath.Join(r.baseDir, manifestFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return manifest{}, nil
		}
		return manifest{}, err
	}

	var parsed manifest
	if err := json.Unmarshal(data, &parsed); err != nil {
		return manifest{}, err
	}
	return parsed, nil
}

// resolve validates an artifact's file, keeps it inside baseDir (no path
// traversal), confirms it exists, and fills SizeBytes and AbsPath. It returns
// false when the artifact has no usable file on disk.
func (r *Repository) resolve(artifact Artifact) (Artifact, bool) {
	if strings.TrimSpace(artifact.ID) == "" || strings.TrimSpace(artifact.File) == "" {
		return Artifact{}, false
	}

	base, err := filepath.Abs(r.baseDir)
	if err != nil {
		return Artifact{}, false
	}

	full := filepath.Join(base, filepath.FromSlash(artifact.File))
	rel, err := filepath.Rel(base, full)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return Artifact{}, false
	}

	info, err := os.Stat(full)
	if err != nil || info.IsDir() {
		return Artifact{}, false
	}

	artifact.AbsPath = full
	artifact.SizeBytes = info.Size()
	return artifact, true
}
