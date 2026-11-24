package engine

import "testing"

func TestParseDownloadProgressLine(t *testing.T) {
	cases := []struct {
		name         string
		line         string
		wantOK       bool
		wantPercent  int
		wantSpeed    int64
		wantHasSpeed bool
		wantETA      int
		wantHasETA   bool
	}{
		{name: "full line", line: "[download]  42.3% of ~12.34MiB at 1.00MiB/s ETA 00:12", wantOK: true, wantPercent: 42, wantSpeed: 1 << 20, wantHasSpeed: true, wantETA: 12, wantHasETA: true},
		{name: "hours in eta", line: "[download]   1.0% of 2.00GiB at 500.00KiB/s ETA 01:02:03", wantOK: true, wantPercent: 1, wantSpeed: 512000, wantHasSpeed: true, wantETA: 3723, wantHasETA: true},
		{name: "unknown size and speed", line: "[download]   5.0% of Unknown at Unknown B/s ETA Unknown", wantOK: true, wantPercent: 5},
		{name: "missing eta", line: "[download] 100% of 10.00MiB at 2.00MiB/s", wantOK: true, wantPercent: 100, wantSpeed: 2 << 20, wantHasSpeed: true},
		{name: "final line without speed", line: "[download] 100% of 10.00MiB in 00:05", wantOK: true, wantPercent: 100},
		{name: "not a download line", line: "[info] something 50%", wantOK: false},
		{name: "download line without percent", line: "[download] Destination: a.mp3", wantOK: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			progress, ok := parseDownloadProgressLine(tc.line)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if progress.Percent != tc.wantPercent {
				t.Errorf("percent = %d, want %d", progress.Percent, tc.wantPercent)
			}
			if (progress.SpeedBytesPerSecond != nil) != tc.wantHasSpeed || (tc.wantHasSpeed && *progress.SpeedBytesPerSecond != tc.wantSpeed) {
				t.Errorf("speed = %v, want %d (present=%v)", progress.SpeedBytesPerSecond, tc.wantSpeed, tc.wantHasSpeed)
			}
			if (progress.ETASeconds != nil) != tc.wantHasETA || (tc.wantHasETA && *progress.ETASeconds != tc.wantETA) {
				t.Errorf("eta = %v, want %d (present=%v)", progress.ETASeconds, tc.wantETA, tc.wantHasETA)
			}
		})
	}
}

func TestParseDownloadDestinationTitle(t *testing.T) {
	title, ok := parseDownloadDestinationTitle("[download] Destination: /tmp/kuranas-fetch-1/My_Clip.f251.webm")
	if !ok || title != "My_Clip.f251" {
		t.Fatalf("title = %q ok=%v", title, ok)
	}
	if _, ok := parseDownloadDestinationTitle("[download]  10.0% of 1MiB"); ok {
		t.Fatal("a progress line carries no destination")
	}
}
