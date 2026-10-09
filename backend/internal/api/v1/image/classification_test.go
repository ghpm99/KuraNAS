package image

import (
	"context"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/png"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/pkg/ai"
	"os"
	"path/filepath"
	"testing"
)

func TestSanitizeSuggestedName(t *testing.T) {
	cases := map[string]string{
		"":                  "",
		"   ":               "",
		"Rikka Takanashi":   "Rikka_Takanashi",
		"beach: sunset!!!":  "beach_sunset",
		"  _weird--name_  ": "weird--name",
		"a/b\\c*d":          "a_b_c_d",
	}
	for input, want := range cases {
		if got := sanitizeSuggestedName(input); got != want {
			t.Errorf("sanitizeSuggestedName(%q) = %q, want %q", input, got, want)
		}
	}

	long := make([]byte, 200)
	for i := range long {
		long[i] = 'a'
	}
	if got := sanitizeSuggestedName(string(long)); len(got) > 80 {
		t.Errorf("expected suggested name capped at 80, got len %d", len(got))
	}
}

func TestEncodeImageForAI(t *testing.T) {
	if encodeImageForAI("") != nil {
		t.Fatalf("expected nil for empty path")
	}
	if encodeImageForAI("/does/not/exist.png") != nil {
		t.Fatalf("expected nil for missing file")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "tiny.png")
	rgba := image.NewRGBA(image.Rect(0, 0, 4, 4))
	rgba.Set(0, 0, color.RGBA{R: 255, A: 255})
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create png: %v", err)
	}
	if err := png.Encode(f, rgba); err != nil {
		t.Fatalf("encode png: %v", err)
	}
	f.Close()

	images := encodeImageForAI(path)
	if len(images) != 1 {
		t.Fatalf("expected 1 encoded image, got %d", len(images))
	}
	if _, err := base64.StdEncoding.DecodeString(images[0]); err != nil {
		t.Fatalf("expected valid base64, got error: %v", err)
	}
}

func TestParseAIClassificationResponseSuggestedName(t *testing.T) {
	result, err := parseAIClassificationResponse(`{"category":"art","confidence":0.9,"suggested_name":"Rikka Takanashi"}`)
	if err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}
	if result.Category != ClassificationCategoryArt {
		t.Fatalf("expected art category, got %s", result.Category)
	}
	if result.SuggestedName != "Rikka_Takanashi" {
		t.Fatalf("expected sanitized suggested name, got %q", result.SuggestedName)
	}
}

type aiServiceMock struct {
	executeFn func(ctx context.Context, req ai.Request) (ai.Response, error)
}

func (m *aiServiceMock) Execute(ctx context.Context, req ai.Request) (ai.Response, error) {
	return m.executeFn(ctx, req)
}

