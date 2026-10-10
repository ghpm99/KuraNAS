package playlist

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type pathTokenIndex struct {
	segmentTokens [][]string
}

func newPathTokenIndex(video VideoEntry) pathTokenIndex {
	index := pathTokenIndex{}
	for _, rawPath := range []string{video.Path, video.ParentPath, video.Name} {
		for _, segment := range splitPathSegments(rawPath) {
			tokens := tokenizeWords(segment)
			if len(tokens) > 0 {
				index.segmentTokens = append(index.segmentTokens, tokens)
			}
		}
	}
	return index
}

func (index pathTokenIndex) containsAnyKeyword(keywords []string) bool {
	for _, keyword := range keywords {
		keywordTokens := tokenizeWords(keyword)
		if len(keywordTokens) == 0 {
			continue
		}
		for _, tokens := range index.segmentTokens {
			if containsConsecutiveTokens(tokens, keywordTokens) {
				return true
			}
		}
	}
	return false
}

func splitPathSegments(rawPath string) []string {
	return strings.FieldsFunc(rawPath, func(character rune) bool {
		return character == '/' || character == '\\'
	})
}

func tokenizeWords(text string) []string {
	return strings.FieldsFunc(foldAccents(strings.ToLower(text)), func(character rune) bool {
		return !unicode.IsLetter(character) && !unicode.IsNumber(character)
	})
}

func foldAccents(text string) string {
	var folded strings.Builder
	for _, character := range norm.NFD.String(text) {
		if unicode.Is(unicode.Mn, character) {
			continue
		}
		folded.WriteRune(character)
	}
	return folded.String()
}

func containsConsecutiveTokens(tokens []string, sequence []string) bool {
	for start := 0; start+len(sequence) <= len(tokens); start++ {
		if tokensMatchAt(tokens, sequence, start) {
			return true
		}
	}
	return false
}

func tokensMatchAt(tokens []string, sequence []string, start int) bool {
	for offset, expected := range sequence {
		if tokens[start+offset] != expected {
			return false
		}
	}
	return true
}
