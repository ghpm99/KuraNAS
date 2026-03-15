package files

import "testing"

func TestClassifyImage(t *testing.T) {
	tests := []struct {
		name     string
		file     FileDto
		metadata ImageMetadataModel
		category ImageClassificationCategory
		minScore float64
	}{
		{
			name: "detects capture from filename",
			file: FileDto{
				Name: "Screenshot_2026-03-14.png",
				Path: "/library/screens/Screenshot_2026-03-14.png",
			},
			category: ImageClassificationCategoryCapture,
			minScore: 0.9,
		},
		{
			name: "detects photo from exif evidence",
			file: FileDto{
				Name: "IMG_0001.jpg",
				Path: "/storage/DCIM/Camera/IMG_0001.jpg",
			},
			metadata: ImageMetadataModel{
				Make:             "Sony",
				Model:            "A7",
				LensModel:        "FE 35mm",
				DateTimeOriginal: "2026:03:14 12:00:00",
				ISO:              200,
				FocalLength:      35,
			},
			category: ImageClassificationCategoryPhoto,
			minScore: 0.8,
		},
		{
			name: "falls back to other",
			file: FileDto{
				Name: "wallpaper.png",
				Path: "/downloads/wallpaper.png",
			},
			category: ImageClassificationCategoryOther,
			minScore: 0.3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			classification := ClassifyImage(tc.file, tc.metadata)
			if classification.Category != tc.category {
				t.Fatalf("expected category %s, got %s", tc.category, classification.Category)
			}
			if classification.Confidence < tc.minScore {
				t.Fatalf("expected confidence >= %.2f, got %.2f", tc.minScore, classification.Confidence)
			}
		})
	}
}
