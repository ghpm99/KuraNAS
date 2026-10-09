package music

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var spaceNormalizerRegexp = regexp.MustCompile(`\s+`)

type CatalogGroupingKeys struct {
	ArtistKey   string
	ArtistLabel string
	AlbumKey    string
	AlbumLabel  string
	GenreKeys   []string
	GenreLabels []string
}

func BuildCatalogGroupingKeys(artist string, albumArtist string, album string, genre string) CatalogGroupingKeys {
	artistLabel := preferredArtist(albumArtist, artist)
	albumLabel := normalizeText(album)

	genreLabels := normalizeGenreLabels(genre)
	genreKeys := make([]string, 0, len(genreLabels))
	for _, genreLabel := range genreLabels {
		genreKeys = append(genreKeys, normalizeLookupKey(genreLabel))
	}
	if genreLabels == nil {
		genreLabels = []string{}
	}

	return CatalogGroupingKeys{
		ArtistKey:   artistGroupKey(artistLabel),
		ArtistLabel: artistLabel,
		AlbumKey:    albumGroupKey(artistLabel, albumLabel),
		AlbumLabel:  albumLabel,
		GenreKeys:   genreKeys,
		GenreLabels: genreLabels,
	}
}

func artistGroupKey(artistLabel string) string {
	if artistLabel == "" {
		return ""
	}
	return normalizeLookupKey(artistLabel)
}

func albumGroupKey(artistLabel string, albumLabel string) string {
	if artistLabel == "" || albumLabel == "" {
		return ""
	}
	return normalizeLookupKey(artistLabel + "::" + albumLabel)
}

func normalizeText(rawText string) string {
	trimmed := strings.TrimSpace(rawText)
	if trimmed == "" {
		return ""
	}

	return spaceNormalizerRegexp.ReplaceAllString(trimmed, " ")
}

func normalizeLookupKey(rawText string) string {
	normalized := normalizeText(strings.ToLower(rawText))
	normalized = strings.NewReplacer("_", " ", "-", " ", ".", " ").Replace(normalized)
	return spaceNormalizerRegexp.ReplaceAllString(normalized, " ")
}

func normalizeGenreLabel(rawGenre string) string {
	normalized := normalizeLookupKey(rawGenre)

	switch normalized {
	case "r&b", "r & b", "rnb", "rhythm and blues":
		return "R&B"
	case "r&b/soul", "r&b / soul", "rnb/soul", "rnb / soul", "soul/r&b", "soul / r&b":
		return "R&B / Soul"
	case "hip hop", "hiphop", "hip-hop":
		return "Hip-Hop"
	case "lo fi", "lofi", "lo-fi":
		return "Lo-Fi"
	case "soundtrack", "ost":
		return "Soundtrack"
	}

	if normalized == "" {
		return ""
	}

	words := strings.Fields(normalized)
	for index, word := range words {
		if word == "&" {
			continue
		}
		words[index] = capitalizeFirstRune(word)
	}

	return strings.Join(words, " ")
}

func capitalizeFirstRune(word string) string {
	firstRune, firstRuneSize := utf8.DecodeRuneInString(word)
	return string(unicode.ToUpper(firstRune)) + word[firstRuneSize:]
}

func normalizeGenreLabels(rawGenres string) []string {
	normalized := normalizeText(rawGenres)
	if normalized == "" {
		return nil
	}

	genreParts := strings.FieldsFunc(normalized, func(separator rune) bool {
		return separator == ';' || separator == ',' || separator == '|'
	})

	if len(genreParts) == 0 {
		genreParts = []string{normalized}
	}

	labels := make([]string, 0, len(genreParts))
	seenLabels := map[string]bool{}
	for _, genrePart := range genreParts {
		label := normalizeGenreLabel(genrePart)
		if label == "" || seenLabels[label] {
			continue
		}
		seenLabels[label] = true
		labels = append(labels, label)
	}

	return labels
}

func preferredArtist(albumArtist string, artist string) string {
	if albumArtistLabel := normalizeText(albumArtist); albumArtistLabel != "" {
		return albumArtistLabel
	}
	return normalizeText(artist)
}
