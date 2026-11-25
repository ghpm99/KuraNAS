package captures

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type captureMetadata struct {
	Platform      string   `json:"platform"`
	SourceURL     string   `json:"source_url"`
	ContentType   string   `json:"content_type"`
	Title         string   `json:"title"`
	EpisodeTitle  string   `json:"episode_title"`
	Season        *int     `json:"season"`
	Episode       *int     `json:"episode"`
	Description   string   `json:"description"`
	ReleaseYear   *int     `json:"release_year"`
	Genres        []string `json:"genres"`
	Cast          []string `json:"cast"`
	Directors     []string `json:"directors"`
	Studio        string   `json:"studio"`
	ContentRating string   `json:"content_rating"`
	ThumbnailURL  string   `json:"thumbnail_url"`
	PosterURL     string   `json:"poster_url"`
}

func parseCaptureMetadata(raw json.RawMessage) captureMetadata {
	var meta captureMetadata
	if len(raw) == 0 {
		return meta
	}
	_ = json.Unmarshal(raw, &meta)
	return meta
}

func resolvedTitle(meta captureMetadata, capture CaptureModel) string {
	title := strings.TrimSpace(meta.Title)
	if title == "" {
		title = capture.Name
	}
	return title
}

func buildCaptureRelPath(meta captureMetadata, capture CaptureModel) string {
	ext := filepath.Ext(capture.FileName)
	if ext == "" {
		ext = ".mp4"
	}

	title := sanitizeFileName(resolvedTitle(meta, capture))

	if meta.Episode != nil {
		episodeTitle := strings.TrimSpace(meta.EpisodeTitle)
		var fileBase string
		if episodeTitle != "" {
			fileBase = fmt.Sprintf("E%d - %s%s", *meta.Episode, sanitizeFileName(episodeTitle), ext)
		} else {
			fileBase = fmt.Sprintf("E%d%s", *meta.Episode, ext)
		}
		if meta.Season != nil {
			return filepath.Join(title, fmt.Sprintf("Temporada %d", *meta.Season), fileBase)
		}
		return filepath.Join(title, fileBase)
	}

	if meta.ReleaseYear != nil {
		return fmt.Sprintf("%s (%d)%s", title, *meta.ReleaseYear, ext)
	}
	return title + ext
}

func collisionAvoidantPath(absPath string) string {
	if !pathExistsOnDisk(absPath) {
		return absPath
	}
	ext := filepath.Ext(absPath)
	base := strings.TrimSuffix(absPath, ext)
	for n := 2; ; n++ {
		candidate := fmt.Sprintf("%s (%d)%s", base, n, ext)
		if !pathExistsOnDisk(candidate) {
			return candidate
		}
	}
}

func pathExistsOnDisk(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func applyMetadataToCapture(capture *CaptureModel, meta captureMetadata) {
	capture.Title = resolvedTitle(meta, *capture)
	capture.EpisodeTitle = strings.TrimSpace(meta.EpisodeTitle)
	capture.Season = meta.Season
	capture.Episode = meta.Episode
	capture.Description = strings.TrimSpace(meta.Description)
	capture.ReleaseYear = meta.ReleaseYear
	capture.Genres = meta.Genres
	capture.Cast = meta.Cast
	capture.Directors = meta.Directors
	capture.Studio = strings.TrimSpace(meta.Studio)
	capture.ContentRating = strings.TrimSpace(meta.ContentRating)
	capture.Platform = strings.TrimSpace(meta.Platform)
	capture.SourceURL = strings.TrimSpace(meta.SourceURL)
	capture.ThumbnailURL = strings.TrimSpace(meta.ThumbnailURL)
	capture.ContentType = strings.TrimSpace(meta.ContentType)
}
