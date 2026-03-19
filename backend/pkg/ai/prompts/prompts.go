package prompts

import (
	_ "embed"
	"fmt"
	"strings"
)

//go:embed search_extraction_system.txt
var searchExtractionSystemPrompt string

//go:embed search_extraction_user.txt
var searchExtractionUserPromptTemplate string

//go:embed analytics_insights_system.txt
var analyticsInsightsSystemPrompt string

//go:embed analytics_insights_user.txt
var analyticsInsightsUserPromptTemplate string

//go:embed video_catalog_descriptions_system.txt
var videoCatalogDescriptionsSystemPrompt string

//go:embed video_catalog_descriptions_user.txt
var videoCatalogDescriptionsUserPromptTemplate string

//go:embed image_classification_system.txt
var imageClassificationSystemPrompt string

//go:embed image_classification_user.txt
var imageClassificationUserPromptTemplate string

func SearchExtractionSystemPrompt() string {
	return strings.TrimSpace(searchExtractionSystemPrompt)
}

func SearchExtractionUserPrompt(query string) string {
	return fmt.Sprintf(strings.TrimSpace(searchExtractionUserPromptTemplate), query)
}

func AnalyticsInsightsSystemPrompt() string {
	return strings.TrimSpace(analyticsInsightsSystemPrompt)
}

func AnalyticsInsightsUserPrompt(summary string) string {
	return fmt.Sprintf(strings.TrimSpace(analyticsInsightsUserPromptTemplate), summary)
}

func VideoCatalogDescriptionsSystemPrompt() string {
	return strings.TrimSpace(videoCatalogDescriptionsSystemPrompt)
}

func VideoCatalogDescriptionsUserPrompt(sections string) string {
	return fmt.Sprintf(strings.TrimSpace(videoCatalogDescriptionsUserPromptTemplate), sections)
}

func ImageClassificationSystemPrompt() string {
	return strings.TrimSpace(imageClassificationSystemPrompt)
}

func ImageClassificationUserPrompt(metadata string) string {
	return fmt.Sprintf(strings.TrimSpace(imageClassificationUserPromptTemplate), metadata)
}
