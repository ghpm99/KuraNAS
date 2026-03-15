package files

import "strings"

type ImageClassificationCategory string

const (
	ImageClassificationCategoryCapture ImageClassificationCategory = "capture"
	ImageClassificationCategoryPhoto   ImageClassificationCategory = "photo"
	ImageClassificationCategoryOther   ImageClassificationCategory = "other"
)

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

func ClassifyImage(file FileDto, metadata ImageMetadataModel) ImageClassificationModel {
	if looksLikeCapture(file, metadata) {
		return ImageClassificationModel{
			Category:   ImageClassificationCategoryCapture,
			Confidence: 0.98,
		}
	}

	if confidence := photoConfidence(file, metadata); confidence > 0 {
		return ImageClassificationModel{
			Category:   ImageClassificationCategoryPhoto,
			Confidence: confidence,
		}
	}

	return ImageClassificationModel{
		Category:   ImageClassificationCategoryOther,
		Confidence: 0.35,
	}
}

func looksLikeCapture(file FileDto, metadata ImageMetadataModel) bool {
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

func photoConfidence(file FileDto, metadata ImageMetadataModel) float64 {
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
