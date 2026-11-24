package documenttext

import (
	"slices"
	"strings"

	"nas-go/api/pkg/utils"
)

const (
	GroupPDF          = "pdf"
	GroupText         = "text"
	GroupSpreadsheet  = "spreadsheet"
	GroupPresentation = "presentation"
	GroupOther        = "other"
)

var documentGroupByFormat = map[string]string{
	".pdf":  GroupPDF,
	".doc":  GroupText,
	".docx": GroupText,
	".odt":  GroupText,
	".rtf":  GroupText,
	".txt":  GroupText,
	".md":   GroupText,
	".xls":  GroupSpreadsheet,
	".xlsx": GroupSpreadsheet,
	".ods":  GroupSpreadsheet,
	".csv":  GroupSpreadsheet,
	".ppt":  GroupPresentation,
	".pptx": GroupPresentation,
	".odp":  GroupPresentation,
	".epub": GroupOther,
}

// DocumentGroupOf returns the library group of a file extension (with or
// without the leading dot, any case) and whether the extension is a document.
func DocumentGroupOf(format string) (group string, isDocument bool) {
	group, isDocument = documentGroupByFormat[withLeadingDot(utils.NormalizeExtension(format))]
	return group, isDocument
}

func withLeadingDot(extension string) string {
	if extension == "" || strings.HasPrefix(extension, ".") {
		return extension
	}
	return "." + extension
}

// DocumentFormats returns every document extension (lowercase, with dot) in a
// stable order, ready for a SQL "= ANY" filter.
func DocumentFormats() []string {
	formats := make([]string, 0, len(documentGroupByFormat))
	for format := range documentGroupByFormat {
		formats = append(formats, format)
	}
	slices.Sort(formats)
	return formats
}

// IndexableFormats returns the document formats plus every other format whose
// text the extractor can read, deduplicated.
func IndexableFormats() []string {
	formats := slices.Concat(DocumentFormats(), utils.DocumentTextFormats)
	slices.Sort(formats)
	return slices.Compact(formats)
}
