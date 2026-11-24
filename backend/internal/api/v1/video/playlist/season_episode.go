package playlist

import (
	"path/filepath"
	"regexp"
	"strconv"
)

const (
	defaultSeasonNumber = 1

	trailingEpisodeBody = `\s-\s(\d{1,3})(?:\s*[\[(][^\])]*[\])])*`
)

var (
	seasonAndEpisodePatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)` + tokenStart + seasonEpisodeBody + tokenEnd),
		regexp.MustCompile(`(?i)` + tokenStart + longFormEpisodeBody + tokenEnd),
		regexp.MustCompile(`(?i)` + tokenStart + crossEpisodeBody + tokenEnd),
	}
	episodeOnlyPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)` + tokenStart + keywordEpisodeBody + tokenEnd),
		regexp.MustCompile(`(?i)` + trailingEpisodeBody + `$`),
	}
	videoExtensionPattern = regexp.MustCompile(`^\.[A-Za-z][A-Za-z0-9]{1,3}$`)
)

// ParseSeasonEpisode extracts the season and episode numbers from a video
// file name. Names that carry only an episode number are assigned season 1.
// It returns ok=false when no episode marker is found; resolution, codec and
// year tokens (1080p, x264, 2019) are never read as episode numbers.
func ParseSeasonEpisode(fileName string) (season int, episode int, ok bool) {
	stem := stripVideoExtension(fileName)

	for _, pattern := range seasonAndEpisodePatterns {
		match := pattern.FindStringSubmatch(stem)
		if match == nil {
			continue
		}
		parsedSeason, seasonErr := strconv.Atoi(match[1])
		parsedEpisode, episodeErr := strconv.Atoi(match[2])
		if seasonErr != nil || episodeErr != nil || parsedEpisode < 1 {
			continue
		}
		return parsedSeason, parsedEpisode, true
	}

	for _, pattern := range episodeOnlyPatterns {
		match := pattern.FindStringSubmatch(stem)
		if match == nil {
			continue
		}
		parsedEpisode, episodeErr := strconv.Atoi(match[1])
		if episodeErr != nil || parsedEpisode < 1 {
			continue
		}
		return defaultSeasonNumber, parsedEpisode, true
	}

	return 0, 0, false
}

func stripVideoExtension(fileName string) string {
	extension := filepath.Ext(fileName)
	if !videoExtensionPattern.MatchString(extension) {
		return fileName
	}
	return fileName[:len(fileName)-len(extension)]
}

// IsSeriesLikePlaylist reports whether an automatic playlist is ordered by
// season and episode: series playlists (including the ones built by the
// sequential and captured series strategies) and series/anime classifications.
func IsSeriesLikePlaylist(playlistType string, classification string) bool {
	if playlistType == "series" {
		return true
	}
	return classification == string(ClassSeries) || classification == string(ClassAnime)
}

// EpisodeNumbering is the season and episode of one video, each nil when unknown.
type EpisodeNumbering struct {
	Season  *int
	Episode *int
}

// AutomaticEpisodeNumbering resolves the automatic numbering of a video:
// capture provenance first, then the file name.
func AutomaticEpisodeNumbering(video VideoEntry) EpisodeNumbering {
	if video.Series != nil {
		season := defaultSeasonNumber
		if video.Series.Season != nil {
			season = *video.Series.Season
		}
		episode := video.Series.Episode
		return EpisodeNumbering{Season: &season, Episode: &episode}
	}
	season, episode, ok := ParseSeasonEpisode(video.Name)
	if !ok {
		return EpisodeNumbering{}
	}
	return EpisodeNumbering{Season: &season, Episode: &episode}
}
