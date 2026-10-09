package documenttext

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	snippetLength        = 160
	snippetLeadingLength = 40
	snippetWordSnapRange = 20
)

// BuildSnippet returns about snippetLength characters of text around the
// earliest occurrence of any term, matching case- and accent-insensitively and
// collapsing whitespace. When no term is found it returns the start of the text.
func BuildSnippet(text string, terms []string) string {
	displayRunes := collapseWhitespace(text)
	if len(displayRunes) <= snippetLength {
		return strings.TrimSpace(string(displayRunes))
	}

	matchStart, matchLength, hasMatch := findEarliestMatch(foldRunes(displayRunes), terms)
	if !hasMatch {
		return strings.TrimSpace(string(displayRunes[:snippetLength]))
	}

	windowStart, windowEnd := centerWindow(len(displayRunes), matchStart, matchLength)
	windowStart, windowEnd = snapToWordBoundaries(displayRunes, windowStart, windowEnd, matchStart, matchLength)
	return strings.TrimSpace(string(displayRunes[windowStart:windowEnd]))
}

func collapseWhitespace(text string) []rune {
	collapsed := make([]rune, 0, len(text))
	isPreviousSpace := true
	for _, character := range text {
		if unicode.IsSpace(character) {
			if !isPreviousSpace {
				collapsed = append(collapsed, ' ')
			}
			isPreviousSpace = true
			continue
		}
		collapsed = append(collapsed, character)
		isPreviousSpace = false
	}
	return collapsed
}

func foldRunes(characters []rune) []rune {
	folded := make([]rune, len(characters))
	for index, character := range characters {
		folded[index] = foldRune(character)
	}
	return folded
}

func foldRune(character rune) rune {
	if character < unicode.MaxASCII {
		return unicode.ToLower(character)
	}
	for _, decomposedRune := range norm.NFD.String(string(character)) {
		if !unicode.Is(unicode.Mn, decomposedRune) {
			return unicode.ToLower(decomposedRune)
		}
	}
	return unicode.ToLower(character)
}

func findEarliestMatch(foldedText []rune, terms []string) (start int, length int, found bool) {
	for _, term := range terms {
		foldedTerm := foldRunes([]rune(term))
		termStart := indexOfRunes(foldedText, foldedTerm)
		if termStart < 0 {
			continue
		}
		if !found || termStart < start {
			start, length, found = termStart, len(foldedTerm), true
		}
	}
	return start, length, found
}

func indexOfRunes(haystack []rune, needle []rune) int {
	if len(needle) == 0 {
		return -1
	}
	lastStart := len(haystack) - len(needle)
	for start := 0; start <= lastStart; start++ {
		if haystack[start] != needle[0] {
			continue
		}
		if runesEqual(haystack[start:start+len(needle)], needle) {
			return start
		}
	}
	return -1
}

func runesEqual(left []rune, right []rune) bool {
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func centerWindow(textLength int, matchStart int, matchLength int) (int, int) {
	windowStart := max(matchStart-snippetLeadingLength, 0)
	windowEnd := min(windowStart+snippetLength, textLength)
	windowStart = max(windowEnd-snippetLength, 0)
	if matchStart+matchLength > windowEnd {
		windowEnd = min(matchStart+matchLength, textLength)
		windowStart = max(windowEnd-snippetLength, 0)
	}
	return windowStart, windowEnd
}

func snapToWordBoundaries(characters []rune, windowStart int, windowEnd int, matchStart int, matchLength int) (int, int) {
	if windowStart > 0 {
		for candidate := windowStart; candidate < min(windowStart+snippetWordSnapRange, matchStart); candidate++ {
			if characters[candidate] == ' ' {
				windowStart = candidate + 1
				break
			}
		}
	}
	if windowEnd < len(characters) {
		for candidate := windowEnd; candidate > max(windowEnd-snippetWordSnapRange, matchStart+matchLength); candidate-- {
			if characters[candidate-1] == ' ' {
				windowEnd = candidate - 1
				break
			}
		}
	}
	return windowStart, windowEnd
}
