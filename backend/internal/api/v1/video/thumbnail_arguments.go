package video

import "fmt"

const (
	masterFrameMaxWidth    = 1280
	masterFrameMaxHeight   = 720
	masterFrameJPEGQuality = "2"
)

var masterFrameScaleFilter = fmt.Sprintf("scale=min(%d\\,iw):min(%d\\,ih):force_original_aspect_ratio=decrease", masterFrameMaxWidth, masterFrameMaxHeight)

func buildThumbnailScaleFilter(width, height int) string {
	return fmt.Sprintf("scale=%d:%d:force_original_aspect_ratio=decrease,pad=%d:%d:(ow-iw)/2:(oh-ih)/2:black", width, height, width, height)
}

func buildMasterFrameArguments(sourcePath string, offsetSeconds float64, outputPath string) []string {
	return []string{
		"-hide_banner",
		"-loglevel", "error",
		"-y",
		"-ss", formatSeconds(offsetSeconds),
		"-i", sourcePath,
		"-frames:v", "1",
		"-vf", masterFrameScaleFilter,
		"-q:v", masterFrameJPEGQuality,
		"-update", "1",
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
