package playlist

import (
	"sort"

	"nas-go/api/pkg/utils"
)

// NumberedEpisode pairs a video with the numbering used to order it.
type NumberedEpisode struct {
	VideoID   int
	Name      string
	Numbering EpisodeNumbering
}

// EffectiveNumbering returns the manual override values where present and the
// automatic values everywhere else, field by field.
func EffectiveNumbering(automatic EpisodeNumbering, override EpisodeNumbering) EpisodeNumbering {
	effective := automatic
	if override.Season != nil {
		effective.Season = override.Season
	}
	if override.Episode != nil {
		effective.Episode = override.Episode
	}
	return effective
}

// SortEpisodesByNumbering orders episodes by season, episode, natural name and
// id. Unknown seasons and episodes come last.
func SortEpisodesByNumbering(episodes []NumberedEpisode) {
	sort.SliceStable(episodes, func(left, right int) bool {
		return isEpisodeBefore(episodes[left], episodes[right])
	})
}

func isEpisodeBefore(left, right NumberedEpisode) bool {
	if seasonOrder := compareNullableNumbers(left.Numbering.Season, right.Numbering.Season); seasonOrder != 0 {
		return seasonOrder < 0
	}
	if episodeOrder := compareNullableNumbers(left.Numbering.Episode, right.Numbering.Episode); episodeOrder != 0 {
		return episodeOrder < 0
	}
	if nameOrder := utils.CompareNaturalOrder(left.Name, right.Name); nameOrder != 0 {
		return nameOrder < 0
	}
	return left.VideoID < right.VideoID
}

func compareNullableNumbers(left, right *int) int {
	switch {
	case left == nil && right == nil:
		return 0
	case left == nil:
		return 1
	case right == nil:
		return -1
	case *left < *right:
		return -1
	case *left > *right:
		return 1
	default:
		return 0
	}
}
