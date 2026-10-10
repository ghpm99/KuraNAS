package utils

import (
	"sort"
	"testing"
)

func TestCompareNaturalOrder(t *testing.T) {
	testCases := []struct {
		name     string
		left     string
		right    string
		expected int
	}{
		{"episode two before ten", "Ep 2", "Ep 10", -1},
		{"season episode codes", "S01E02", "S01E10", -1},
		{"case insensitive equal", "ep 1", "EP 1", 0},
		{"accent insensitive equal", "Café 1", "cafe 1", 0},
		{"leading zeros equal value", "Ep 02", "Ep 2", 0},
		{"mixed text and digits", "Show 1 Part 10", "Show 1 Part 9", 1},
		{"plain text order", "alpha", "beta", -1},
		{"prefix shorter first", "Ep", "Ep 1", -1},
		{"digits before letters", "1 abc", "a abc", -1},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			comparison := CompareNaturalOrder(testCase.left, testCase.right)
			if sign(comparison) != testCase.expected {
				t.Fatalf("CompareNaturalOrder(%q, %q) = %d, want sign %d", testCase.left, testCase.right, comparison, testCase.expected)
			}
		})
	}
}

func TestNaturalOrderLessKeepsStableTiesByOriginalOrder(t *testing.T) {
	type namedEntry struct {
		id   int
		name string
	}
	entries := []namedEntry{{1, "Ep 10"}, {2, "ep 2"}, {3, "Ep 2"}, {4, "Ep 1"}}

	sort.SliceStable(entries, func(i, j int) bool {
		return NaturalOrderLess(entries[i].name, entries[j].name)
	})

	expectedIDs := []int{4, 2, 3, 1}
	for index, entry := range entries {
		if entry.id != expectedIDs[index] {
			t.Fatalf("position %d: got id %d, want %d", index, entry.id, expectedIDs[index])
		}
	}
}

func sign(number int) int {
	if number < 0 {
		return -1
	}
	if number > 0 {
		return 1
	}
	return 0
}
