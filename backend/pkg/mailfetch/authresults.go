package mailfetch

import (
	"regexp"
	"strings"
)

// authMethodRe captures the verdict word that follows an auth method, e.g.
// "spf=pass", "dkim=fail", "dmarc=none". The verdict is the first token after
// the '=' (letters only), ignoring any "(...)" comment or "key=value" detail
// that follows.
var authMethodRe = regexp.MustCompile(`(?i)\b(spf|dkim|dmarc)\s*=\s*([a-z]+)`)

// ParseAuthResults extracts the SPF/DKIM/DMARC verdicts from one or more
// Authentication-Results header values. The first verdict seen for each method
// wins (the top-most header is the receiving MTA's own check). Values are
// lower-cased for stable comparison; unknown methods are ignored.
func ParseAuthResults(headerValues ...string) AuthResults {
	var results AuthResults
	for _, header := range headerValues {
		for _, match := range authMethodRe.FindAllStringSubmatch(header, -1) {
			method := strings.ToLower(match[1])
			verdict := strings.ToLower(match[2])
			switch method {
			case "spf":
				if results.SPF == "" {
					results.SPF = verdict
				}
			case "dkim":
				if results.DKIM == "" {
					results.DKIM = verdict
				}
			case "dmarc":
				if results.DMARC == "" {
					results.DMARC = verdict
				}
			}
		}
	}
	return results
}
