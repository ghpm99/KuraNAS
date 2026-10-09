package utils

import (
	"slices"
	"testing"
)

func TestGetFormatTypeByExtensionClassifiesAudioAndVideoContainers(t *testing.T) {
	formatTypeByExtension := map[string]string{
		".mp3": FormatTypeAudio, ".wav": FormatTypeAudio, ".aac": FormatTypeAudio, ".flac": FormatTypeAudio,
		".m4a": FormatTypeAudio, ".ogg": FormatTypeAudio, ".oga": FormatTypeAudio, ".opus": FormatTypeAudio,
		".wma": FormatTypeAudio, ".alac": FormatTypeAudio, ".aiff": FormatTypeAudio, ".aif": FormatTypeAudio,
		".ape": FormatTypeAudio, ".wv": FormatTypeAudio, ".OGG": FormatTypeAudio,
		".ogv": FormatTypeVideo, ".mp4": FormatTypeVideo,
	}
	for extension, expectedType := range formatTypeByExtension {
		if got := GetFormatTypeByExtension(extension).Type; got != expectedType {
			t.Errorf("%s: expected %s, got %s", extension, expectedType, got)
		}
	}
}

func TestAudioFormatListAgreesWithClassification(t *testing.T) {
	for _, extension := range AudioFormats {
		if GetFormatTypeByExtension(extension).Type != FormatTypeAudio {
			t.Errorf("%s listed in AudioFormats but not classified as audio", extension)
		}
	}
	for _, extension := range VideoFormats {
		if slices.Contains(AudioFormats, extension) {
			t.Errorf("%s listed in both AudioFormats and VideoFormats", extension)
		}
	}
}

func TestContentTypeByFormatResolvesEveryAudioFormatExplicitly(t *testing.T) {
	expectedContentTypeByExtension := map[string]string{
		".mp3": "audio/mpeg", ".wav": "audio/wav", ".aac": "audio/aac", ".flac": "audio/flac",
		".m4a": "audio/mp4", ".alac": "audio/mp4", ".ogg": "audio/ogg", ".oga": "audio/ogg",
		".opus": "audio/opus", ".wma": "audio/x-ms-wma", ".aiff": "audio/aiff", ".aif": "audio/aiff",
		".ape": "audio/x-ape", ".wv": "audio/x-wavpack",
	}
	if len(expectedContentTypeByExtension) != len(AudioFormats) {
		t.Fatalf("expected one MIME per audio format, got %d for %d formats", len(expectedContentTypeByExtension), len(AudioFormats))
	}
	for extension, expectedContentType := range expectedContentTypeByExtension {
		if got := ContentTypeByFormat(extension, "application/octet-stream"); got != expectedContentType {
			t.Errorf("%s: expected %s, got %s", extension, expectedContentType, got)
		}
		if got := ContentTypeByFormat(extension[1:], "application/octet-stream"); got != expectedContentType {
			t.Errorf("%s without dot: expected %s, got %s", extension, expectedContentType, got)
		}
	}
	if got := ContentTypeByFormat(".M4A", "application/octet-stream"); got != "audio/mp4" {
		t.Errorf("uppercase extension: got %s", got)
	}
}
