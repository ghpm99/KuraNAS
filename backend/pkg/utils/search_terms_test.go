package utils

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
			if got := SplitSearchTerms(testCase.query); !reflect.DeepEqual(got, testCase.expected) {
				t.Fatalf("SplitSearchTerms(%q) = %v, want %v", testCase.query, got, testCase.expected)
			}
		})
	}
}

func TestBuildSearchTermPatternsDedupesAndPicksLongestTermAsDriver(t *testing.T) {
	if _, hasTerms := BuildSearchTermPatterns("   "); hasTerms {
		t.Fatalf("blank text must not produce patterns")
	}

	patterns, hasTerms := BuildSearchTermPatterns("Queen queen Bohemian")
	if !hasTerms || len(patterns.AllPatterns) != 2 {
		t.Fatalf("expected 2 distinct terms, got %+v", patterns)
	}
	if patterns.DrivingPattern != "%bohemian%" {
		t.Fatalf("driving pattern should be longest term, got %q", patterns.DrivingPattern)
	}

	capped, _ := BuildSearchTermPatterns("a b c d e f g")
	if len(capped.AllPatterns) != MaxSearchTerms {
		t.Fatalf("expected cap of %d terms, got %d", MaxSearchTerms, len(capped.AllPatterns))
	}
}
