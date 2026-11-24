package video

import "strconv"

const defaultAudioStreamIndex = -1

func buildAudioTrackProbeArguments(sourcePath string) []string {
	return []string{
		"-v", "error",
		"-select_streams", "a",
		"-show_streams",
		"-of", "json",
		sourcePath,
	}
}

func buildAudioMapArguments(audioStreamIndex int) []string {
	if audioStreamIndex == defaultAudioStreamIndex {
		return []string{"-map", "0:a:0?"}
	}
	return []string{"-map", "0:" + strconv.Itoa(audioStreamIndex)}
}
