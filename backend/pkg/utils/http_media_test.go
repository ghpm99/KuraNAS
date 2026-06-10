package utils

import "testing"

func TestParseHTTPRange(t *testing.T) {
	if _, _, ok := ParseHTTPRange("bytes=0-5", 0); ok {
		t.Fatalf("expected invalid range for empty file")
	}
	if _, _, ok := ParseHTTPRange("items=0-5", 100); ok {
		t.Fatalf("expected invalid range for wrong unit")
	}
	if _, _, ok := ParseHTTPRange("bytes=", 100); ok {
		t.Fatalf("expected invalid range for empty value")
	}
	if _, _, ok := ParseHTTPRange("bytes=abc-5", 100); ok {
		t.Fatalf("expected invalid range for bad start")
	}
	if _, _, ok := ParseHTTPRange("bytes=5-abc", 100); ok {
		t.Fatalf("expected invalid range for bad end")
	}
	if _, _, ok := ParseHTTPRange("bytes=50-10", 100); ok {
		t.Fatalf("expected invalid range for inverted bounds")
	}
	if _, _, ok := ParseHTTPRange("bytes=-0", 100); ok {
		t.Fatalf("expected invalid range for zero suffix")
	}
	if start, end, ok := ParseHTTPRange("bytes=0-5,10-20", 100); !ok || start != 0 || end != 5 {
		t.Fatalf("expected first range only, got start=%d end=%d ok=%v", start, end, ok)
	}
	if start, end, ok := ParseHTTPRange("bytes=-200", 100); !ok || start != 0 || end != 99 {
		t.Fatalf("expected clamped suffix range, got start=%d end=%d ok=%v", start, end, ok)
	}
	if start, end, ok := ParseHTTPRange("bytes=10-500", 100); !ok || start != 10 || end != 99 {
		t.Fatalf("expected clamped end, got start=%d end=%d ok=%v", start, end, ok)
	}
	if start, end, ok := ParseHTTPRange("bytes=10-", 100); !ok || start != 10 || end != 99 {
		t.Fatalf("expected open-ended range, got start=%d end=%d ok=%v", start, end, ok)
	}
}

func TestContentTypeByFormat(t *testing.T) {
	if got := ContentTypeByFormat("", "audio/mpeg"); got != "audio/mpeg" {
		t.Fatalf("expected fallback for empty format, got %s", got)
	}
	if got := ContentTypeByFormat(".mp3", "application/octet-stream"); got == "application/octet-stream" {
		t.Fatalf("expected resolved content type for .mp3")
	}
	if got := ContentTypeByFormat("mp3", "application/octet-stream"); got == "application/octet-stream" {
		t.Fatalf("expected resolved content type for extension without dot")
	}
	if got := ContentTypeByFormat("zzz", "video/mp4"); got != "video/mp4" {
		t.Fatalf("expected fallback for unknown format, got %s", got)
	}
}
