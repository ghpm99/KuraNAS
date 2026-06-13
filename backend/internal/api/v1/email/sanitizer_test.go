package email

import (
	"strings"
	"testing"
)

func TestSanitizeBodyStripsScriptAndTags(t *testing.T) {
	html := `<html><head><style>p{color:red}</style></head>
		<body><script>alert('x')</script><p>Hello <b>world</b></p><p>Second line</p></body></html>`

	body, snippet := sanitizeBody(html, true)
	if strings.Contains(body, "<") || strings.Contains(body, "alert") || strings.Contains(body, "color") {
		t.Fatalf("body still contains markup/script: %q", body)
	}
	if !strings.Contains(body, "Hello world") || !strings.Contains(body, "Second line") {
		t.Fatalf("expected text content preserved, got %q", body)
	}
	if snippet == "" || strings.Contains(snippet, "\n") {
		t.Fatalf("snippet should be a single non-empty line: %q", snippet)
	}
}

func TestSanitizeBodyRemovesInvisibleUnicode(t *testing.T) {
	// Built from explicit code points so there are no invisible bytes in the
	// source: zero-width space (U+200B), word joiner (U+2060), bidi override
	// (U+202E) and BOM (U+FEFF) interleaved with visible text.
	zwsp, wj, rlo, bom := string(rune(0x200B)), string(rune(0x2060)), string(rune(0x202E)), string(rune(0xFEFF))
	raw := "He" + zwsp + "ll" + wj + "o" + rlo + "there" + bom + "!"
	body, _ := sanitizeBody(raw, false)
	if body != "Hellothere!" {
		t.Fatalf("invisible runes not stripped: %q", body)
	}
}

func TestSanitizeBodyTruncatesTo16KB(t *testing.T) {
	raw := strings.Repeat("a", maxBodyBytes+5000)
	body, _ := sanitizeBody(raw, false)
	if len(body) > maxBodyBytes {
		t.Fatalf("body not truncated: %d bytes", len(body))
	}
}

func TestSanitizeBodyBoundsRawInputBeforeParse(t *testing.T) {
	// A pathologically large HTML body must not blow up; it is capped before
	// parsing and the result stays within the body ceiling.
	raw := "<p>" + strings.Repeat("x", maxRawBytes*2) + "</p>"
	body, _ := sanitizeBody(raw, true)
	if len(body) > maxBodyBytes {
		t.Fatalf("oversized body not capped: %d bytes", len(body))
	}
}

func TestExtractLinkDomains(t *testing.T) {
	raw := `Click <a href="https://login.example.com/path?x=1">here</a> or visit
		http://Tracker.Ads.net/click. Repeat https://login.example.com/other.`

	got := extractLinkDomains(raw)
	want := []string{"login.example.com", "tracker.ads.net"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}

func TestExtractLinkDomainsNone(t *testing.T) {
	if got := extractLinkDomains("no links here, just text"); len(got) != 0 {
		t.Fatalf("expected no domains, got %v", got)
	}
}
