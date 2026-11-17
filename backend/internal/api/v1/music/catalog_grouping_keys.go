package music

import (
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const CurrentCatalogKeysVersion = 1

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
		genreKeys = append(genreKeys, genreLookupKey(genreLabel))
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

func genreLookupKey(genreLabel string) string {
	diacriticsRemover := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	unaccentedLabel, _, err := transform.String(diacriticsRemover, normalizeLookupKey(genreLabel))
	if err != nil {
		return normalizeLookupKey(genreLabel)
	}
	return unaccentedLabel
}

func knownCompoundGenreLabel(normalizedGenre string) (string, bool) {
	switch normalizedGenre {
	case "r&b", "r & b", "rnb", "rhythm and blues":
		return "R&B", true
	case "r&b/soul", "r&b / soul", "rnb/soul", "rnb / soul", "soul/r&b", "soul / r&b":
		return "R&B / Soul", true
	case "hip hop", "hiphop", "hip-hop":
		return "Hip-Hop", true
	case "lo fi", "lofi", "lo-fi":
		return "Lo-Fi", true
	case "soundtrack", "ost":
		return "Soundtrack", true
	}
	return "", false
}

func normalizeGenreLabel(rawGenre string) string {
	normalized := normalizeLookupKey(rawGenre)

	if knownLabel, isKnown := knownCompoundGenreLabel(normalized); isKnown {
		return knownLabel
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

func isGenreListSeparator(character rune) bool {
	return character == ';' || character == ',' || character == '|'
}

func isGenreSeparator(character rune) bool {
	return isGenreListSeparator(character) || character == '/'
}

func splitGenrePart(genrePart string) []string {
	if _, isKnown := knownCompoundGenreLabel(normalizeLookupKey(genrePart)); isKnown {
		return []string{genrePart}
	}
	return strings.FieldsFunc(genrePart, func(character rune) bool { return character == '/' })
}

func normalizeGenreLabels(rawGenres string) []string {
	normalized := normalizeText(rawGenres)
	if normalized == "" {
		return nil
	}

	genreParts := []string{}
	for _, listEntry := range strings.FieldsFunc(normalized, isGenreListSeparator) {
		genreParts = append(genreParts, splitGenrePart(listEntry)...)
	}

	labels := make([]string, 0, len(genreParts))
	seenKeys := map[string]bool{}
	for _, genrePart := range genreParts {
		label := normalizeGenreLabel(genrePart)
		labelKey := genreLookupKey(label)
		if label == "" || seenKeys[labelKey] {
			continue
		}
		seenKeys[labelKey] = true
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
