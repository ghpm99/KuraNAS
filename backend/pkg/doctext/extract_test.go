package doctext

import (
	"archive/zip"
	"bytes"
	"compress/zlib"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf16"
)

func writeFixture(t *testing.T, name string, content []byte) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func extractFixture(t *testing.T, name string, content []byte) (Result, error) {
	t.Helper()
	path := writeFixture(t, name, content)
	return Extract(context.Background(), path, filepath.Ext(name))
}

func encodeUTF16WithBOM(text string, byteOrder binary.AppendByteOrder, bom []byte) []byte {
	encoded := append([]byte{}, bom...)
	for _, codeUnit := range utf16.Encode([]rune(text)) {
		encoded = byteOrder.AppendUint16(encoded, codeUnit)
	}
	return encoded
}

func TestExtractPlainTextEncodings(t *testing.T) {
	testCases := []struct {
		name     string
		fileName string
		content  []byte
		expected string
	}{
		{"utf8", "note.txt", []byte("Relatório de ação"), "Relatório de ação"},
		{"utf8 with bom", "note.md", append([]byte{0xEF, 0xBB, 0xBF}, []byte("# Título")...), "# Título"},
		{"utf16 little endian", "note.txt", encodeUTF16WithBOM("Olá mundo", binary.LittleEndian, []byte{0xFF, 0xFE}), "Olá mundo"},
		{"utf16 big endian", "note.log", encodeUTF16WithBOM("Olá mundo", binary.BigEndian, []byte{0xFE, 0xFF}), "Olá mundo"},
		{"latin1", "note.csv", []byte{'a', 0xE7, 0xE3, 'o'}, "ação"},
		{"crlf and control characters", "note.txt", []byte("one\r\ntwo\x07"), "one\ntwo"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := extractFixture(t, testCase.fileName, testCase.content)
			if err != nil {
				t.Fatalf("Extract returned error: %v", err)
			}
			if result.Text != testCase.expected || result.Truncated {
				t.Fatalf("got %q truncated=%v, want %q", result.Text, result.Truncated, testCase.expected)
			}
		})
	}
}

func TestExtractSkipsBinaryContent(t *testing.T) {
	_, err := extractFixture(t, "payload.txt", []byte("abc\x00def"))
	if !errors.Is(err, ErrBinary) || ErrorCode(err) != "binary" {
		t.Fatalf("expected ErrBinary, got %v", err)
	}
}

func TestExtractRejectsUnsupportedFormat(t *testing.T) {
	_, err := extractFixture(t, "photo.jpg", []byte("text"))
	if !errors.Is(err, ErrUnsupportedFormat) {
		t.Fatalf("expected ErrUnsupportedFormat, got %v", err)
	}
}

func TestExtractMissingFileIsUnreadable(t *testing.T) {
	_, err := Extract(context.Background(), filepath.Join(t.TempDir(), "gone.txt"), ".txt")
	if ErrorCode(err) != "unreadable" {
		t.Fatalf("expected unreadable code, got %q (%v)", ErrorCode(err), err)
	}
}

func TestExtractRejectsFilesAboveSizeLimit(t *testing.T) {
	path := filepath.Join(t.TempDir(), "huge.txt")
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := file.Truncate(MaxFileSizeBytes + 1); err != nil {
		t.Fatal(err)
	}
	file.Close()

	_, err = Extract(context.Background(), path, ".txt")
	if !errors.Is(err, ErrTooLarge) || ErrorCode(err) != "too_large" {
		t.Fatalf("expected ErrTooLarge, got %v", err)
	}
}

func TestExtractTruncatesLongTextAtRuneBoundary(t *testing.T) {
	longText := strings.Repeat("ção ", MaxStoredTextBytes)
	result, err := extractFixture(t, "long.txt", []byte(longText))
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if !result.Truncated || len(result.Text) > MaxStoredTextBytes {
		t.Fatalf("expected truncated text within limit, got len=%d truncated=%v", len(result.Text), result.Truncated)
	}
	if strings.ContainsRune(result.Text, '�') {
		t.Fatal("truncation split a rune")
	}
}

