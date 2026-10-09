package scan

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"

	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/utils"
)

type stillConverterStub struct {
	jpegBytes []byte
	err       error
}

func (stub stillConverterStub) ConvertToJPEG(string) ([]byte, error) {
	return stub.jpegBytes, stub.err
}

func encodedJPEG(t *testing.T, width, height int) []byte {
	t.Helper()
	var buffer bytes.Buffer
	if err := jpeg.Encode(&buffer, image.NewRGBA(image.Rect(0, 0, width, height)), nil); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func failingScriptRunner(utils.ScriptType, string) (string, error) {
	return "", errors.New("pillow cannot open this file")
}

func TestImageMetadataFallsBackToConvertedStillDimensions(t *testing.T) {
	SetImageStillConverterForTesting(stillConverterStub{jpegBytes: encodedJPEG(t, 40, 30)})
	t.Cleanup(func() { SetImageStillConverterForTesting(nil) })

	metadata, err := getImageMetadata(files.FileDto{ID: 9, Path: "/p/photo.heic", Format: ".heic"}, failingScriptRunner)
	if err != nil {
		t.Fatalf("expected degraded metadata without error, got %v", err)
	}
	if metadata.FileId != 9 || metadata.Width != 40 || metadata.Height != 30 {
		t.Fatalf("expected file 9 with 40x30 from converted still, got %+v", metadata)
	}
}

func TestImageMetadataKeepsScriptDimensionsWhenPresent(t *testing.T) {
	SetImageStillConverterForTesting(stillConverterStub{err: errors.New("must not be called")})
	t.Cleanup(func() { SetImageStillConverterForTesting(nil) })

	runner := func(utils.ScriptType, string) (string, error) {
		return `{"width": 800, "height": 600}`, nil
	}
	metadata, err := getImageMetadata(files.FileDto{ID: 1, Path: "/p/photo.heic", Format: ".heic"}, runner)
	if err != nil || metadata.Width != 800 || metadata.Height != 600 {
		t.Fatalf("expected script dimensions kept, got %+v err=%v", metadata, err)
	}
}

func TestImageMetadataSurvivesWhenNoPreviewIsAvailable(t *testing.T) {
	SetImageStillConverterForTesting(stillConverterStub{err: errors.New("ffmpeg missing")})
	t.Cleanup(func() { SetImageStillConverterForTesting(nil) })

	metadata, err := getImageMetadata(files.FileDto{ID: 4, Path: "/p/photo.avif", Format: ".avif"}, failingScriptRunner)
	if err != nil || metadata.FileId != 4 || metadata.Width != 0 {
		t.Fatalf("expected empty-dimension metadata row without error, got %+v err=%v", metadata, err)
	}
}

func TestImageMetadataReadsDimensionsFromEmbeddedRawPreview(t *testing.T) {
	rawPath := filepath.Join(t.TempDir(), "shot.cr3")
	padding := bytes.Repeat([]byte{0x01}, 512)
	content := append(append(append([]byte{}, padding...), encodedJPEG(t, 64, 48)...), padding...)
	if err := os.WriteFile(rawPath, content, 0644); err != nil {
		t.Fatal(err)
	}

	metadata, err := getImageMetadata(files.FileDto{ID: 5, Path: rawPath, Format: ".cr3"}, failingScriptRunner)
	if err != nil || metadata.Width != 64 || metadata.Height != 48 {
		t.Fatalf("expected 64x48 from embedded preview, got %+v err=%v", metadata, err)
	}
}

func TestImageMetadataPropagatesScriptTimeout(t *testing.T) {
	timeoutRunner := func(utils.ScriptType, string) (string, error) {
		return "", fmt.Errorf("script excedeu o timeout: %w", context.DeadlineExceeded)
	}
	_, err := getImageMetadata(files.FileDto{ID: 6, Path: "/p/a.jpg", Format: ".jpg"}, timeoutRunner)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected deadline error to propagate, got %v", err)
	}
}
