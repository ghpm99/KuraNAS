package utils

import "testing"

func TestNormalizeExtensionLowercases(t *testing.T) {
	if got := NormalizeExtension(".JpEg"); got != ".jpeg" {
		t.Fatalf("got %q", got)
	}
}

func TestExtensionOfLowercasesFileExtension(t *testing.T) {
	cases := map[string]string{
		"IMG_0001.JPG":   ".jpg",
		"Song.MP3":       ".mp3",
		"clip.Mp4":       ".mp4",
		"archive.tar.GZ": ".gz",
		"README":         "",
	}
	for fileName, expectedExtension := range cases {
		if got := ExtensionOf(fileName); got != expectedExtension {
			t.Fatalf("ExtensionOf(%q) = %q, want %q", fileName, got, expectedExtension)
		}
	}
}

func TestGetFormatTypeByExtensionIsCaseInsensitive(t *testing.T) {
	cases := map[string]string{
		".JPG":  FormatTypeImage,
		".Mp3":  FormatTypeAudio,
		".MP4":  FormatTypeVideo,
		".PDF":  FormatTypeDocument,
		".ZZZ9": FormatTypeUnknown,
	}
	for extension, expectedType := range cases {
		if got := GetFormatTypeByExtension(extension).Type; got != expectedType {
			t.Fatalf("type of %q = %q, want %q", extension, got, expectedType)
		}
	}
}
