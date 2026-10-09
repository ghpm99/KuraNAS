package search

import (
	"strings"

	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

type nameMatch struct {
	DrivingPattern  string
	AllPatterns     []string
	ExactName       string
	PrefixPattern   string
	ContainsPattern string
}

func buildNameMatch(query string) (nameMatch, bool) {
	termPatterns, hasTerms := utils.BuildSearchTermPatterns(query)
	if !hasTerms {
		return nameMatch{}, false
	}

	normalizedQuery := strings.Join(strings.Fields(strings.ToLower(query)), " ")
	return nameMatch{
		DrivingPattern:  termPatterns.DrivingPattern,
		AllPatterns:     termPatterns.AllPatterns,
		ExactName:       normalizedQuery,
		PrefixPattern:   utils.BuildPrefixLikePattern(normalizedQuery),
		ContainsPattern: utils.BuildContainsLikePattern(normalizedQuery),
	}, true
}

func (match nameMatch) allPatternsArg() any {
	return pq.Array(match.AllPatterns)
}

func normalizeFuzzyQuery(query string) string {
	return strings.Join(strings.Fields(strings.ToLower(query)), " ")
}
