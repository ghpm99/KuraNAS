package music

import "errors"

var ErrUnsupportedTranscodeFormat = errors.New("unsupported transcode format")

type TranscodeFormat struct {
	ContentType     string
	EncoderArgument []string
}

var transcodeFormatsByName = map[string]TranscodeFormat{
	"mp3":  {ContentType: "audio/mpeg", EncoderArgument: []string{"-c:a", "libmp3lame", "-b:a", "256k", "-f", "mp3"}},
	"aac":  {ContentType: "audio/aac", EncoderArgument: []string{"-c:a", "aac", "-b:a", "256k", "-f", "adts"}},
	"opus": {ContentType: "audio/ogg", EncoderArgument: []string{"-c:a", "libopus", "-b:a", "192k", "-f", "ogg"}},
}

func ResolveTranscodeFormat(formatName string) (TranscodeFormat, error) {
	transcodeFormat, isSupported := transcodeFormatsByName[formatName]
	if !isSupported {
		return TranscodeFormat{}, ErrUnsupportedTranscodeFormat
	}
	return transcodeFormat, nil
}

func BuildTranscodeArguments(sourcePath string, transcodeFormat TranscodeFormat, startSeconds float64) []string {
	arguments := []string{"-hide_banner", "-loglevel", "error"}
	if startSeconds > 0 {
		arguments = append(arguments, "-ss", formatSeconds(startSeconds))
	}
	arguments = append(arguments, "-i", sourcePath, "-vn", "-map_metadata", "-1")
	arguments = append(arguments, transcodeFormat.EncoderArgument...)
	return append(arguments, "pipe:1")
}
