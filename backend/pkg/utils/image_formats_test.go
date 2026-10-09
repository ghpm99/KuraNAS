package utils

import (
	"slices"
	"strings"
	"testing"
)

func TestModernAndRawPhotoFormatsAreImages(t *testing.T) {
	extensions := []string{
		".heic", ".heif", ".avif", ".tif", ".tiff", ".jfif",
		".cr2", ".cr3", ".nef", ".arw", ".dng", ".orf", ".rw2", ".raf", ".srw", ".pef",
	}
	for _, extension := range extensions {
		if !slices.Contains(ImageFormats, extension) {
			t.Fatalf("%s missing from ImageFormats", extension)
		}
		for _, variant := range []string{extension, strings.ToUpper(extension)} {
			formatType := GetFormatTypeByExtension(variant)
			if formatType.Type != FormatTypeImage || formatType.Mime == "" || formatType.Description == "" {
				t.Fatalf("%s: expected image type with mime and description, got %+v", variant, formatType)
			}
		}
	}
}

func TestEveryImageFormatIsDetectedAsImage(t *testing.T) {
	for _, extension := range ImageFormats {
		if GetFormatTypeByExtension(extension).Type != FormatTypeImage {
			t.Fatalf("%s listed in ImageFormats but not detected as image", extension)
		}
	}
}

func TestFormatSpecificMimeTypes(t *testing.T) {
	expectedMimes := map[string]string{
		".jfif": "image/jpeg",
		".tif":  "image/tiff",
		".tiff": "image/tiff",
		".heic": "image/heic",
		".heif": "image/heif",
		".avif": "image/avif",
		".cr2":  "image/x-raw",
	}
	for extension, expectedMime := range expectedMimes {
		if got := GetFormatTypeByExtension(extension).Mime; got != expectedMime {
			t.Fatalf("%s mime = %q, want %q", extension, got, expectedMime)
		}
	}
}

func TestPhotoFamilyPredicates(t *testing.T) {
	if !IsRawPhotoExtension(".CR3") || IsRawPhotoExtension(".heic") || IsRawPhotoExtension(".jpg") {
		t.Fatal("unexpected raw predicate result")
	}
	if !IsHeifFamilyExtension(".HEIC") || !IsHeifFamilyExtension(".avif") || IsHeifFamilyExtension(".cr2") {
		t.Fatal("unexpected heif predicate result")
	}
}

func TestDocumentAndArchiveFormatsMatchTheirFormatType(t *testing.T) {
	for _, extension := range DocumentFormats {
		if GetFormatTypeByExtension(extension).Type != FormatTypeDocument {
			t.Fatalf("%s listed in DocumentFormats but not detected as document", extension)
		}
	}
	for _, extension := range ArchiveFormats {
		if GetFormatTypeByExtension(extension).Type != FormatTypeArchive {
			t.Fatalf("%s listed in ArchiveFormats but not detected as archive", extension)
		}
	}
}
