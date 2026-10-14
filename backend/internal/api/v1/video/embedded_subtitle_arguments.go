package video

import "strconv"

func buildEmbeddedSubtitleProbeArguments(sourcePath string) []string {
	return []string{
		"-v", "error",
		"-select_streams", "s",
		"-show_streams",
		"-of", "json",
		sourcePath,
	}
}

func buildEmbeddedSubtitleExtractArguments(sourcePath string, streamIndex int, outputPath string) []string {
	return []string{
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-i", sourcePath,
		"-map", "0:" + strconv.Itoa(streamIndex),
		"-f", "webvtt",
		outputPath,
	}
}
