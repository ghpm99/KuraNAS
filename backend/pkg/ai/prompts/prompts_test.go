package prompts

import (
	"strings"
	"testing"
)

func TestSystemPromptsAreEmbedded(t *testing.T) {
	tests := []struct {
		name      string
		prompt    string
		mustMatch string
	}{
		{
			name:      "search extraction",
			prompt:    SearchExtractionSystemPrompt(),
			mustMatch: "Respond ONLY with JSON",
		},
		{
			name:      "analytics insights",
			prompt:    AnalyticsInsightsSystemPrompt(),
			mustMatch: "storage analytics assistant",
		},
		{
			name:      "video descriptions",
			prompt:    VideoCatalogDescriptionsSystemPrompt(),
			mustMatch: "short contextual descriptions",
		},
		{
			name:      "image classification",
			prompt:    ImageClassificationSystemPrompt(),
			mustMatch: "image analyst",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if strings.TrimSpace(tc.prompt) == "" {
				t.Fatalf("expected non-empty prompt")
			}
			if !strings.Contains(tc.prompt, tc.mustMatch) {
				t.Fatalf("expected prompt to contain %q", tc.mustMatch)
			}
		})
	}
}

func TestUserPromptFormatting(t *testing.T) {
	searchPrompt := SearchExtractionUserPrompt("my files")
	if !strings.Contains(searchPrompt, "Query: 'my files'") {
		t.Fatalf("unexpected search prompt formatting")
	}

	summary := "Storage: 80% used"
	analyticsPrompt := AnalyticsInsightsUserPrompt(summary)
	if !strings.Contains(analyticsPrompt, summary) {
		t.Fatalf("analytics prompt missing summary payload")
	}

	sections := "Section 'Series' (2 items): S01E01, S01E02"
	videoPrompt := VideoCatalogDescriptionsUserPrompt(sections)
	if !strings.Contains(videoPrompt, sections) {
		t.Fatalf("video prompt missing sections payload")
	}

	metadata := "Filename: photo.jpg\nDimensions: 4000x3000"
	imagePrompt := ImageClassificationUserPrompt(metadata)
	if !strings.Contains(imagePrompt, metadata) {
		t.Fatalf("image prompt missing metadata payload")
	}

	formattedPrompts := []string{searchPrompt, analyticsPrompt, videoPrompt, imagePrompt}
	for _, prompt := range formattedPrompts {
		if strings.Contains(prompt, "%!") {
			t.Fatalf("prompt formatting error: %s", prompt)
		}
	}
}

func TestEmailPromptsEmbedAndInterpolate(t *testing.T) {
	if !strings.Contains(EmailClassificationSystemPrompt(), "UNTRUSTED DATA") {
		t.Fatal("classification system prompt must flag untrusted data")
	}
	if !strings.Contains(EmailSummarySystemPrompt(), "UNTRUSTED DATA") {
		t.Fatal("summary system prompt must flag untrusted data")
	}

	user := EmailClassificationUserPrompt("NONCE123", "- SPF: pass", "Hello", "Body text")
	for _, want := range []string{"<<EMAIL-NONCE123>>", "<</EMAIL-NONCE123>>", "- SPF: pass", "Hello", "Body text"} {
		if !strings.Contains(user, want) {
			t.Fatalf("classification user prompt missing %q:\n%s", want, user)
		}
	}

	summary := EmailSummaryUserPrompt("ABC", "Subj", "Some body")
	for _, want := range []string{"<<EMAIL-ABC>>", "<</EMAIL-ABC>>", "Subj", "Some body"} {
		if !strings.Contains(summary, want) {
			t.Fatalf("summary user prompt missing %q:\n%s", want, summary)
		}
	}
}
