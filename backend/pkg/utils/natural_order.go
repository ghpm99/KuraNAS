package utils

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

type naturalChunk struct {
	text      string
	isNumeric bool
}

func CompareNaturalOrder(left, right string) int {
	leftChunks := splitNaturalChunks(left)
	rightChunks := splitNaturalChunks(right)

	for index := 0; index < len(leftChunks) && index < len(rightChunks); index++ {
		if comparison := compareNaturalChunks(leftChunks[index], rightChunks[index]); comparison != 0 {
			return comparison
		}
	}
	return len(leftChunks) - len(rightChunks)
}

func NaturalOrderLess(left, right string) bool {
	return CompareNaturalOrder(left, right) < 0
}

func compareNaturalChunks(left, right naturalChunk) int {
	if left.isNumeric && right.isNumeric {
		return compareDigitStrings(left.text, right.text)
	}
	if left.isNumeric != right.isNumeric {
		if left.isNumeric {
			return -1
		}
		return 1
	}
	return strings.Compare(left.text, right.text)
}

func compareDigitStrings(left, right string) int {
	left = strings.TrimLeft(left, "0")
	right = strings.TrimLeft(right, "0")
	if len(left) != len(right) {
		return len(left) - len(right)
	}
	return strings.Compare(left, right)
}

func splitNaturalChunks(text string) []naturalChunk {
	var chunks []naturalChunk
	var current strings.Builder
	isCurrentNumeric := false

	flush := func() {
		if current.Len() == 0 {
			return
		}
		chunks = append(chunks, naturalChunk{text: current.String(), isNumeric: isCurrentNumeric})
		current.Reset()
	}

	for _, character := range foldForNaturalOrder(text) {
		isDigit := unicode.IsDigit(character)
		if current.Len() > 0 && isDigit != isCurrentNumeric {
			flush()
		}
		isCurrentNumeric = isDigit
		current.WriteRune(character)
	}
	flush()
	return chunks
}

func foldForNaturalOrder(text string) string {
	var folded strings.Builder
	for _, decomposedRune := range norm.NFD.String(strings.ToLower(text)) {
		if unicode.Is(unicode.Mn, decomposedRune) {
			continue
		}
		folded.WriteRune(decomposedRune)
	}
	return folded.String()
}
