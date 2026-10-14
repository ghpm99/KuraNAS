package playlist

import (
	"sort"
	"strings"

	"nas-go/api/pkg/i18n"
)

type NameKeyMatches struct {
	Exact    []string
	Prefix   []string
	Contains []string
}

func MatchNameKeys(searchTerm string) NameKeyMatches {
	foldedTerm := foldForSearch(searchTerm)
	matches := NameKeyMatches{Exact: []string{}, Prefix: []string{}, Contains: []string{}}
	if foldedTerm == "" {
		return matches
	}
	for _, nameKey := range generatedNameKeys() {
		foldedName := foldForSearch(i18n.GetMessage(nameKey))
		if foldedName == foldedTerm {
			matches.Exact = append(matches.Exact, nameKey)
		}
		if strings.HasPrefix(foldedName, foldedTerm) {
			matches.Prefix = append(matches.Prefix, nameKey)
		}
		if strings.Contains(foldedName, foldedTerm) {
			matches.Contains = append(matches.Contains, nameKey)
		}
	}
	return matches
}

func generatedNameKeys() []string {
	nameKeys := []string{NameKeyContinueWatching, NameKeyRecommended, NameKeyMixed, NameKeyUnknownSeries, NameKeyVideos}
	for _, classificationNameKey := range classificationNameKeys {
		nameKeys = append(nameKeys, classificationNameKey)
	}
	sort.Strings(nameKeys)
	return nameKeys
}

func foldForSearch(text string) string {
	return strings.Join(strings.Fields(strings.ToLower(foldAccents(text))), " ")
}
