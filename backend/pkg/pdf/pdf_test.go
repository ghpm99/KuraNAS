package pdf

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func pdfAssetPath(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("failed to resolve caller path")
	}
	return filepath.Join(filepath.Dir(filename), "pdf.jpg")
}

func TestThumbnailWithoutAsset(t *testing.T) {
	asset := pdfAssetPath(t)
	backup := asset + ".bak_test"
	_ = os.Remove(backup)
	if _, err := os.Stat(asset); err == nil {
		if err := os.Rename(asset, backup); err != nil {
			t.Fatalf("failed to move existing asset: %v", err)
		}
		t.Cleanup(func() {
			_ = os.Rename(backup, asset)
		})
	}

	img, err := Thumbnail()
	if err == nil {
		t.Fatalf("expected error when pdf.jpg is missing")
	}
	if img != nil {
		t.Fatalf("expected nil image on error")
	}
}

func TestThumbnailWithAsset(t *testing.T) {
	asset := pdfAssetPath(t)
	_ = os.Remove(asset)
	t.Cleanup(func() { _ = os.Remove(asset) })

	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	img.Set(1, 0, color.RGBA{R: 0, G: 255, B: 0, A: 255})
	img.Set(0, 1, color.RGBA{R: 0, G: 0, B: 255, A: 255})
	img.Set(1, 1, color.RGBA{R: 255, G: 255, B: 0, A: 255})

	f, err := os.Create(asset)
	if err != nil {
		t.Fatalf("failed to create asset: %v", err)
	}
	if err := jpeg.Encode(f, img, nil); err != nil {
		_ = f.Close()
		t.Fatalf("failed to encode jpeg: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close jpeg file: %v", err)
	}

	thumb, err := Thumbnail()
	if err != nil {
		t.Fatalf("expected thumbnail success, got %v", err)
	}
	if thumb == nil {
		t.Fatalf("expected non-nil image")
	}
}
