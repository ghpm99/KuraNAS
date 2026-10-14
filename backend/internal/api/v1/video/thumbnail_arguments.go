package video

import "fmt"

func buildThumbnailScaleFilter(width, height int) string {
	return fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black", width, height, width, height)
}

func buildThumbnailFrameArguments(sourcePath string, offsetSeconds float64, width, height int, outputPath string) []string {
	return []string{
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-ss", formatSeconds(offsetSeconds),
		"-i", sourcePath,
		"-frames:v", "1",
		"-vf", buildThumbnailScaleFilter(width, height),
		outputPath,
	}
}

func buildPreviewGifArguments(sourcePath string, offsetSeconds float64, width, height int, outputPath string) []string {
	return []string{
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-ss", formatSeconds(offsetSeconds),
		"-t", formatSeconds(previewClipSeconds),
		"-i", sourcePath,
		"-vf", "fps=4," + buildThumbnailScaleFilter(width, height),
		"-loop", "0",
		outputPath,
	}
}
