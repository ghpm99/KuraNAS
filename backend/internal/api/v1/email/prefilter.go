package email

import (
	"regexp"
	"strings"
)

// The pre-filter is the determinant gate of the e-mail threat model: spam and
// obvious phishing flagged here become prefiltered_spam and NEVER reach the LLM
// (hard rule 7), cutting both cost and prompt-injection surface. Every rule is a
// pure, testable function; the names of the rules that fired are stored as
// evidence for the analysis stage (task 16).

// dangerousExtRe matches attachment names whose final extension is executable.
var dangerousExtRe = regexp.MustCompile(`(?i)\.(exe|scr|js|jse|vbs|vbe|bat|cmd|com|pif|hta|jar|msi|ps1|wsf|lnk)$`)

// doubleExtRe matches the classic "invoice.pdf.exe" disguise: a document-looking
// extension immediately followed by an executable one.
var doubleExtRe = regexp.MustCompile(`(?i)\.(pdf|docx?|xlsx?|pptx?|jpe?g|png|gif|txt|zip|rar)\.(exe|scr|js|jse|vbs|vbe|bat|cmd|com|pif|hta|jar|msi|ps1|wsf|lnk)$`)

// spamSubjectRe matches a few unambiguous spam/scam subject patterns. It is
// deliberately conservative — false positives here silently hide real mail.
var spamSubjectRe = regexp.MustCompile(`(?i)(viagra|cialis|\bv[i1]agr[a@]\b|you('?ve| have) won|claim your (prize|reward)|free money|act now|nigerian prince|crypto(currency)? giveaway|verify your account now|congratulations,? you('?re| are) (a )?winner)`)

type prefilterRule struct {
	name string
	test func(MessageModel) bool
}

var prefilterRules = []prefilterRule{
	{
		name: "dmarc_fail",
		test: func(m MessageModel) bool {
			return strings.EqualFold(m.AuthResults.DMARC, "fail")
		},
	},
	{
		name: "dangerous_attachment",
		test: func(m MessageModel) bool {
			for _, attachment := range m.Attachments {
				name := strings.TrimSpace(attachment.Filename)
				if dangerousExtRe.MatchString(name) || doubleExtRe.MatchString(name) {
					return true
				}
			}
			return false
		},
	},
	{
		// Sender domain absent from the link domains is a phishing tell, but only
		// when paired with another signal (failed/missing SPF or DKIM): a clean
		// newsletter routinely links elsewhere, so the mismatch alone is not enough.
		name: "sender_link_mismatch",
		test: func(m MessageModel) bool {
			senderDomain := domainOf(m.SenderAddress)
			if senderDomain == "" || len(m.LinkDomains) == 0 {
				return false
			}
			if linkContainsDomain(m.LinkDomains, senderDomain) {
				return false
			}
			authWeak := !strings.EqualFold(m.AuthResults.SPF, "pass") ||
				!strings.EqualFold(m.AuthResults.DKIM, "pass")
			return authWeak
		},
	},
	{
		name: "spam_subject",
		test: func(m MessageModel) bool {
			return spamSubjectRe.MatchString(m.Subject)
		},
	},
}

// prefilter runs every rule and reports the resulting status plus the names of
// the rules that fired (empty when the message stays pending).
func prefilter(m MessageModel) (MessageStatus, []string) {
	var fired []string
	for _, rule := range prefilterRules {
		if rule.test(m) {
			fired = append(fired, rule.name)
		}
	}
	if len(fired) > 0 {
		return MsgStatusPrefilteredSpam, fired
	}
	return MsgStatusPending, nil
}

func domainOf(address string) string {
	at := strings.LastIndex(address, "@")
	if at < 0 || at == len(address)-1 {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(address[at+1:]))
}

// linkContainsDomain reports whether the sender domain matches any link domain,
// treating a subdomain (mail.example.com) and its registrable parent
// (example.com) as the same organization.
func linkContainsDomain(linkDomains []string, senderDomain string) bool {
	for _, domain := range linkDomains {
		if domain == senderDomain ||
			strings.HasSuffix(domain, "."+senderDomain) ||
			strings.HasSuffix(senderDomain, "."+domain) {
			return true
		}
	}
	return false
}
