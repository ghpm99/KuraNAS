package icons

import "testing"

func TestIconFunctionsResolveAssetsForCurrentBuildConfig(t *testing.T) {
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
