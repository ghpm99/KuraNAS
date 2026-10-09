package image

import (
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	maxContentCaptionLength = 300
	maxContentTagCount      = 15
	maxContentTagLength     = 40
	maxContentOCRTextLength = 2000
)

var (
	errContentCaptionMissing = errors.New("AI classification is missing caption")
	errContentTagsMissing    = errors.New("AI classification is missing tags")
	errContentOCRTextMissing = errors.New("AI classification is missing ocr_text")
	errContentTooManyTags    = errors.New("AI classification has too many tags")
	errContentTagTooLong     = errors.New("AI classification has an oversized tag")
)

type ContentDescription struct {
	Caption string
	Tags    []string
	OCRText string
}

func (description ContentDescription) SearchText() string {
	searchableParts := make([]string, 0, len(description.Tags)+2)
	searchableParts = append(searchableParts, description.Caption)
	searchableParts = append(searchableParts, description.Tags...)
	searchableParts = append(searchableParts, description.OCRText)
	return strings.ToLower(strings.Join(searchableParts, " "))
}

func buildContentDescription(rawCaption *string, rawTags *[]string, rawOCRText *string) (ContentDescription, error) {
	if rawCaption == nil || strings.TrimSpace(*rawCaption) == "" {
		return ContentDescription{}, errContentCaptionMissing
	}
	if rawTags == nil {
		return ContentDescription{}, errContentTagsMissing
	}
	if rawOCRText == nil {
		return ContentDescription{}, errContentOCRTextMissing
	}

	tags, err := normalizeContentTags(*rawTags)
	if err != nil {
		return ContentDescription{}, err
	}

	return ContentDescription{
		Caption: truncateRunes(strings.TrimSpace(*rawCaption), maxContentCaptionLength),
		Tags:    tags,
		OCRText: truncateRunes(strings.TrimSpace(*rawOCRText), maxContentOCRTextLength),
	}, nil
}

func normalizeContentTags(rawTags []string) ([]string, error) {
	if len(rawTags) > maxContentTagCount {
		return nil, fmt.Errorf("%w: %d", errContentTooManyTags, len(rawTags))
	}

	seenTags := make(map[string]bool, len(rawTags))
	tags := make([]string, 0, len(rawTags))
	for _, rawTag := range rawTags {
		tag := strings.ToLower(strings.TrimSpace(rawTag))
		if tag == "" || seenTags[tag] {
			continue
		}
		if utf8.RuneCountInString(tag) > maxContentTagLength {
			return nil, errContentTagTooLong
		}
		seenTags[tag] = true
		tags = append(tags, tag)
	}
	return tags, nil
}

func truncateRunes(text string, maxRunes int) string {
	if utf8.RuneCountInString(text) <= maxRunes {
		return text
	}
	return string([]rune(text)[:maxRunes])
}
