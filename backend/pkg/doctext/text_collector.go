package doctext

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const byteOrderMarkRune = 0xFEFF

type textCollector struct {
	builder strings.Builder
	isFull  bool
}

func (collector *textCollector) add(fragment string) {
	if collector.isFull || fragment == "" {
		return
	}
	remainingBytes := MaxStoredTextBytes - collector.builder.Len()
	if len(fragment) <= remainingBytes {
		collector.builder.WriteString(fragment)
		return
	}
	collector.builder.WriteString(cutAtRuneBoundary(fragment, remainingBytes))
	collector.isFull = true
}

func (collector *textCollector) text() string {
	return strings.TrimSpace(collector.builder.String())
}

func cutAtRuneBoundary(text string, maxBytes int) string {
	if len(text) <= maxBytes {
		return text
	}
	cutIndex := maxBytes
	for cutIndex > 0 && !utf8.RuneStart(text[cutIndex]) {
		cutIndex--
	}
	return text[:cutIndex]
}

func removeUnstorableRunes(text string) string {
	return strings.Map(func(character rune) rune {
		switch {
		case character == '\n' || character == '\t':
			return character
		case character == utf8.RuneError, character == byteOrderMarkRune:
			return -1
		case unicode.IsControl(character):
			return -1
		default:
			return character
		}
	}, text)
}

func finalizeText(text string, isSourceTruncated bool) Result {
	cleanText := strings.TrimSpace(removeUnstorableRunes(text))
	if len(cleanText) <= MaxStoredTextBytes {
		return Result{Text: cleanText, Truncated: isSourceTruncated}
	}
	return Result{Text: strings.TrimSpace(cutAtRuneBoundary(cleanText, MaxStoredTextBytes)), Truncated: true}
}
