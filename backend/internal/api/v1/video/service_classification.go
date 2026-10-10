package video

import (
	"nas-go/api/internal/api/v1/video/playlist"
)

const CurrentVideoClassificationVersion = 2

var persistedVideoClassifier = playlist.NewVideoClassifier()

func ClassifyVideoForPersistence(name string, path string, parentPath string, durationText string, height int) string {
	entry := playlist.VideoEntry{
		Name:       name,
		Path:       path,
		ParentPath: parentPath,
		Meta: &playlist.VideoMeta{
			Duration: parseDurationSeconds(durationText),
			Height:   height,
		},
	}
	return string(persistedVideoClassifier.Classify(entry).Classification)
}

func (metadata *VideoMetadataModel) ApplyClassification(name string, path string, parentPath string) {
	metadata.Classification = ClassifyVideoForPersistence(name, path, parentPath, metadata.Duration, metadata.Height)
	metadata.ClassificationVersion = CurrentVideoClassificationVersion
}
