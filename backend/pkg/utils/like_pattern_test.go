package utils

import "testing"

func TestBuildLikePatternsEscapeWildcards(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		contains string
		prefix   string
	}{
		{"plain", "abc", "%abc%", "abc%"},
		{"percent", "100%", `%100\%%`, `100\%%`},
		{"underscore", "a_b", `%a\_b%`, `a\_b%`},
		{"backslash", `a\b`, `%a\\b%`, `a\\b%`},
		{"empty", "", "%%", "%"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := BuildContainsLikePattern(testCase.input); got != testCase.contains {
				t.Fatalf("contains = %q, want %q", got, testCase.contains)
			}
			if got := BuildPrefixLikePattern(testCase.input); got != testCase.prefix {
				t.Fatalf("prefix = %q, want %q", got, testCase.prefix)
			}
		})
	}
}
