package music

import (
	"encoding/json"
	"testing"
)

func TestParseNumberWithTotal(t *testing.T) {
	cases := []struct {
		rawNumbering  string
		expectedNum   *int
		expectedTotal *int
	}{
		{"3/12", intPointer(3), intPointer(12)},
		{"1/2", intPointer(1), intPointer(2)},
		{" 07 / 10 ", intPointer(7), intPointer(10)},
		{"5", intPointer(5), nil},
		{"4/", intPointer(4), nil},
		{"", nil, nil},
		{"abc", nil, nil},
	}

	for _, testCase := range cases {
		number, total := ParseNumberWithTotal(testCase.rawNumbering)
		if !equalOptionalInt(number, testCase.expectedNum) || !equalOptionalInt(total, testCase.expectedTotal) {
			t.Errorf("ParseNumberWithTotal(%q) = %v, %v", testCase.rawNumbering, number, total)
		}
	}
}

func TestResolveNumberingFillsTrackAndDisc(t *testing.T) {
	metadata := AudioMetadataModel{TrackNumber: "3/12", DiscNumberText: "1/2"}

	metadata.ResolveNumbering()

	if !equalOptionalInt(metadata.TrackNo, intPointer(3)) || !equalOptionalInt(metadata.TrackTotal, intPointer(12)) {
		t.Fatalf("unexpected track numbering: %v/%v", metadata.TrackNo, metadata.TrackTotal)
	}
	if !equalOptionalInt(metadata.DiscNumber, intPointer(1)) || !equalOptionalInt(metadata.DiscTotal, intPointer(2)) {
		t.Fatalf("unexpected disc numbering: %v/%v", metadata.DiscNumber, metadata.DiscTotal)
	}
}

func TestAudioMetadataDecodesDiscNumberFromScriptJSON(t *testing.T) {
	metadata := AudioMetadataModel{}
	if err := json.Unmarshal([]byte(`{"title":"T","track_number":"3/12","disc_number":"1/2","lyrics":"la"}`), &metadata); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if metadata.DiscNumberText != "1/2" || metadata.TrackNumber != "3/12" || metadata.Lyrics != "la" {
		t.Fatalf("unexpected decode: %+v", metadata)
	}
}

func intPointer(number int) *int { return &number }

func equalOptionalInt(actual *int, expected *int) bool {
	if actual == nil || expected == nil {
		return actual == expected
	}
	return *actual == *expected
}