func TestExtractReadsOnlyLeadingBytesOfLargeTextFile(t *testing.T) {
	content := append([]byte(strings.Repeat("a", MaxReadBytes)), []byte("tail-marker")...)
	result, err := extractFixture(t, "big.log", content)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if !result.Truncated || strings.Contains(result.Text, "tail-marker") {
		t.Fatalf("expected leading window only, truncated=%v", result.Truncated)
	}
}

func TestExtractMarkupStripsTagsScriptsAndEntities(t *testing.T) {
	page := `<html><head><style>p{color:red}</style><script>var secret = 1;</script></head>
<body><!-- hidden --><h1>Título</h1><p>Caf&eacute; &amp; <b>pão</b></p><SCRIPT type="x">alert(2)</SCRIPT></body></html>`
	result, err := extractFixture(t, "page.html", []byte(page))
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if result.Text != "Título Café & pão" {
		t.Fatalf("unexpected markup text %q", result.Text)
	}
}

func TestExtractXMLKeepsElementText(t *testing.T) {
	result, err := extractFixture(t, "data.xml", []byte(`<?xml version="1.0"?><root><a>um</a><b attr="x>y">dois</b></root>`))
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if !strings.Contains(result.Text, "um") || !strings.Contains(result.Text, "dois") {
		t.Fatalf("unexpected xml text %q", result.Text)
	}
}

