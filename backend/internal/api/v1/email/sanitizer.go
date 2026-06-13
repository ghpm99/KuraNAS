package email

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"golang.org/x/net/html"
)

// urlRe matches absolute http(s) URLs in either markup (href="...") or plain
// text. Matches are used ONLY to record their host as a bare domain — the URLs
// themselves are never fetched (hard rule of the e-mail feature).
var urlRe = regexp.MustCompile(`(?i)https?://[^\s"'<>)\]]+`)

// extractLinkDomains returns the sorted, de-duplicated set of hostnames of every
// http(s) URL found in the raw body. The full URLs are intentionally discarded:
// only the domain is kept, as evidence for the pre-filter and analysis (task 16).
func extractLinkDomains(raw string) []string {
	matches := urlRe.FindAllString(raw, -1)
	domains := make([]string, 0, len(matches))
	for _, match := range matches {
		parsed, err := url.Parse(strings.TrimRight(match, ".,;"))
		if err != nil {
			continue
		}
		host := strings.ToLower(parsed.Hostname())
		if host != "" {
			domains = append(domains, host)
		}
	}
	return sortedUnique(domains)
}

const (
	// maxBodyBytes is the hard ceiling on the stored plain-text body (hard rule
	// of the e-mail feature). Anything beyond it is dropped.
	maxBodyBytes = 16 * 1024
	// maxRawBytes bounds the raw input handed to the HTML parser, so a hostile
	// multi-megabyte body cannot turn parsing into a DoS. It is generous
	// relative to maxBodyBytes because markup is mostly tags.
	maxRawBytes = 512 * 1024
	// snippetRunes is the preview length (in runes) shown in lean listings.
	snippetRunes = 280
)

// sanitizeBody turns a raw message body into safe, stored plain text: HTML is
// reduced to its text nodes (scripts/styles discarded), invisible/bidi Unicode
// is stripped, whitespace is collapsed and the result is capped at 16 KB. It
// returns the body and a short snippet for listings. URLs are never followed —
// see extractLinkDomains for how their domains are recorded as data.
func sanitizeBody(raw string, isHTML bool) (body string, snippet string) {
	if len(raw) > maxRawBytes {
		raw = raw[:maxRawBytes]
	}

	text := raw
	if isHTML {
		text = htmlToText(raw)
	}

	text = stripInvisible(text)
	text = collapseWhitespace(text)
	text = truncateBytes(text, maxBodyBytes)

	return text, buildSnippet(text)
}

// htmlToText walks the parse tree and concatenates text nodes, skipping the
// content of <script>, <style> and similar non-display elements. Block-level
// tags inject a newline so paragraphs do not run together.
func htmlToText(raw string) string {
	root, err := html.Parse(strings.NewReader(raw))
	if err != nil {
		return raw
	}

	var builder strings.Builder
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "script", "style", "head", "title", "noscript", "iframe", "object", "embed":
				return
			case "br", "p", "div", "li", "tr", "table", "h1", "h2", "h3", "h4", "h5", "h6", "blockquote":
				builder.WriteByte('\n')
			}
		}
		if node.Type == html.TextNode {
			builder.WriteString(node.Data)
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}
	walk(root)
	return builder.String()
}

// stripInvisible removes zero-width characters, bidirectional control codes and
// other invisible formatting runes that could hide content or smuggle a prompt
// past the analysis stage (task 16). Standard whitespace is preserved.
func stripInvisible(s string) string {
	return strings.Map(func(r rune) rune {
		if r == '\n' || r == '\t' || r == ' ' {
			return r
		}
		// unicode.Cf (format) covers the zero-width space/joiners (U+200B–200D),
		// word joiner (U+2060), BOM (U+FEFF), soft hyphen (U+00AD) and every
		// bidirectional control (U+202A–202E, U+2066–2069). Dropping Cf, all
		// other control runes and CR removes anything invisible while keeping
		// the displayable text and the whitespace handled above.
		if r == '\r' || unicode.Is(unicode.Cf, r) || unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}

// collapseWhitespace trims runs of spaces/tabs and limits blank lines so the
// stored body is compact without losing paragraph structure.
func collapseWhitespace(s string) string {
	lines := strings.Split(s, "\n")
	cleaned := make([]string, 0, len(lines))
	blankRun := 0
	for _, line := range lines {
		line = strings.TrimSpace(strings.Join(strings.Fields(line), " "))
		if line == "" {
			blankRun++
			if blankRun > 1 {
				continue
			}
		} else {
			blankRun = 0
		}
		cleaned = append(cleaned, line)
	}
	return strings.TrimSpace(strings.Join(cleaned, "\n"))
}

// truncateBytes caps s at n bytes without splitting a multi-byte rune.
func truncateBytes(s string, n int) string {
	if len(s) <= n {
		return s
	}
	cut := n
	for cut > 0 && !utf8RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}

func utf8RuneStart(b byte) bool {
	// Continuation bytes are 0b10xxxxxx; any other byte starts a rune.
	return b&0xC0 != 0x80
}

func buildSnippet(body string) string {
	oneLine := strings.Join(strings.Fields(body), " ")
	runes := []rune(oneLine)
	if len(runes) <= snippetRunes {
		return oneLine
	}
	return strings.TrimSpace(string(runes[:snippetRunes]))
}

// sortedUnique returns the de-duplicated, sorted set of non-empty values. It
// keeps link-domain lists stable for storage and tests.
func sortedUnique(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v == "" || seen[v] {
			continue
		}
		seen[v] = true
		out = append(out, v)
	}
	sort.Strings(out)
	return out
}
