package documenttext

import (
	"slices"
	"testing"
)

func TestDocumentGroupOfMapsEveryDocumentFormat(t *testing.T) {
	expectedGroups := map[string]string{
		".pdf": GroupPDF,
		".doc": GroupText, ".docx": GroupText, ".odt": GroupText, ".rtf": GroupText, ".txt": GroupText, ".md": GroupText,
		".xls": GroupSpreadsheet, ".xlsx": GroupSpreadsheet, ".ods": GroupSpreadsheet, ".csv": GroupSpreadsheet,
		".ppt": GroupPresentation, ".pptx": GroupPresentation, ".odp": GroupPresentation,
		".epub": GroupOther,
	}
	for format, expectedGroup := range expectedGroups {
		group, isDocument := DocumentGroupOf(format)
		if !isDocument || group != expectedGroup {
			t.Fatalf("%s: got %q/%v, want %q", format, group, isDocument, expectedGroup)
		}
	}
	if len(DocumentFormats()) != len(expectedGroups) {
		t.Fatalf("document format count %d, want %d", len(DocumentFormats()), len(expectedGroups))
	}
}

func TestDocumentGroupOfNormalizesExtension(t *testing.T) {
	for _, format := range []string{"PDF", ".Pdf", "pdf"} {
		if group, isDocument := DocumentGroupOf(format); !isDocument || group != GroupPDF {
			t.Fatalf("%s: got %q/%v", format, group, isDocument)
		}
	}
}

func TestDocumentGroupOfExcludesCodeLogsAndConfig(t *testing.T) {
	for _, format := range []string{".go", ".log", ".json", ".yml", ".ini", ".py", ".sql", ".html", ".jpg", ".mp4", ""} {
		if group, isDocument := DocumentGroupOf(format); isDocument {
			t.Fatalf("%q must not be a document, got group %q", format, group)
		}
	}
}

func TestIndexableFormatsCoverDocumentsAndExtractableText(t *testing.T) {
	indexableFormats := IndexableFormats()
	for _, format := range []string{".doc", ".epub", ".pdf", ".go", ".log", ".odt"} {
		if !slices.Contains(indexableFormats, format) {
			t.Fatalf("%s missing from indexable formats", format)
		}
	}
	if len(slices.Compact(slices.Clone(indexableFormats))) != len(indexableFormats) {
		t.Fatal("indexable formats must be unique")
	}
}