func buildDocx(t *testing.T, documentXML string, entryName string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	entry, err := archive.Create(entryName)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write([]byte(documentXML)); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

const wordDocumentXML = `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>
<w:p><w:r><w:t>Contrato de </w:t></w:r><w:r><w:t>locação</w:t></w:r></w:p>
<w:p><w:r><w:t>Segunda</w:t></w:r><w:r><w:tab/></w:r><w:r><w:t>linha</w:t></w:r></w:p>
<w:p><w:r><w:instrText>IGNORED</w:instrText></w:r></w:p>
</w:body></w:document>`

func TestExtractDocxCollectsWordprocessingText(t *testing.T) {
	result, err := extractFixture(t, "contract.docx", buildDocx(t, wordDocumentXML, "word/document.xml"))
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if result.Text != "Contrato de locação\nSegunda linha" {
		t.Fatalf("unexpected docx text %q", result.Text)
	}
}

func TestExtractDocxWithoutMainDocumentFails(t *testing.T) {
	_, err := extractFixture(t, "broken.docx", buildDocx(t, wordDocumentXML, "word/other.xml"))
	if !errors.Is(err, errDocxMainDocumentMissing) || ErrorCode(err) != "extract_failed" {
		t.Fatalf("expected missing main document, got %v", err)
	}
}

func TestExtractDocxNotAZipFails(t *testing.T) {
	_, err := extractFixture(t, "fake.docx", []byte("not a zip"))
	if err == nil || ErrorCode(err) != "extract_failed" {
		t.Fatalf("expected extract_failed, got %v", err)
	}
}

func TestExtractDocxStopsAtStoredTextLimit(t *testing.T) {
	var documentXML strings.Builder
	documentXML.WriteString(`<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:body>`)
	for paragraph := 0; paragraph < 5000; paragraph++ {
		documentXML.WriteString("<w:p><w:r><w:t>" + strings.Repeat("palavra ", 10) + "</w:t></w:r></w:p>")
	}
	documentXML.WriteString(`</w:body></w:document>`)

	result, err := extractFixture(t, "big.docx", buildDocx(t, documentXML.String(), "word/document.xml"))
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if !result.Truncated || len(result.Text) > MaxStoredTextBytes {
		t.Fatalf("expected truncated docx text, len=%d truncated=%v", len(result.Text), result.Truncated)
	}
}

func deflate(t *testing.T, content []byte) []byte {
	t.Helper()
	var buffer bytes.Buffer
	writer := zlib.NewWriter(&buffer)
	if _, err := writer.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

func buildPDF(t *testing.T, contentStream string, isFlate bool, extraTrailer string) []byte {
	t.Helper()
	streamData := []byte(contentStream)
	filter := ""
	if isFlate {
		streamData = deflate(t, streamData)
		filter = " /Filter /FlateDecode"
	}

	var document bytes.Buffer
	document.WriteString("%PDF-1.4\n")
	document.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	document.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	document.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /Contents 4 0 R >>\nendobj\n")
	document.WriteString(fmt.Sprintf("4 0 obj\n<< /Length %d%s >>\nstream\n", len(streamData), filter))
	document.Write(streamData)
	document.WriteString("\nendstream\nendobj\n")
	document.WriteString("5 0 obj\n<< /Length 4 /Subtype /Image /Filter /DCTDecode >>\nstream\nBT()\nendstream\nendobj\n")
	document.WriteString("trailer\n<< /Root 1 0 R" + extraTrailer + " >>\n%%EOF\n")
	return document.Bytes()
}

const pdfContentStream = `BT /F1 12 Tf 72 700 Td (Hello ) Tj [(Wor) -300 (ld)] TJ 0 -14 Td (Acentua\347\343o \(par\)) Tj T* (linha dois) Tj ET`

func TestExtractPDFReadsFlateTextStream(t *testing.T) {
	result, err := extractFixture(t, "doc.pdf", buildPDF(t, pdfContentStream, true, ""))
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	for _, expected := range []string{"Hello", "Wor ld", "Acentuação (par)", "linha dois"} {
		if !strings.Contains(result.Text, expected) {
			t.Fatalf("expected %q in %q", expected, result.Text)
		}
	}
	if strings.Contains(result.Text, "BT") {
		t.Fatalf("image stream leaked into text: %q", result.Text)
	}
}

func TestExtractPDFReadsUncompressedTextStream(t *testing.T) {
	result, err := extractFixture(t, "plain.pdf", buildPDF(t, pdfContentStream, false, ""))
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if !strings.Contains(result.Text, "Hello") {
		t.Fatalf("unexpected pdf text %q", result.Text)
	}
}

func TestExtractPDFEncryptedFails(t *testing.T) {
	_, err := extractFixture(t, "locked.pdf", buildPDF(t, pdfContentStream, true, " /Encrypt 9 0 R"))
	if !errors.Is(err, ErrEncrypted) || ErrorCode(err) != "encrypted" {
		t.Fatalf("expected ErrEncrypted, got %v", err)
	}
}

func TestExtractPDFNotAPDFFails(t *testing.T) {
	_, err := extractFixture(t, "fake.pdf", []byte("plain text pretending"))
	if !errors.Is(err, errNotPDF) {
		t.Fatalf("expected errNotPDF, got %v", err)
	}
}

func TestExtractPDFCorruptStreamYieldsEmptyText(t *testing.T) {
	corrupt := []byte("%PDF-1.4\n4 0 obj\n<< /Filter /FlateDecode >>\nstream\n\x01\x02garbage\nendstream\n")
	result, err := extractFixture(t, "corrupt.pdf", corrupt)
	if err != nil {
		t.Fatalf("Extract returned error: %v", err)
	}
	if result.Text != "" {
		t.Fatalf("expected empty text, got %q", result.Text)
	}
}

func TestExtractHonorsCancelledContext(t *testing.T) {
	path := writeFixture(t, "note.txt", []byte("text"))
	cancelledContext, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Extract(cancelledContext, path, ".txt")
	if ErrorCode(err) != "timeout" {
		t.Fatalf("expected timeout code, got %q (%v)", ErrorCode(err), err)
	}
}

func TestExtractPDFStopsWhenContextExpires(t *testing.T) {
	path := writeFixture(t, "doc.pdf", buildPDF(t, pdfContentStream, true, ""))
	expiredContext, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := extractPDFText(expiredContext, path)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}
