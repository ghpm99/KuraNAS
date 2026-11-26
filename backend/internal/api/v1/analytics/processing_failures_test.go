package analytics

import (
	"testing"

	"nas-go/api/internal/config"
	"nas-go/api/pkg/i18n"
)

func TestProcessingFailureDtoTranslatesErrorPerStageWithoutRawError(t *testing.T) {
	previousLocale := config.AppConfig.Lang
	config.AppConfig.Lang = "pt-BR"
	if err := i18n.LoadTranslations(); err != nil {
		t.Fatalf("load translations: %v", err)
	}
	t.Cleanup(func() {
		config.AppConfig.Lang = previousLocale
		_ = i18n.LoadTranslations()
	})

	stages := []ProcessingStage{ProcessingStageChecksum, ProcessingStageMetadata, ProcessingStageThumbnail, ProcessingStageDocumentText, ProcessingStageAI}
	seenMessages := map[string]bool{}
	for _, stage := range stages {
		failureDto := toProcessingFailureDto(ProcessingFailureModel{FileID: 7, Name: "a.jpg", Stage: stage, RawError: "exec: ffmpeg: signal killed"})

		if failureDto.Error == "" || failureDto.Error == processingFailureMessageKeys[stage] || failureDto.Error == "exec: ffmpeg: signal killed" {
			t.Fatalf("stage %s: error not translated: %q", stage, failureDto.Error)
		}
		if seenMessages[failureDto.Error] {
			t.Fatalf("stage %s: duplicated message %q", stage, failureDto.Error)
		}
		seenMessages[failureDto.Error] = true
		if failureDto.Stage != string(stage) || failureDto.FileID != 7 || failureDto.Name != "a.jpg" {
			t.Fatalf("unexpected dto: %+v", failureDto)
		}
	}
}
