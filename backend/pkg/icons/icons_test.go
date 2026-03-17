package icons

import (
	"image"
	"image/color"
	"image/png"
	"nas-go/api/internal/config"
	"os"
	"path/filepath"
	"testing"
)

func ensureIconFixtures(t *testing.T) {
	t.Helper()

	t.Setenv("ProgramFiles", filepath.Join(t.TempDir(), "ProgramFiles"))

	iconsPath := config.GetBuildConfig("IconPath")
	if iconsPath == "" {
		t.Fatalf("expected icon path from build config")
	}

	if err := os.MkdirAll(iconsPath, 0o755); err != nil {
		t.Fatalf("expected icon directory creation success, got %v", err)
	}

	for _, name := range []string{"pdf", "folder", "mp3", "mp4", "unknown"} {
		filePath := filepath.Join(iconsPath, name+".png")
		if _, err := os.Stat(filePath); err == nil {
			continue
		}

		file, err := os.Create(filePath)
		if err != nil {
			t.Fatalf("expected fixture file creation success for %s, got %v", name, err)
		}

		icon := image.NewNRGBA(image.Rect(0, 0, 1, 1))
		icon.Set(0, 0, color.NRGBA{R: 109, G: 93, B: 246, A: 255})

		if err := png.Encode(file, icon); err != nil {
			_ = file.Close()
			t.Fatalf("expected png fixture encoding success for %s, got %v", name, err)
		}

		if err := file.Close(); err != nil {
			t.Fatalf("expected fixture file close success for %s, got %v", name, err)
		}
	}
}

func TestIconFunctionsResolveAssetsForCurrentBuildConfig(t *testing.T) {
	ensureIconFixtures(t)

	tests := []struct {
		name string
		fn   func() (interface{}, error)
	}{
		{
			name: "pdf",
			fn: func() (interface{}, error) {
				return PdfIcon()
			},
		},
		{
			name: "folder",
			fn: func() (interface{}, error) {
				return FolderIcon()
			},
		},
		{
			name: "mp3",
			fn: func() (interface{}, error) {
				return Mp3Icon()
			},
		},
		{
			name: "mp4",
			fn: func() (interface{}, error) {
				return Mp4Icon()
			},
		},
		{
			name: "unknown",
			fn: func() (interface{}, error) {
				return Icon()
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			img, err := tc.fn()
			if err != nil {
				t.Fatalf("expected icon resolution success for %s, got %v", tc.name, err)
			}
			if img == nil {
				t.Fatalf("expected non-nil image for %s", tc.name)
			}
		})
	}
}
