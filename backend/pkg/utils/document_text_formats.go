package utils

import "slices"

var MarkupTextFormats = []string{".html", ".htm", ".xhtml", ".xml"}

var PlainTextFormats = []string{
	".txt", ".md", ".markdown", ".rst", ".log", ".csv", ".tsv", ".json", ".yml", ".yaml", ".toml", ".ini", ".cfg", ".conf",
	".properties", ".srt", ".vtt", ".tex", ".sql", ".sh", ".bat", ".ps1", ".css", ".scss", ".go", ".py", ".js", ".jsx", ".tsx",
	".java", ".kt", ".c", ".h", ".cpp", ".hpp", ".cs", ".rb", ".php", ".rs", ".swift", ".lua", ".vue", ".gradle",
}

var RichDocumentFormats = []string{".pdf", ".docx"}

var DocumentTextFormats = slices.Concat(PlainTextFormats, MarkupTextFormats, RichDocumentFormats)

func IsDocumentTextFormat(extension string) bool {
	return slices.Contains(DocumentTextFormats, NormalizeExtension(extension))
}
