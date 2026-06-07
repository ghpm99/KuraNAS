package distribution

import (
	"nas-go/api/pkg/i18n"
)

// downloadURLPrefix is the route that serves the artifact bytes; the frontend
// builds its links from the DownloadURL field, so the prefix lives here next to
// the route definition in routes.go.
const downloadURLPrefix = "/api/v1/downloads/"

type Service struct {
	repository RepositoryInterface
}

func NewService(repository RepositoryInterface) *Service {
	return &Service{repository: repository}
}

func (s *Service) ListDownloads() ([]DownloadItemDto, error) {
	artifacts, err := s.repository.ListArtifacts()
	if err != nil {
		return nil, err
	}

	items := make([]DownloadItemDto, 0, len(artifacts))
	for _, artifact := range artifacts {
		items = append(items, toDto(artifact))
	}
	return items, nil
}

func (s *Service) ResolveDownload(id string) (string, string, error) {
	artifact, err := s.repository.GetArtifact(id)
	if err != nil {
		return "", "", err
	}
	return artifact.AbsPath, baseName(artifact.File), nil
}

// toDto resolves the i18n keys to the active locale (text is translated at the
// source) and assembles the public download URL.
func toDto(artifact Artifact) DownloadItemDto {
	return DownloadItemDto{
		ID:          artifact.ID,
		Name:        resolveKey(artifact.NameKey, artifact.ID),
		Description: resolveKey(artifact.DescriptionKey, ""),
		Platform:    artifact.Platform,
		Version:     artifact.Version,
		MinOS:       artifact.MinOS,
		SizeBytes:   artifact.SizeBytes,
		SHA256:      artifact.SHA256,
		DownloadURL: downloadURLPrefix + artifact.ID,
	}
}

// resolveKey translates a catalog key, falling back to a sensible default when
// the manifest omits the key (rather than surfacing a raw KEY string).
func resolveKey(key, fallback string) string {
	if key == "" {
		return fallback
	}
	if message := i18n.GetMessage(key); message != "" && message != key {
		return message
	}
	return fallback
}

func baseName(file string) string {
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' || file[i] == '\\' {
			return file[i+1:]
		}
	}
	return file
}
