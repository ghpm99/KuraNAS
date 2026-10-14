package video

import "strings"

var textSubtitleCodecs = map[string]bool{
	"subrip":   true,
	"ass":      true,
	"ssa":      true,
	"mov_text": true,
	"webvtt":   true,
	"text":     true,
}

func IsTextSubtitleCodec(codecName string) bool {
	return textSubtitleCodecs[strings.ToLower(codecName)]
}
