package autoshutdown

import "testing"

func TestNormalizeTime(t *testing.T) {
	valid := map[string]string{
		"0:0":   "00:00",
		"7:5":   "07:05",
		"23:59": "23:59",
		" 9:30": "09:30",
	}
	for input, want := range valid {
		got, err := normalizeTime(input)
		if err != nil || got != want {
			t.Fatalf("normalizeTime(%q) = %q, %v; want %q", input, got, err, want)
		}
	}

	invalid := []string{"", "24:00", "12:60", "-1:00", "noon", "12", "12:00:00", "ab:cd"}
	for _, input := range invalid {
		if _, err := normalizeTime(input); err == nil {
			t.Fatalf("normalizeTime(%q) expected error", input)
		}
	}
}

func TestSecondsToTime(t *testing.T) {
	cases := map[float64]string{
		0:        "00:00",
		10800:    "03:00",
		10800.49: "03:00",
		86399:    "23:59",
		86400:    "00:00", // wraps a full day
		-60:      "23:59", // negative wraps back into the day
	}
	for seconds, want := range cases {
		if got := secondsToTime(seconds); got != want {
			t.Fatalf("secondsToTime(%v) = %q; want %q", seconds, got, want)
		}
	}
}

func TestDecodeSettingsBackfillsDefaults(t *testing.T) {
	settings, err := decodeSettings(`{"enabled":true}`)
	if err != nil {
		t.Fatalf("decodeSettings: %v", err)
	}
	if settings.Time != defaultTime || settings.GracePeriodSeconds != defaultGracePeriodSeconds {
		t.Fatalf("expected defaults backfilled, got %+v", settings)
	}
}

func TestDecodeSettingsInvalidJSON(t *testing.T) {
	if _, err := decodeSettings(`{nope`); err == nil {
		t.Fatal("expected error decoding invalid JSON")
	}
}
