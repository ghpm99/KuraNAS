package img

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

type stillConverterStub struct {
	jpegBytes []byte
	err       error
	callCount int
}

func (stub *stillConverterStub) ConvertToJPEG(string) ([]byte, error) {
	stub.callCount++
	return stub.jpegBytes, stub.err
}

func writeTempFile(t *testing.T, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOpenPreviewPerFormat(t *testing.T) {
	previewJPEG := encodedTestJPEG(t, 60, 40)
	rawContent := tiffFixture{byteOrder: binary.LittleEndian, orientation: 8, subIfdPreviews: [][]byte{previewJPEG}}.build()

	testCases := []struct {
		name                string
		extension           string
		content             []byte
		converter           *stillConverterStub
		expectedOrientation int
		expectedConversions int
	}{
		{name: "jpeg", extension: ".jpg", content: previewJPEG, converter: &stillConverterStub{}, expectedOrientation: 1},
		{name: "jfif decodes as jpeg", extension: ".jfif", content: previewJPEG, converter: &stillConverterStub{}, expectedOrientation: 1},
		{name: "raw cr2 uses embedded preview and raw orientation", extension: ".cr2", content: rawContent, converter: &stillConverterStub{}, expectedOrientation: 8},
		{name: "raw uppercase extension", extension: ".NEF", content: rawContent, converter: &stillConverterStub{}, expectedOrientation: 8},
		{name: "heic goes through converter", extension: ".heic", content: []byte("ignored"), converter: &stillConverterStub{jpegBytes: previewJPEG}, expectedOrientation: 1, expectedConversions: 1},
		{name: "avif goes through converter", extension: ".avif", content: []byte("ignored"), converter: &stillConverterStub{jpegBytes: previewJPEG}, expectedOrientation: 1, expectedConversions: 1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			path := writeTempFile(t, "source", testCase.content)

			preview, err := OpenPreview(path, testCase.extension, testCase.converter)
			if err != nil {
				t.Fatal(err)
			}
			if preview.Image.Bounds().Dx() != 60 || preview.Image.Bounds().Dy() != 40 {
				t.Fatalf("unexpected preview size %v", preview.Image.Bounds())
			}
			if preview.Orientation != testCase.expectedOrientation {
				t.Fatalf("expected orientation %d, got %d", testCase.expectedOrientation, preview.Orientation)
			}
			if testCase.converter.callCount != testCase.expectedConversions {
				t.Fatalf("expected %d conversions, got %d", testCase.expectedConversions, testCase.converter.callCount)
			}
		})
	}
}

func TestOpenPreviewFailures(t *testing.T) {
	conversionFailure := errors.New("ffmpeg unavailable")
	path := writeTempFile(t, "source", []byte("not an image"))

	if _, err := OpenPreview(path, ".heic", &stillConverterStub{err: conversionFailure}); !errors.Is(err, conversionFailure) {
		t.Fatalf("expected converter error, got %v", err)
	}
	if _, err := OpenPreview(path, ".heic", &stillConverterStub{jpegBytes: []byte("not jpeg")}); err == nil {
		t.Fatal("expected decode error for invalid converter output")
	}
	if _, err := OpenPreview(path, ".dng", &stillConverterStub{}); !errors.Is(err, ErrNoEmbeddedPreview) {
		t.Fatalf("expected ErrNoEmbeddedPreview, got %v", err)
	}
	if _, err := OpenPreview(path, ".png", &stillConverterStub{}); err == nil {
		t.Fatal("expected decode error for corrupt png")
	}
}

func TestPreviewDimensionsPerFormat(t *testing.T) {
	previewJPEG := encodedTestJPEG(t, 60, 40)
	rawPath := writeTempFile(t, "raw", tiffFixture{byteOrder: binary.LittleEndian, orientation: 1, subIfdPreviews: [][]byte{previewJPEG}}.build())
	jpegPath := writeTempFile(t, "plain", previewJPEG)
	converter := &stillConverterStub{jpegBytes: previewJPEG}

	for name, path := range map[string]string{".arw": rawPath, ".jpg": jpegPath, ".heif": jpegPath} {
		width, height, err := PreviewDimensions(path, name, converter)
		if err != nil || width != 60 || height != 40 {
			t.Fatalf("%s: expected 60x40, got %dx%d err=%v", name, width, height, err)
		}
	}

	if _, _, err := PreviewDimensions(jpegPath, ".heic", &stillConverterStub{err: errors.New("no ffmpeg")}); err == nil {
		t.Fatal("expected converter error")
	}
	if _, _, err := PreviewDimensions(jpegPath, ".heic", &stillConverterStub{jpegBytes: []byte("x")}); err == nil {
		t.Fatal("expected decode error")
	}
	if _, _, err := PreviewDimensions(filepath.Join(t.TempDir(), "missing.jpg"), ".jpg", converter); err == nil {
		t.Fatal("expected missing file error")
	}
	if _, _, err := PreviewDimensions(writeTempFile(t, "bad", []byte("x")), ".dng", converter); !errors.Is(err, ErrNoEmbeddedPreview) {
		t.Fatalf("expected ErrNoEmbeddedPreview, got %v", err)
	}
}
