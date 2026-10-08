package video

type VideoSummaryDto struct {
	Duration   string  `json:"duration"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	CodecName  string  `json:"codec_name"`
	FrameRate  float64 `json:"frame_rate"`
	BitRate    string  `json:"bit_rate"`
	AudioCodec string  `json:"audio_codec"`
	FormatName string  `json:"format_name"`
}
