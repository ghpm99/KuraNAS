package search

import (
	"strings"
	"unicode/utf8"

	"nas-go/api/pkg/utils"

	"github.com/lib/pq"
)

const maxSearchTerms = 5

type nameMatch struct {
	DrivingPattern  string
	AllPatterns     []string
	ExactName       string
	PrefixPattern   string
	ContainsPattern string
}

func splitSearchTerms(query string) []string {
	seenTerms := make(map[string]bool)
	terms := make([]string, 0, maxSearchTerms)
	for _, term := range strings.Fields(strings.ToLower(query)) {
		if seenTerms[term] {
			continue
		}
		seenTerms[term] = true
		terms = append(terms, term)
		if len(terms) == maxSearchTerms {
			break
		}
	}
	return terms
}

func findLongestTerm(terms []string) string {
	longestTerm := terms[0]
	for _, term := range terms[1:] {
		if utf8.RuneCountInString(term) > utf8.RuneCountInString(longestTerm) {
			longestTerm = term
		}
	}
	return longestTerm
}

func buildNameMatch(query string) (nameMatch, bool) {
	terms := splitSearchTerms(query)
	if len(terms) == 0 {
		return nameMatch{}, false
	}

	allPatterns := make([]string, 0, len(terms))
	for _, term := range terms {
		allPatterns = append(allPatterns, utils.BuildContainsLikePattern(term))
	}

	normalizedQuery := strings.Join(strings.Fields(strings.ToLower(query)), " ")
	return nameMatch{
		DrivingPattern:  utils.BuildContainsLikePattern(findLongestTerm(terms)),
		AllPatterns:     allPatterns,
		ExactName:       normalizedQuery,
		PrefixPattern:   utils.BuildPrefixLikePattern(normalizedQuery),
		ContainsPattern: utils.BuildContainsLikePattern(normalizedQuery),
	}, true
}

func (match nameMatch) allPatternsArg() any {
	return pq.Array(match.AllPatterns)
}
