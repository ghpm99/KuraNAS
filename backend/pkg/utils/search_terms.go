package utils

import (
	"strings"
	"unicode/utf8"
)

const MaxSearchTerms = 5

type SearchTermPatterns struct {
	DrivingPattern string
	AllPatterns    []string
}

func SplitSearchTerms(query string) []string {
	seenTerms := make(map[string]bool)
	terms := make([]string, 0, MaxSearchTerms)
	for _, term := range strings.Fields(strings.ToLower(query)) {
		if seenTerms[term] {
			continue
		}
		seenTerms[term] = true
		terms = append(terms, term)
		if len(terms) == MaxSearchTerms {
			break
		}
	}
	return terms
}

func BuildSearchTermPatterns(query string) (SearchTermPatterns, bool) {
	terms := SplitSearchTerms(query)
	if len(terms) == 0 {
		return SearchTermPatterns{}, false
	}

	allPatterns := make([]string, 0, len(terms))
	for _, term := range terms {
		allPatterns = append(allPatterns, BuildContainsLikePattern(term))
	}

	return SearchTermPatterns{
		DrivingPattern: BuildContainsLikePattern(findLongestSearchTerm(terms)),
		AllPatterns:    allPatterns,
	}, true
}

func findLongestSearchTerm(terms []string) string {
	longestTerm := terms[0]
	for _, term := range terms[1:] {
		if utf8.RuneCountInString(term) > utf8.RuneCountInString(longestTerm) {
			longestTerm = term
		}
	}
	return longestTerm
}
