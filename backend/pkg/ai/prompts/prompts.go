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

//go:embed image_classification_system.txt
var imageClassificationSystemPrompt string

//go:embed image_classification_user.txt
var imageClassificationUserPromptTemplate string

//go:embed music_artist_clusters_system.txt
var musicArtistClustersSystemPrompt string

//go:embed music_artist_clusters_user.txt
var musicArtistClustersUserPromptTemplate string

//go:embed assistant_chat_system.txt
var assistantChatSystemPrompt string

//go:embed email_classification_system.txt
var emailClassificationSystemPrompt string

//go:embed email_classification_user.txt
var emailClassificationUserPromptTemplate string

//go:embed email_summary_system.txt
var emailSummarySystemPrompt string

//go:embed email_summary_user.txt
var emailSummaryUserPromptTemplate string

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

func ImageClassificationSystemPrompt() string {
	return strings.TrimSpace(imageClassificationSystemPrompt)
}

func ImageClassificationUserPrompt(metadata string, language string) string {
	return fmt.Sprintf(strings.TrimSpace(imageClassificationUserPromptTemplate), metadata, language)
}

func MusicArtistClustersSystemPrompt() string {
	return strings.TrimSpace(musicArtistClustersSystemPrompt)
}

func MusicArtistClustersUserPrompt(maxNewClusters int, existingClusters string, artists string) string {
	return fmt.Sprintf(strings.TrimSpace(musicArtistClustersUserPromptTemplate), maxNewClusters, existingClusters, artists)
}

func AssistantChatSystemPrompt() string {
	return strings.TrimSpace(assistantChatSystemPrompt)
}

func EmailClassificationSystemPrompt() string {
	return strings.TrimSpace(emailClassificationSystemPrompt)
}

// EmailClassificationUserPrompt embeds the trusted deterministic evidence and
// the untrusted subject/body between per-request random delimiters (nonce), so
// the model can tell quoted e-mail data from its instructions.
func EmailClassificationUserPrompt(nonce, evidence, subject, body string) string {
	return fmt.Sprintf(strings.TrimSpace(emailClassificationUserPromptTemplate), nonce, evidence, subject, body)
}

func EmailSummarySystemPrompt() string {
	return strings.TrimSpace(emailSummarySystemPrompt)
}

// EmailSummaryUserPrompt wraps the untrusted subject/body between per-request
// random delimiters (nonce), same data-not-instructions framing as the
// classification prompt.
func EmailSummaryUserPrompt(nonce, subject, body string) string {
	return fmt.Sprintf(strings.TrimSpace(emailSummaryUserPromptTemplate), nonce, subject, body)
}
