package search

import (
	"reflect"
	"testing"
)

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
