package video

import "strconv"

const (
	maxTranscodeHeight    = 1080
	transcodeAudioBitrate = "160k"
)

func ClampTranscodeHeight(requestedHeight int, sourceHeight int) int {
	ceiling := maxTranscodeHeight
	if requestedHeight > 0 && requestedHeight < ceiling {
		ceiling = requestedHeight
	}
	if sourceHeight > 0 && sourceHeight < ceiling {
		ceiling = sourceHeight
	}
	return ceiling - ceiling%2
}

func buildTranscodeScaleFilter(requestedHeight int, sourceHeight int) string {
	targetHeight := ClampTranscodeHeight(requestedHeight, sourceHeight)
	if sourceHeight > 0 {
		return "scale=-2:" + strconv.Itoa(targetHeight)
	}
	return "scale=-2:min(" + strconv.Itoa(targetHeight) + "\\,ih)"
}

func BuildTranscodeArguments(sourcePath string, startSeconds float64, requestedHeight int, sourceHeight int) []string {
	arguments := []string{"-hide_banner", "-loglevel", "error"}
	if startSeconds > 0 {
		arguments = append(arguments, "-ss", formatSeconds(startSeconds))
	}
	return append(arguments,
		"-i", sourcePath,
		"-map", "0:v:0",
		"-map", "0:a:0?",
		"-c:v", "libx264",
		"-preset", "veryfast",
		"-crf", "23",
		"-vf", buildTranscodeScaleFilter(requestedHeight, sourceHeight),
		"-pix_fmt", "yuv420p",
		"-c:a", "aac",
		"-b:a", transcodeAudioBitrate,
		"-ac", "2",
		"-movflags", "frag_keyframe+empty_moov+default_base_moof",
		"-f", "mp4",
		"pipe:1",
	)
}
