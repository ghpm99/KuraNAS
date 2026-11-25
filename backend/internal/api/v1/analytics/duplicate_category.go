package analytics

import (
	"slices"

	"nas-go/api/pkg/utils"
)

type DuplicateCategory string

const (
	DuplicateCategoryAll      DuplicateCategory = ""
	DuplicateCategoryImage    DuplicateCategory = "image"
	DuplicateCategoryAudio    DuplicateCategory = "audio"
	DuplicateCategoryVideo    DuplicateCategory = "video"
	DuplicateCategoryDocument DuplicateCategory = "document"
	DuplicateCategoryOther    DuplicateCategory = "other"
)

var officeDocumentFormats = []string{".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx", ".md"}

// ParseDuplicateCategory validates a "type" query value for the paginated
// duplicate groups listing. An empty value means every category.
func ParseDuplicateCategory(rawCategory string) (DuplicateCategory, bool) {
	category := DuplicateCategory(rawCategory)
	switch category {
	case DuplicateCategoryAll, DuplicateCategoryImage, DuplicateCategoryAudio,
		DuplicateCategoryVideo, DuplicateCategoryDocument, DuplicateCategoryOther:
		return category, true
	}
	return DuplicateCategoryAll, false
}

func documentFormats() []string {
	return slices.Concat(utils.DocumentFormats, officeDocumentFormats)
}

func (category DuplicateCategory) includedFormats() []string {
	switch category {
	case DuplicateCategoryImage:
		return utils.ImageFormats
	case DuplicateCategoryAudio:
		return utils.AudioFormats
	case DuplicateCategoryVideo:
		return utils.VideoFormats
	case DuplicateCategoryDocument:
		return documentFormats()
	}
	return nil
}

func (category DuplicateCategory) excludedFormats() []string {
	if category != DuplicateCategoryOther {
		return []string{}
	}
	return slices.Concat(utils.ImageFormats, utils.AudioFormats, utils.VideoFormats, documentFormats())
}
