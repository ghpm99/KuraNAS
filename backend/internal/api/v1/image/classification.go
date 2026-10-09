package image

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"nas-go/api/internal/api/v1/files"
	"nas-go/api/internal/config"
	"nas-go/api/pkg/ai"
	"nas-go/api/pkg/ai/prompts"
	"nas-go/api/pkg/img"
	"regexp"
	"strings"
)

const defaultContentLanguage = "en-US"

// visionMaxDimension caps the longest edge of the image sent to the AI. A
// downscaled copy is enough for recognition and keeps the base64 payload (and
// inference time) small.
const visionMaxDimension = 768

// suggestedNameSanitizer keeps suggested filenames filesystem-safe.
var suggestedNameSanitizer = regexp.MustCompile(`[^a-zA-Z0-9-_]+`)

// multiUnderscore collapses runs of underscores into a single one.
var multiUnderscore = regexp.MustCompile(`_+`)

// ClassificationCategory is the type for an image classification category.
type ClassificationCategory string

const (
	ClassificationCategoryCapture    ClassificationCategory = "capture"
	ClassificationCategoryPhoto      ClassificationCategory = "photo"
	ClassificationCategoryOther      ClassificationCategory = "other"
	ClassificationCategoryDocument   ClassificationCategory = "document"
	ClassificationCategoryReceipt    ClassificationCategory = "receipt"
	ClassificationCategoryLandscape  ClassificationCategory = "landscape"
	ClassificationCategoryPortrait   ClassificationCategory = "portrait"
	ClassificationCategoryMeme       ClassificationCategory = "meme"
	ClassificationCategoryArt        ClassificationCategory = "art"
	ClassificationCategoryScreenshot ClassificationCategory = "screenshot_app"
)

// AIClassificationConfidenceThreshold is the heuristic confidence below which
// the AI classifier takes over. The backfill targets exactly the images that
// would have gone to the AI under normal indexing (confidence below this).
const AIClassificationConfidenceThreshold = 0.70

// AIClassificationJobScopePath is the fixed scope path shared by every job that
// drains the AI classification backlog, so at most one such job is pending.
const AIClassificationJobScopePath = "image_ai_classify"

// ErrAIServiceUnavailable means no AI service is wired in for classification.
var ErrAIServiceUnavailable = errors.New("AI service is unavailable for image classification")

var screenshotKeywords = []string{
	"screenshot",
	"screen shot",
	"captura",
	"snipping tool",
	"snip & sketch",
}

var photoPathHints = []string{
	"/dcim/",
	"/camera/",
	"/photos/",
	"/pictures/",
}

// ClassifyImage returns a heuristic classification for an image based on
// file metadata (EXIF fields, filename, path). It does not call the AI.
func ClassifyImage(file files.FileDto, metadata MetadataModel) ClassificationModel {
	if looksLikeCapture(file, metadata) {
		return ClassificationModel{
			Category:   ClassificationCategoryCapture,
			Confidence: 0.98,
		}
	}

	if confidence := photoConfidence(file, metadata); confidence > 0 {
		return ClassificationModel{
			Category:   ClassificationCategoryPhoto,
			Confidence: confidence,
		}
	}

	return ClassificationModel{
		Category:   ClassificationCategoryOther,
		Confidence: 0.35,
	}
}

func looksLikeCapture(file files.FileDto, metadata MetadataModel) bool {
	sample := strings.ToLower(strings.Join([]string{
		file.Name,
		file.Path,
		metadata.Software,
		metadata.ImageDescription,
	}, " "))

	for _, keyword := range screenshotKeywords {
		if strings.Contains(sample, keyword) {
			return true
		}
	}

	return false
}

var validAICategories = map[ClassificationCategory]bool{
	ClassificationCategoryCapture:    true,
	ClassificationCategoryPhoto:      true,
	ClassificationCategoryOther:      true,
	ClassificationCategoryDocument:   true,
	ClassificationCategoryReceipt:    true,
	ClassificationCategoryLandscape:  true,
	ClassificationCategoryPortrait:   true,
	ClassificationCategoryMeme:       true,
	ClassificationCategoryArt:        true,
	ClassificationCategoryScreenshot: true,
}

// ClassifyImageByAI asks the vision model to classify an image, honoring the
// deadline carried by ctx. It returns an error when the service is missing, the
// call fails or the answer cannot be parsed, so the caller decides how to record
// the failure.
func ClassifyImageByAI(ctx context.Context, file files.FileDto, metadata MetadataModel, aiService ai.ServiceInterface) (ClassificationModel, error) {
	if aiService == nil {
		return ClassificationModel{}, ErrAIServiceUnavailable
	}

	response, err := aiService.Execute(ctx, ai.Request{
		TaskType:     ai.TaskClassification,
		SystemPrompt: prompts.ImageClassificationSystemPrompt(),
		Prompt:       buildClassificationPrompt(file, metadata),
		MaxTokens:    1200,
		Temperature:  0.1,
		Images:       encodeImageForAI(file.ResolveContentPath()),
	})
	if err != nil {
		return ClassificationModel{}, fmt.Errorf("AI image classification failed: %w", err)
	}

	classification, err := parseAIClassificationResponse(response.Content)
	if err != nil {
		return ClassificationModel{}, err
	}

	classification.ClassifiedByAI = true
	return classification, nil
}

