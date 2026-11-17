package music

import "strings"

func replaceGenreLabel(genre string, genreKey string, newLabel string) []string {
	newKey := normalizeLookupKey(newLabel)
	replacedLabels := []string{}
	seenKeys := map[string]bool{}
	for _, label := range normalizeGenreLabels(genre) {
		labelKey := normalizeLookupKey(label)
		if labelKey == genreKey {
			label, labelKey = newLabel, newKey
		}
		if seenKeys[labelKey] {
			continue
		}
		seenKeys[labelKey] = true
		replacedLabels = append(replacedLabels, label)
	}
	return replacedLabels
}

func buildGenreRenamePatchGroups(sources []GenreTrackSource, genreKey string, newLabel string) []TrackMetadataPatchGroup {
	groupIndexByGenres := map[string]int{}
	groups := []TrackMetadataPatchGroup{}
	for _, source := range sources {
		replacedLabels := replaceGenreLabel(source.Genre, genreKey, newLabel)
		joinedGenres := strings.Join(replacedLabels, genreSeparator)
		groupIndex, isKnownGroup := groupIndexByGenres[joinedGenres]
		if !isKnownGroup {
			groupIndex = len(groups)
			groupIndexByGenres[joinedGenres] = groupIndex
			groups = append(groups, TrackMetadataPatchGroup{
				Patch: TrackMetadataPatch{Genres: optionalTextList{IsSet: true, Values: replacedLabels}},
			})
		}
		groups[groupIndex].FileIDs = append(groups[groupIndex].FileIDs, source.FileID)
	}
	return groups
}

func validateGenreName(name string) (newLabel string, validationError *fieldValidationError) {
	trimmedName := strings.TrimSpace(name)
	if isTextTooLong(trimmedName) {
		return "", newTextTooLongError("name")
	}
	if strings.ContainsFunc(trimmedName, isGenreSeparator) {
		return "", &fieldValidationError{Field: "name", MessageKey: "ERROR_MUSIC_GENRE_NAME_SEPARATOR"}
	}
	newLabel = normalizeGenreLabel(trimmedName)
	if newLabel == "" {
		return "", &fieldValidationError{Field: "name", MessageKey: "ERROR_MUSIC_GENRE_NAME_EMPTY"}
	}
	return newLabel, nil
}
