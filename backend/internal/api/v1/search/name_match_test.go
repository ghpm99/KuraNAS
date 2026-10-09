package search

import (
	"reflect"
	"testing"
)

func TestSplitSearchTermsLowercasesDedupesAndCapsTerms(t *testing.T) {
	cases := []struct {
		name     string
		query    string
		expected []string
	}{
		{"single", "Beach", []string{"beach"}},
		{"multiple spaces and tabs", "  Beach \t Holiday\n", []string{"beach", "holiday"}},
		{"duplicates ignored", "trip TRIP trip", []string{"trip"}},
		{"keeps one letter terms", "a b", []string{"a", "b"}},
		{"caps at five terms", "a b c d e f g", []string{"a", "b", "c", "d", "e"}},
		{"blank", "   ", []string{}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := splitSearchTerms(testCase.query); !reflect.DeepEqual(got, testCase.expected) {
				t.Fatalf("splitSearchTerms(%q) = %v, want %v", testCase.query, got, testCase.expected)
			}
		})
	}
}

func TestBuildNameMatchEscapesWildcardsAndPicksLongestTermAsDriver(t *testing.T) {
	match, hasTerms := buildNameMatch("100%_ Holiday")
	if !hasTerms {
		t.Fatalf("expected terms")
	}
	if match.DrivingPattern != "%holiday%" {
		t.Fatalf("driving pattern = %q", match.DrivingPattern)
	}
	expectedPatterns := []string{`%100\%\_%`, "%holiday%"}
	if !reflect.DeepEqual(match.AllPatterns, expectedPatterns) {
		t.Fatalf("all patterns = %v, want %v", match.AllPatterns, expectedPatterns)
	}
	if match.ExactName != "100%_ holiday" {
		t.Fatalf("exact name = %q", match.ExactName)
	}
	if match.PrefixPattern != `100\%\_ holiday%` {
		t.Fatalf("prefix pattern = %q", match.PrefixPattern)
	}
	if match.ContainsPattern != `%100\%\_ holiday%` {
		t.Fatalf("contains pattern = %q", match.ContainsPattern)
	}
}

func TestBuildNameMatchReportsNoTermsForBlankQuery(t *testing.T) {
	if _, hasTerms := buildNameMatch("  \t "); hasTerms {
		t.Fatalf("blank query must not produce a name match")
	}
}
