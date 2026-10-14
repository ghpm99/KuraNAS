package video

func technicalMetadataScanTargets(video *VideoFileModel) []any {
	return []any{
		&video.DurationSeconds,
		&video.Width,
		&video.Height,
		&video.VideoCodec,
	}
}
