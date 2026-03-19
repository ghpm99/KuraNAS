package files

import (
	"context"
	"errors"
	"nas-go/api/pkg/ai"
	"testing"
)

type aiServiceMock struct {
	executeFn func(ctx context.Context, req ai.Request) (ai.Response, error)
}

func (m *aiServiceMock) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	return m.executeFn(ctx, req)
}

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

func TestClassifyImageWithAI_NilServiceUsesHeuristic(t *testing.T) {
	file := FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	result := ClassifyImageWithAI(file, ImageMetadataModel{}, nil)
	if result.Category != ImageClassificationCategoryOther {
		t.Fatalf("expected other category, got %s", result.Category)
	}
}

func TestClassifyImageWithAI_HighConfidenceSkipsAI(t *testing.T) {
	file := FileDto{
		Name: "Screenshot_2026-03-14.png",
		Path: "/library/screens/Screenshot_2026-03-14.png",
	}
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			t.Fatal("AI should not be called for high confidence heuristic")
			return ai.Response{}, nil
		},
	}
	result := ClassifyImageWithAI(file, ImageMetadataModel{}, mock)
	if result.Category != ImageClassificationCategoryCapture {
		t.Fatalf("expected capture, got %s", result.Category)
	}
}

func TestClassifyImageWithAI_LowConfidenceCallsAI(t *testing.T) {
	file := FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			if req.TaskType != ai.TaskClassification {
				t.Fatalf("expected classification task, got %s", req.TaskType)
			}
			return ai.Response{Content: `{"category": "landscape", "confidence": 0.85}`}, nil
		},
	}
	result := ClassifyImageWithAI(file, ImageMetadataModel{}, mock)
	if result.Category != ImageClassificationCategoryLandscape {
		t.Fatalf("expected landscape, got %s", result.Category)
	}
	if result.Confidence != 0.85 {
		t.Fatalf("expected 0.85 confidence, got %f", result.Confidence)
	}
}

func TestClassifyImageWithAI_AIErrorFallsBackToHeuristic(t *testing.T) {
	file := FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{}, errors.New("provider timeout")
		},
	}
	result := ClassifyImageWithAI(file, ImageMetadataModel{}, mock)
	if result.Category != ImageClassificationCategoryOther {
		t.Fatalf("expected fallback to other, got %s", result.Category)
	}
}

func TestClassifyImageWithAI_InvalidJSONFallsBackToHeuristic(t *testing.T) {
	file := FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{Content: "not json"}, nil
		},
	}
	result := ClassifyImageWithAI(file, ImageMetadataModel{}, mock)
	if result.Category != ImageClassificationCategoryOther {
		t.Fatalf("expected fallback to other, got %s", result.Category)
	}
}

func TestClassifyImageWithAI_UnknownCategoryFallsBackToHeuristic(t *testing.T) {
	file := FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{Content: `{"category": "unknown_cat", "confidence": 0.9}`}, nil
		},
	}
	result := ClassifyImageWithAI(file, ImageMetadataModel{}, mock)
	if result.Category != ImageClassificationCategoryOther {
		t.Fatalf("expected fallback to other, got %s", result.Category)
	}
}

func TestClassifyImageWithAI_MarkdownCodeFenceStripped(t *testing.T) {
	file := FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{Content: "```json\n{\"category\": \"meme\", \"confidence\": 0.80}\n```"}, nil
		},
	}
	result := ClassifyImageWithAI(file, ImageMetadataModel{}, mock)
	if result.Category != ImageClassificationCategoryMeme {
		t.Fatalf("expected meme, got %s", result.Category)
	}
}

func TestClassifyImageWithAI_InvalidConfidenceDefaultsTo075(t *testing.T) {
	file := FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{Content: `{"category": "art", "confidence": -1}`}, nil
		},
	}
	result := ClassifyImageWithAI(file, ImageMetadataModel{}, mock)
	if result.Category != ImageClassificationCategoryArt {
		t.Fatalf("expected art, got %s", result.Category)
	}
	if result.Confidence != 0.75 {
		t.Fatalf("expected 0.75 default confidence, got %f", result.Confidence)
	}
}

func TestBuildClassificationPrompt(t *testing.T) {
	file := FileDto{Name: "photo.jpg", Path: "/photos/photo.jpg", Format: ".jpg"}
	metadata := ImageMetadataModel{
		Width:            4000,
		Height:           3000,
		Make:             "Canon",
		Model:            "EOS R5",
		Software:         "Lightroom",
		ImageDescription: "A sunset",
	}
	prompt := buildClassificationPrompt(file, metadata)
	if !contains(prompt, "Filename: photo.jpg") {
		t.Fatalf("expected filename in prompt")
	}
	if !contains(prompt, "4000x3000") {
		t.Fatalf("expected dimensions in prompt")
	}
	if !contains(prompt, "Canon EOS R5") {
		t.Fatalf("expected camera in prompt")
	}
}

func TestParseAIClassificationResponse_AllValidCategories(t *testing.T) {
	categories := []string{"capture", "photo", "other", "document", "receipt", "landscape", "portrait", "meme", "art", "screenshot_app"}
	for _, cat := range categories {
		result, err := parseAIClassificationResponse(`{"category": "` + cat + `", "confidence": 0.8}`)
		if err != nil {
			t.Fatalf("unexpected error for category %s: %v", cat, err)
		}
		if string(result.Category) != cat {
			t.Fatalf("expected %s, got %s", cat, result.Category)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsStr(s, substr)
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
