package doctext

import (
	"bytes"
	"fmt"
	"slices"
	"strings"
	"testing"
)

func pageContent(text string) string {
	return fmt.Sprintf("BT /F1 12 Tf 72 700 Td (%s) Tj ET", text)
}

func buildPagedPDF(t *testing.T, pageTexts []string, contentsAsArray bool) []byte {
	t.Helper()
	var document bytes.Buffer
	document.WriteString("%PDF-1.4\n")
	document.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	kids := make([]string, 0, len(pageTexts))
	for pageIndex := range pageTexts {
		kids = append(kids, fmt.Sprintf("%d 0 R", 10+pageIndex))
	}
	document.WriteString(fmt.Sprintf("2 0 obj\n<< /Type /Pages /Kids [%s] /Count %d >>\nendobj\n", strings.Join(kids, " "), len(pageTexts)))

	for pageIndex, pageText := range pageTexts {
		contentNumber := 100 + pageIndex
		contents := fmt.Sprintf("%d 0 R", contentNumber)
		if contentsAsArray {
			contents = "[" + contents + "]"
		}
		document.WriteString(fmt.Sprintf("%d 0 obj\n<< /Type /Page /Parent 2 0 R /Contents %s >>\nendobj\n", 10+pageIndex, contents))
		stream := pageContent(pageText)
		document.WriteString(fmt.Sprintf("%d 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", contentNumber, len(stream), stream))
	}
	document.WriteString("trailer\n<< /Root 1 0 R >>\n%%EOF\n")
	return document.Bytes()
}

func pagesOf(text string) []string {
	pages := strings.Split(text, "\f")
	for position, page := range pages {
		pages[position] = strings.TrimSpace(page)
	}
	return pages
}

func assertPages(t *testing.T, text string, want ...string) {
	t.Helper()
	if got := pagesOf(text); !slices.Equal(got, want) {
		t.Fatalf("pages = %q, want %q", got, want)
	}
}

func TestExtractPDFSeparatesPagesWithFormFeed(t *testing.T) {
	for _, contentsAsArray := range []bool{false, true} {
		pdf := buildPagedPDF(t, []string{"primeira pagina", "segunda pagina", "terceira pagina"}, contentsAsArray)
		result, err := extractFixture(t, "paged.pdf", pdf)
		if err != nil {
			t.Fatal(err)
		}
		assertPages(t, result.Text, "primeira pagina", "segunda pagina", "terceira pagina")
	}
}

func TestExtractPDFKeepsPageNumbersWhenLeadingPagesAreEmpty(t *testing.T) {
	result, err := extractFixture(t, "blank-first.pdf", buildPagedPDF(t, []string{"", "", "conteudo"}, false))
	if err != nil {
		t.Fatal(err)
	}
	assertPages(t, result.Text, "", "", "conteudo")
}

func TestExtractPDFFollowsPageTreeOrderNotFileOrder(t *testing.T) {
	var document bytes.Buffer
	document.WriteString("%PDF-1.4\n")
	document.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	document.WriteString("5 0 obj\n<< /Type /Page /Parent 2 0 R /Contents 7 0 R >>\nendobj\n")
	document.WriteString("7 0 obj\n<< /Length 30 >>\nstream\n" + pageContent("fim") + "\nendstream\nendobj\n")
	document.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /Contents 8 0 R >>\nendobj\n")
	document.WriteString("8 0 obj\n<< /Length 30 >>\nstream\n" + pageContent("inicio") + "\nendstream\nendobj\n")
	document.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R 4 0 R] /Count 2 >>\nendobj\n")
	document.WriteString("4 0 obj\n<< /Type /Pages /Parent 2 0 R /Kids [5 0 R] /Count 1 >>\nendobj\n")

	result, err := extractFixture(t, "tree.pdf", document.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	assertPages(t, result.Text, "inicio", "fim")
}

func TestExtractPDFReadsPageTreeFromObjectStream(t *testing.T) {
	members := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R 4 0 R] /Count 2 >>",
		"<< /Type /Page /Parent 2 0 R /Contents 10 0 R >>",
		"<< /Type /Page /Parent 2 0 R /Contents 11 0 R >>",
	}
	numbers := []int{1, 2, 3, 4}
	var header, body strings.Builder
	for position, member := range members {
		header.WriteString(fmt.Sprintf("%d %d ", numbers[position], body.Len()))
		body.WriteString(member + " ")
	}
	objectStream := deflate(t, []byte(header.String()+body.String()))

	var document bytes.Buffer
	document.WriteString("%PDF-1.5\n")
	document.WriteString(fmt.Sprintf("9 0 obj\n<< /Type /ObjStm /N 4 /First %d /Length %d /Filter /FlateDecode >>\nstream\n", header.Len(), len(objectStream)))
	document.Write(objectStream)
	document.WriteString("\nendstream\nendobj\n")
	document.WriteString("10 0 obj\n<< /Length 30 >>\nstream\n" + pageContent("alfa") + "\nendstream\nendobj\n")
	document.WriteString("11 0 obj\n<< /Length 30 >>\nstream\n" + pageContent("beta") + "\nendstream\nendobj\n")

	result, err := extractFixture(t, "objstm.pdf", document.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	assertPages(t, result.Text, "alfa", "beta")
}

func TestExtractPDFWithoutResolvablePageTreeFallsBackWithoutSeparators(t *testing.T) {
	var document bytes.Buffer
	document.WriteString("%PDF-1.4\n")
	document.WriteString("4 0 obj\n<< /Length 30 >>\nstream\n" + pageContent("um") + "\nendstream\nendobj\n")
	document.WriteString("5 0 obj\n<< /Length 30 >>\nstream\n" + pageContent("dois") + "\nendstream\nendobj\n")

	result, err := extractFixture(t, "notree.pdf", document.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result.Text, "\f") || !strings.Contains(result.Text, "um") || !strings.Contains(result.Text, "dois") {
		t.Fatalf("unexpected text %q", result.Text)
	}
}

func TestExtractPDFPageTreeCycleFallsBack(t *testing.T) {
	var document bytes.Buffer
	document.WriteString("%PDF-1.4\n")
	document.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	document.WriteString("2 0 obj\n<< /Type /Pages /Kids [2 0 R] >>\nendobj\n")
	document.WriteString("4 0 obj\n<< /Length 30 >>\nstream\n" + pageContent("solto") + "\nendstream\nendobj\n")

	result, err := extractFixture(t, "cycle.pdf", document.Bytes())
	if err != nil {
		t.Fatal(err)
	}
	if result.Text != "solto" {
		t.Fatalf("unexpected text %q", result.Text)
	}
}

func TestExtractNonPDFFormatsNeverKeepFormFeeds(t *testing.T) {
	result, err := extractFixture(t, "paged.txt", []byte("antes\fdepois"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(result.Text, "\f") {
		t.Fatalf("form feed leaked into %q", result.Text)
	}
}
