package doctext

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	byteOrderMarkRune = 0xFEFF
	pageBreakRune     = '\f'
)

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

func removeUnstorableRunes(text string, keepPageBreaks bool) string {
	return strings.Map(func(character rune) rune {
		switch {
		case character == '\n' || character == '\t':
			return character
		case character == pageBreakRune && keepPageBreaks:
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
	return finalizeStoredText(text, isSourceTruncated, false)
}

func finalizePagedText(text string, isSourceTruncated bool) Result {
	return finalizeStoredText(text, isSourceTruncated, true)
}

func finalizeStoredText(text string, isSourceTruncated bool, keepPageBreaks bool) Result {
	cleanText := trimStoredText(removeUnstorableRunes(text, keepPageBreaks), keepPageBreaks)
	if len(cleanText) <= MaxStoredTextBytes {
		return Result{Text: cleanText, Truncated: isSourceTruncated}
	}
	return Result{Text: trimStoredText(cutAtRuneBoundary(cleanText, MaxStoredTextBytes), keepPageBreaks), Truncated: true}
}

func trimStoredText(text string, keepPageBreaks bool) string {
	if !keepPageBreaks {
		return strings.TrimSpace(text)
	}
	withoutLeadingSpace := strings.TrimLeftFunc(text, func(character rune) bool {
		return unicode.IsSpace(character) && character != pageBreakRune
	})
	return strings.TrimRightFunc(withoutLeadingSpace, unicode.IsSpace)
}

func hasVisibleText(text string) bool {
	return strings.TrimFunc(text, unicode.IsSpace) != ""
}
