package mailfetch

import "testing"

func TestParseAuthResults(t *testing.T) {
	header := "mx.google.com; spf=pass (google.com: domain of a@b.com) smtp.mailfrom=a@b.com; " +
		"dkim=pass header.i=@b.com; dmarc=fail (p=NONE sp=NONE dis=NONE) header.from=b.com"

	got := ParseAuthResults(header)
	if got.SPF != "pass" || got.DKIM != "pass" || got.DMARC != "fail" {
		t.Fatalf("unexpected verdicts: %+v", got)
	}
}

func TestParseAuthResultsFirstWins(t *testing.T) {
	got := ParseAuthResults("spf=pass; dmarc=pass", "spf=fail; dmarc=fail")
	if got.SPF != "pass" || got.DMARC != "pass" {
		t.Fatalf("expected top-most header to win, got %+v", got)
	}
}

func TestParseAuthResultsEmpty(t *testing.T) {
	got := ParseAuthResults("no auth methods here")
	if got.SPF != "" || got.DKIM != "" || got.DMARC != "" {
		t.Fatalf("expected empty verdicts, got %+v", got)
	}
}
