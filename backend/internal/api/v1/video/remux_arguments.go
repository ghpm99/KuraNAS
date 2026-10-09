package video

import "strings"

const remuxContentType = "video/mp4"

var remuxableVideoCodecs = map[string]bool{
	"h264": true,
	"vp9":  true,
	"av1":  true,
}

func IsRemuxableVideoCodec(codecName string) bool {
	return remuxableVideoCodecs[strings.ToLower(strings.TrimSpace(codecName))]
}

func BuildRemuxArguments(sourcePath string, startSeconds float64) []string {
	arguments := []string{"-hide_banner", "-loglevel", "error"}
	if startSeconds > 0 {
		arguments = append(arguments, "-ss", formatSeconds(startSeconds))
	}
	return append(arguments,
		"-i", sourcePath,
		"-map", "0:v:0",
		"-map", "0:a:0?",
		"-c", "copy",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof",
		"-f", "mp4",
		"pipe:1",
	)
}
