package documenttext

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestBuildSnippetReturnsShortTextWhole(t *testing.T) {
	snippet := BuildSnippet("  Contrato   de\n locação ", []string{"contrato"})
	if snippet != "Contrato de locação" {
		t.Fatalf("unexpected snippet %q", snippet)
	}
}

func TestBuildSnippetCentersOnAccentInsensitiveMatch(t *testing.T) {
	text := strings.Repeat("palavra comum ", 40) + "Relatório financeiro anual " + strings.Repeat("final ", 60)

	snippet := BuildSnippet(text, []string{"relatorio"})

	if !strings.Contains(snippet, "Relatório financeiro") {
		t.Fatalf("snippet lost the match: %q", snippet)
	}
	if utf8.RuneCountInString(snippet) > snippetLength {
		t.Fatalf("snippet longer than %d: %d", snippetLength, utf8.RuneCountInString(snippet))
	}
}

func TestBuildSnippetPicksEarliestOfSeveralTerms(t *testing.T) {
	text := strings.Repeat("a ", 100) + "primeiro " + strings.Repeat("b ", 100) + "segundo " + strings.Repeat("c ", 100)

	snippet := BuildSnippet(text, []string{"segundo", "primeiro"})

	if !strings.Contains(snippet, "primeiro") || strings.Contains(snippet, "segundo") {
		t.Fatalf("expected snippet around earliest term, got %q", snippet)
	}
}

func TestBuildSnippetMatchAtTextEdges(t *testing.T) {
	body := strings.Repeat("x ", 200)

	startSnippet := BuildSnippet("inicio "+body, []string{"inicio"})
	if !strings.HasPrefix(startSnippet, "inicio") {
		t.Fatalf("expected snippet to start at the match, got %q", startSnippet)
	}

	endSnippet := BuildSnippet(body+" fim", []string{"fim"})
	if !strings.HasSuffix(endSnippet, "fim") {
		t.Fatalf("expected snippet to end at the match, got %q", endSnippet)
	}
}

func TestBuildSnippetWithoutMatchReturnsTextStart(t *testing.T) {
	text := strings.Repeat("abcdefghij ", 50)

	snippet := BuildSnippet(text, []string{"inexistente"})

	if !strings.HasPrefix(text, snippet) || utf8.RuneCountInString(snippet) > snippetLength {
		t.Fatalf("unexpected fallback snippet %q", snippet)
	}
}

func TestBuildSnippetMultibyteWindowKeepsValidUTF8(t *testing.T) {
	text := strings.Repeat("ção ", 100) + "ÁRVORE " + strings.Repeat("ação ", 100)

	snippet := BuildSnippet(text, []string{"arvore"})

	if !utf8.ValidString(snippet) || !strings.Contains(snippet, "ÁRVORE") {
		t.Fatalf("unexpected snippet %q", snippet)
	}
}
