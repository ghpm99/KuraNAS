package captures

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// captureMetadata is the subset of the plugin's standardized metadata payload
// the promotion consumes. The plugin owns the full schema (see
// plugin/content/metadata-detector.js); every field is optional, so a missing
// one simply leaves the corresponding capture column empty. The raw payload is
// preserved verbatim in captures.raw_metadata, so new fields never need a
// backend change to be retained.
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

// parseCaptureMetadata decodes the raw payload, tolerating an absent or invalid
// blob by returning a zero value — promotion always proceeds (decided: sempre
// promove; metadado ruim é problema do usuário).
func parseCaptureMetadata(raw json.RawMessage) captureMetadata {
	var meta captureMetadata
	if len(raw) == 0 {
		return meta
	}
	_ = json.Unmarshal(raw, &meta)
	return meta
}

// resolvedTitle is the title used both for the library path and the captures
// row, falling back to the recording name when the metadata carries no title.
func resolvedTitle(meta captureMetadata, capture CaptureModel) string {
	title := strings.TrimSpace(meta.Title)
	if title == "" {
		title = capture.Name
	}
	return title
}

// buildCaptureRelPath computes the destination path (relative to the videos
// library root) from the data, not the content_type — the real discriminator is
// whether an episode number is present:
//   - with episode -> <title>/Temporada <season>/E<episode> - <episode_title>.ext
//     (no season -> no Temporada folder);
//   - otherwise    -> <title> (<release_year>).ext (no year -> <title>.ext).
//
// The extension is preserved from the uploaded recording (the plugin delivers a
// finished MP4; no transcode/remux happens here).
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

// collisionAvoidantPath returns absPath when free, otherwise the first
// "<base> (n)<ext>" (n>=2) variant that does not yet exist on disk — each
// capture is its own file (decided: sufixar (n)).
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

// applyMetadataToCapture copies the parsed semantic fields onto the capture row
// so captures becomes the source of truth for the rich metadata.
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
