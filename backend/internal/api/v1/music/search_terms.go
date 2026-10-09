package music

import (
	"strings"
	"unicode/utf8"

	"nas-go/api/pkg/utils"
)

const maxSearchTerms = 5

type trackSearchPatterns struct {
	drivingPattern string
	allPatterns    []string
}

func buildTrackSearchPatterns(searchText string) (trackSearchPatterns, bool) {
	seenTerms := make(map[string]bool)
	patterns := trackSearchPatterns{}
	longestTerm := ""
	for _, term := range strings.Fields(strings.ToLower(searchText)) {
		if seenTerms[term] {
			continue
		}
		seenTerms[term] = true
		patterns.allPatterns = append(patterns.allPatterns, utils.BuildContainsLikePattern(term))
		if utf8.RuneCountInString(term) > utf8.RuneCountInString(longestTerm) {
			longestTerm = term
		}
		if len(patterns.allPatterns) == maxSearchTerms {
			break
		}
	}
	if len(patterns.allPatterns) == 0 {
		return trackSearchPatterns{}, false
	}
	patterns.drivingPattern = utils.BuildContainsLikePattern(longestTerm)
	return patterns, true
}