// encodeImageForAI loads an image, downscales it and returns it as a one-element
// slice of base64 PNG data ready for a multimodal request. Returns nil on any
// failure so the caller degrades to a text-only request.
func encodeImageForAI(path string) []string {
	if path == "" {
		return nil
	}
	src, _, err := img.OpenImageFromFile(path)
	if err != nil {
		log.Printf("AI vision: failed to open image %q: %v\n", path, err)
		return nil
	}
	resized := img.Thumbnail(src, visionMaxDimension, visionMaxDimension)
	encoded, err := img.EncodePNG(resized)
	if err != nil {
		log.Printf("AI vision: failed to encode image %q: %v\n", path, err)
		return nil
	}
	return []string{base64.StdEncoding.EncodeToString(encoded)}
}

// sanitizeSuggestedName makes an AI-proposed filename filesystem-safe: keeps
// alphanumerics, dashes and underscores, collapses the rest, and bounds length.
func sanitizeSuggestedName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	name = strings.ReplaceAll(name, " ", "_")
	name = suggestedNameSanitizer.ReplaceAllString(name, "_")
	name = multiUnderscore.ReplaceAllString(name, "_")
	name = strings.Trim(name, "_-")
	if len(name) > 80 {
		name = name[:80]
		name = strings.Trim(name, "_-")
	}
	return name
}

func buildClassificationPrompt(file files.FileDto, metadata MetadataModel) string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Filename: %s", file.Name))
	parts = append(parts, fmt.Sprintf("Path: %s", file.Path))
	parts = append(parts, fmt.Sprintf("Format: %s", file.Format))

	if metadata.Width > 0 && metadata.Height > 0 {
		parts = append(parts, fmt.Sprintf("Dimensions: %dx%d", metadata.Width, metadata.Height))
	}
	if metadata.Make != "" {
		parts = append(parts, fmt.Sprintf("Camera: %s %s", metadata.Make, metadata.Model))
	}
	if metadata.Software != "" {
		parts = append(parts, fmt.Sprintf("Software: %s", metadata.Software))
	}
	if metadata.ImageDescription != "" {
		parts = append(parts, fmt.Sprintf("Description: %s", metadata.ImageDescription))
	}

	return prompts.ImageClassificationUserPrompt(strings.Join(parts, "\n"), contentLanguage())
}

func contentLanguage() string {
	if configuredLanguage := strings.TrimSpace(config.AppConfig.Lang); configuredLanguage != "" {
		return configuredLanguage
	}
	return defaultContentLanguage
}

type aiClassificationResponse struct {
	Category      string    `json:"category"`
	Confidence    float64   `json:"confidence"`
	SuggestedName string    `json:"suggested_name"`
	Caption       *string   `json:"caption"`
	Tags          *[]string `json:"tags"`
	OCRText       *string   `json:"ocr_text"`
}

func parseAIClassificationResponse(content string) (ClassificationModel, error) {
	content = strings.TrimSpace(content)

	// Strip markdown code fences if present
	if strings.HasPrefix(content, "```") {
		lines := strings.Split(content, "\n")
		filtered := make([]string, 0, len(lines))
		for _, line := range lines {
			if !strings.HasPrefix(strings.TrimSpace(line), "```") {
				filtered = append(filtered, line)
			}
		}
		content = strings.Join(filtered, "\n")
	}

	var resp aiClassificationResponse
	if err := json.Unmarshal([]byte(content), &resp); err != nil {
		return ClassificationModel{}, fmt.Errorf("invalid AI classification JSON: %w", err)
	}

	category := ClassificationCategory(strings.ToLower(resp.Category))
	if !validAICategories[category] {
		return ClassificationModel{}, fmt.Errorf("unknown AI category: %s", resp.Category)
	}

	contentDescription, err := buildContentDescription(resp.Caption, resp.Tags, resp.OCRText)
	if err != nil {
		return ClassificationModel{}, err
	}

	confidence := resp.Confidence
	if confidence <= 0 || confidence > 1 {
		confidence = 0.75
	}

	return ClassificationModel{
		Category:      category,
		Confidence:    confidence,
		SuggestedName: sanitizeSuggestedName(resp.SuggestedName),
		Content:       contentDescription,
	}, nil
}

func photoConfidence(file files.FileDto, metadata MetadataModel) float64 {
	evidence := 0

	if metadata.Make != "" || metadata.Model != "" {
		evidence += 2
	}
	if metadata.LensModel != "" || metadata.SerialNumber != "" {
		evidence++
	}
	if metadata.DateTimeOriginal != "" || metadata.DateTimeDigitized != "" {
		evidence++
	}
	if metadata.ExposureTime > 0 || metadata.FNumber > 0 || metadata.ISO > 0 || metadata.FocalLength > 0 {
		evidence++
	}
	if metadata.GPSLatitude != 0 || metadata.GPSLongitude != 0 {
		evidence++
	}

	pathSample := strings.ToLower(file.Path)
	for _, hint := range photoPathHints {
		if strings.Contains(pathSample, hint) {
			evidence++
			break
		}
	}

	switch {
	case evidence >= 5:
		return 0.97
	case evidence >= 4:
		return 0.9
	case evidence >= 3:
		return 0.82
	case evidence >= 2:
		return 0.72
	default:
		return 0
	}
}

var allClassificationCategories = []ClassificationCategory{
	ClassificationCategoryCapture,
	ClassificationCategoryPhoto,
	ClassificationCategoryOther,
	ClassificationCategoryDocument,
	ClassificationCategoryReceipt,
	ClassificationCategoryLandscape,
	ClassificationCategoryPortrait,
	ClassificationCategoryMeme,
	ClassificationCategoryArt,
	ClassificationCategoryScreenshot,
}

func ParseClassificationCategory(rawCategory string) (ClassificationCategory, bool) {
	category := ClassificationCategory(strings.ToLower(strings.TrimSpace(rawCategory)))
	for _, knownCategory := range allClassificationCategories {
		if category == knownCategory {
			return category, true
		}
	}
	return "", false
}
