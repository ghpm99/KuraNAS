package video

import (
	"regexp"
	"strings"
)

const webVttHeader = "WEBVTT"

var (
	srtTimestampCommaPattern = regexp.MustCompile(`(\d+:\d{2}:\d{2}),(\d{1,3})`)
	srtSequenceNumberPattern = regexp.MustCompile(`^\d+$`)
	blankLinesSeparator      = regexp.MustCompile(`\n[ \t]*\n+`)
)

func ConvertSubtitleToWebVtt(subtitleText string, subtitleFormat string) string {
	normalizedText := normalizeLineBreaks(subtitleText)
	if subtitleFormat == SubtitleFormatVtt {
		return ensureWebVttHeader(normalizedText)
	}
	return convertSrtToWebVtt(normalizedText)
}

func normalizeLineBreaks(text string) string {
	return strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
}

func ensureWebVttHeader(normalizedText string) string {
	trimmedText := strings.TrimLeft(normalizedText, "\n")
	if strings.HasPrefix(trimmedText, webVttHeader) {
		return trimmedText
	}
	return webVttHeader + "\n\n" + trimmedText
}

func convertSrtToWebVtt(normalizedText string) string {
	cueBlocks := []string{}
	for _, rawBlock := range blankLinesSeparator.Split(strings.TrimSpace(normalizedText), -1) {
		cueBlock, isCue := buildWebVttCueBlock(rawBlock)
		if isCue {
			cueBlocks = append(cueBlocks, cueBlock)
		}
	}
	if len(cueBlocks) == 0 {
		return webVttHeader + "\n"
	}
	return webVttHeader + "\n\n" + strings.Join(cueBlocks, "\n\n") + "\n"
}

func buildWebVttCueBlock(rawBlock string) (string, bool) {
	blockLines := strings.Split(strings.TrimSpace(rawBlock), "\n")
	if len(blockLines) > 0 && srtSequenceNumberPattern.MatchString(strings.TrimSpace(blockLines[0])) {
		blockLines = blockLines[1:]
	}
	if len(blockLines) < 2 || !strings.Contains(blockLines[0], "-->") {
		return "", false
	}
	timingLine := srtTimestampCommaPattern.ReplaceAllString(strings.TrimSpace(blockLines[0]), "$1.$2")
	return strings.Join(append([]string{timingLine}, blockLines[1:]...), "\n"), true
}
