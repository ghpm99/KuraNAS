package doctext

import (
	"html"
	"strings"
)

func extractMarkupText(path string, sizeBytes int64) (Result, error) {
	rawBytes, err := readLeadingBytes(path)
	if err != nil {
		return Result{}, err
	}
	decodedText, err := decodeTextBytes(rawBytes, sizeBytes > MaxReadBytes)
	if err != nil {
		return Result{}, err
	}
	visibleText := html.UnescapeString(stripMarkupTags(decodedText))
	return finalizeText(strings.Join(strings.Fields(visibleText), " "), sizeBytes > MaxReadBytes), nil
}

func stripMarkupTags(source string) string {
	var visibleText strings.Builder
	position := 0
	for position < len(source) {
		tagOffset := strings.IndexByte(source[position:], '<')
		if tagOffset < 0 {
			visibleText.WriteString(source[position:])
			break
		}
		visibleText.WriteString(source[position : position+tagOffset])
		visibleText.WriteByte(' ')
		position = skipMarkupConstruct(source, position+tagOffset)
	}
	return visibleText.String()
}

func skipMarkupConstruct(source string, tagStart int) int {
	switch {
	case strings.HasPrefix(source[tagStart:], "<!--"):
		return skipPast(source, tagStart, "-->")
	case hasTagNamePrefix(source[tagStart:], "<script"):
		return skipPastFold(source, tagStart, "</script")
	case hasTagNamePrefix(source[tagStart:], "<style"):
		return skipPastFold(source, tagStart, "</style")
	default:
		return skipPast(source, tagStart, ">")
	}
}

func hasTagNamePrefix(source string, tagPrefix string) bool {
	if len(source) <= len(tagPrefix) || !strings.EqualFold(source[:len(tagPrefix)], tagPrefix) {
		return false
	}
	nextCharacter := source[len(tagPrefix)]
	return nextCharacter == '>' || nextCharacter == ' ' || nextCharacter == '\t' || nextCharacter == '\n' || nextCharacter == '\r' || nextCharacter == '/'
}

func skipPast(source string, from int, terminator string) int {
	terminatorOffset := strings.Index(source[from:], terminator)
	if terminatorOffset < 0 {
		return len(source)
	}
	return from + terminatorOffset + len(terminator)
}

func skipPastFold(source string, from int, closingTagPrefix string) int {
	for position := from + 1; position+len(closingTagPrefix) <= len(source); position++ {
		if source[position] == '<' && strings.EqualFold(source[position:position+len(closingTagPrefix)], closingTagPrefix) {
			return skipPast(source, position, ">")
		}
	}
	return len(source)
}
