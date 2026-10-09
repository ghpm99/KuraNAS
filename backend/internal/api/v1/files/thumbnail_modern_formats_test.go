package files

import (
	"bytes"
	"errors"
	"fmt"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"nas-go/api/internal/config"
)

type stillConverterStub struct {
	jpegBytes []byte
	err       error
	sources   []string
}

func (stub *stillConverterStub) ConvertToJPEG(sourcePath string) ([]byte, error) {
	stub.sources = append(stub.sources, sourcePath)
	return stub.jpegBytes, stub.err
}

func uniqueFileIDWithCacheCleanup(t *testing.T) int {
	t.Helper()
	fileID := int(time.Now().UnixNano() % 1_000_000_000)
	t.Cleanup(func() {
		staleFiles, _ := filepath.Glob(filepath.Join(config.GetBuildConfig("ThumbnailPath"), fmt.Sprintf("%d_*", fileID)))
		for _, staleFile := range staleFiles {
			_ = os.Remove(staleFile)
		}
	})
	return fileID
}

func photoJPEGBytes(t *testing.T, width, height int) []byte {
	t.Helper()
	photoPath := filepath.Join(t.TempDir(), "source.jpg")
	writeSourceImage(t, photoPath, width, height, 255, true)
	content, err := os.ReadFile(photoPath)
	if err != nil {
		t.Fatal(err)
	}
	return content
}

func TestFileService_ModernFormatThumbnails(t *testing.T) {
	conversionFailure := errors.New("ffmpeg missing")
	photo := photoJPEGBytes(t, 400, 200)
	rawWithEmbeddedPreview := append(append([]byte("RAWHEADER"), bytes.Repeat([]byte{7}, 300)...), photo...)

	testCases := []struct {
		name                   string
		extension              string
		content                []byte
		converter              *stillConverterStub
		expectJPEGThumbnail    bool
		expectedConversionRuns int
	}{
		{name: "jfif", extension: ".jfif", content: photo, converter: &stillConverterStub{}, expectJPEGThumbnail: true},
		{name: "heic converted", extension: ".heic", content: []byte("heic"), converter: &stillConverterStub{jpegBytes: photo}, expectJPEGThumbnail: true, expectedConversionRuns: 1},
		{name: "heif converted", extension: ".heif", content: []byte("heif"), converter: &stillConverterStub{jpegBytes: photo}, expectJPEGThumbnail: true, expectedConversionRuns: 1},
		{name: "avif converted", extension: ".avif", content: []byte("avif"), converter: &stillConverterStub{jpegBytes: photo}, expectJPEGThumbnail: true, expectedConversionRuns: 1},
		{name: "heic without ffmpeg falls back to icon", extension: ".heic", content: []byte("heic"), converter: &stillConverterStub{err: conversionFailure}, expectedConversionRuns: 1},
		{name: "raw embedded preview", extension: ".cr3", content: rawWithEmbeddedPreview, converter: &stillConverterStub{}, expectJPEGThumbnail: true},
		{name: "raw without preview falls back to icon", extension: ".nef", content: []byte("no jpeg inside"), converter: &stillConverterStub{}},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			sourcePath := filepath.Join(t.TempDir(), "source"+testCase.extension)
			if err := os.WriteFile(sourcePath, testCase.content, 0644); err != nil {
				t.Fatal(err)
			}
			ensureTestIcons(t)
			service := newFilesServiceForTest(t, &filesRepoMock{})
			service.StillConverter = testCase.converter
			fileDto := FileDto{ID: uniqueFileIDWithCacheCleanup(t), Path: sourcePath, Type: File, Format: testCase.extension}

			thumbnail, err := service.GetFileThumbnail(fileDto, 100, 100)
			if err != nil {
				t.Fatal(err)
			}

			if len(testCase.converter.sources) != testCase.expectedConversionRuns {
				t.Fatalf("expected %d conversions, got %d", testCase.expectedConversionRuns, len(testCase.converter.sources))
			}
			if testCase.expectJPEGThumbnail {
				decoded, err := jpeg.Decode(bytes.NewReader(thumbnail))
				if err != nil || decoded.Bounds().Dx() != 100 || decoded.Bounds().Dy() != 50 {
					t.Fatalf("expected 100x50 jpeg thumbnail, got err=%v", err)
				}
				return
			}
			if _, err := png.Decode(bytes.NewReader(thumbnail)); err != nil {
				t.Fatalf("expected generic icon png, got %v", err)
			}
		})
	}
}

func TestFileService_UsesDefaultFFmpegConverterWhenNoneInjected(t *testing.T) {
	service := newFilesServiceForTest(t, &filesRepoMock{})

	first := service.stillImageConverter()
	if first == nil || service.stillImageConverter() != first {
		t.Fatal("expected a stable default converter")
	}
}
