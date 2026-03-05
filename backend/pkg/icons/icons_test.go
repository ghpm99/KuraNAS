package icons

import "testing"

func TestIconFunctionsReturnErrorWhenAssetMissing(t *testing.T) {
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
			if err == nil {
				t.Fatalf("expected error when icon file does not exist")
			}
			if img != nil {
				t.Fatalf("expected nil image on error")
			}
		})
	}
}