func TestClassifyImage(t *testing.T) {
	tests := []struct {
		name     string
		file     files.FileDto
		metadata MetadataModel
		category ClassificationCategory
		minScore float64
	}{
		{
			name: "detects capture from filename",
			file: files.FileDto{
				Name: "Screenshot_2026-03-14.png",
				Path: "/library/screens/Screenshot_2026-03-14.png",
			},
			category: ClassificationCategoryCapture,
			minScore: 0.9,
		},
		{
			name: "detects photo from exif evidence",
			file: files.FileDto{
				Name: "IMG_0001.jpg",
				Path: "/storage/DCIM/Camera/IMG_0001.jpg",
			},
			metadata: MetadataModel{
				Make:             "Sony",
				Model:            "A7",
				LensModel:        "FE 35mm",
				DateTimeOriginal: "2026:03:14 12:00:00",
				ISO:              200,
				FocalLength:      35,
			},
			category: ClassificationCategoryPhoto,
			minScore: 0.8,
		},
		{
			name: "falls back to other",
			file: files.FileDto{
				Name: "wallpaper.png",
				Path: "/downloads/wallpaper.png",
			},
			category: ClassificationCategoryOther,
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

func classifyWithAnswer(t *testing.T, content string) (ClassificationModel, error) {
	t.Helper()
	file := files.FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{Content: content}, nil
		},
	}
	return ClassifyImageByAI(context.Background(), file, MetadataModel{}, mock)
}

func TestClassifyImageByAI_NilServiceReturnsUnavailable(t *testing.T) {
	file := files.FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	_, err := ClassifyImageByAI(context.Background(), file, MetadataModel{}, nil)
	if !errors.Is(err, ErrAIServiceUnavailable) {
		t.Fatalf("expected ErrAIServiceUnavailable, got %v", err)
	}
}

func TestClassifyImageByAI_SuccessMarksClassifiedByAI(t *testing.T) {
	file := files.FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			if req.TaskType != ai.TaskClassification {
				t.Fatalf("expected classification task, got %s", req.TaskType)
			}
			return ai.Response{Content: `{"category": "landscape", "confidence": 0.85}`}, nil
		},
	}
	classification, err := ClassifyImageByAI(context.Background(), file, MetadataModel{}, mock)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if classification.Category != ClassificationCategoryLandscape || classification.Confidence != 0.85 {
		t.Fatalf("unexpected classification: %+v", classification)
	}
	if !classification.ClassifiedByAI {
		t.Fatal("expected ClassifiedByAI to be true when the AI service answered")
	}
}

func TestClassifyImageByAI_ProviderErrorIsReturned(t *testing.T) {
	file := files.FileDto{Name: "wallpaper.png", Path: "/downloads/wallpaper.png"}
	providerErr := errors.New("provider timeout")
	mock := &aiServiceMock{
		executeFn: func(ctx context.Context, req ai.Request) (ai.Response, error) {
			return ai.Response{}, providerErr
		},
	}
	_, err := ClassifyImageByAI(context.Background(), file, MetadataModel{}, mock)
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func TestClassifyImageByAI_InvalidJSONReturnsError(t *testing.T) {
	if _, err := classifyWithAnswer(t, "not json"); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestClassifyImageByAI_UnknownCategoryReturnsError(t *testing.T) {
	if _, err := classifyWithAnswer(t, `{"category": "unknown_cat", "confidence": 0.9}`); err == nil {
		t.Fatal("expected error for unknown category")
	}
}

func TestClassifyImageByAI_MarkdownCodeFenceStripped(t *testing.T) {
	classification, err := classifyWithAnswer(t, "```json\n{\"category\": \"meme\", \"confidence\": 0.80}\n```")
	if err != nil || classification.Category != ClassificationCategoryMeme {
		t.Fatalf("expected meme, got %+v err=%v", classification, err)
	}
}

func TestClassifyImageByAI_InvalidConfidenceDefaultsTo075(t *testing.T) {
	classification, err := classifyWithAnswer(t, `{"category": "art", "confidence": -1}`)
	if err != nil || classification.Category != ClassificationCategoryArt {
		t.Fatalf("expected art, got %+v err=%v", classification, err)
	}
	if classification.Confidence != 0.75 {
		t.Fatalf("expected 0.75 default confidence, got %f", classification.Confidence)
	}
}

func TestBuildClassificationPrompt(t *testing.T) {
	file := files.FileDto{Name: "photo.jpg", Path: "/photos/photo.jpg", Format: ".jpg"}
	metadata := MetadataModel{
		Width:            4000,
		Height:           3000,
		Make:             "Canon",
		Model:            "EOS R5",
		Software:         "Lightroom",
		ImageDescription: "A sunset",
	}
	prompt := buildClassificationPrompt(file, metadata)
	if !containsStr(prompt, "Filename: photo.jpg") {
		t.Fatalf("expected filename in prompt")
	}
	if !containsStr(prompt, "4000x3000") {
		t.Fatalf("expected dimensions in prompt")
	}
	if !containsStr(prompt, "Canon EOS R5") {
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

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
