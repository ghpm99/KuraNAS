package email

import (
	"slices"
	"testing"
)

func TestPrefilterCleanMessageStaysPending(t *testing.T) {
	msg := MessageModel{
		SenderAddress: "newsletter@example.com",
		Subject:       "Your weekly digest",
		AuthResults:   AuthResults{SPF: "pass", DKIM: "pass", DMARC: "pass"},
		LinkDomains:   []string{"news.example.com"},
	}
	status, rules := prefilter(msg)
	if status != MsgStatusPending || len(rules) != 0 {
		t.Fatalf("expected pending with no rules, got %s %v", status, rules)
	}
}

func TestPrefilterDMARCFail(t *testing.T) {
	status, rules := prefilter(MessageModel{
		SenderAddress: "spoof@bank.com",
		AuthResults:   AuthResults{DMARC: "fail"},
	})
	if status != MsgStatusPrefilteredSpam || !slices.Contains(rules, "dmarc_fail") {
		t.Fatalf("expected dmarc_fail spam, got %s %v", status, rules)
	}
}

func TestPrefilterDangerousAttachment(t *testing.T) {
	for _, name := range []string{"invoice.exe", "photo.jpg.scr", "macro.js"} {
		status, rules := prefilter(MessageModel{
			AuthResults: AuthResults{DMARC: "pass"},
			Attachments: []AttachmentMeta{{Filename: name}},
		})
		if status != MsgStatusPrefilteredSpam || !slices.Contains(rules, "dangerous_attachment") {
			t.Fatalf("expected dangerous_attachment for %q, got %s %v", name, status, rules)
		}
	}
}

func TestPrefilterSafeAttachmentPasses(t *testing.T) {
	status, _ := prefilter(MessageModel{
		SenderAddress: "a@example.com",
		AuthResults:   AuthResults{SPF: "pass", DKIM: "pass", DMARC: "pass"},
		Attachments:   []AttachmentMeta{{Filename: "report.pdf"}},
	})
	if status != MsgStatusPending {
		t.Fatalf("safe attachment should stay pending, got %s", status)
	}
}

func TestPrefilterSenderLinkMismatchNeedsWeakAuth(t *testing.T) {
	base := MessageModel{
		SenderAddress: "ceo@mycompany.com",
		LinkDomains:   []string{"evil-phish.ru"},
	}

	// Strong auth: mismatch alone must NOT flag (a clean newsletter links out).
	strong := base
	strong.AuthResults = AuthResults{SPF: "pass", DKIM: "pass", DMARC: "pass"}
	if status, _ := prefilter(strong); status != MsgStatusPending {
		t.Fatalf("mismatch with strong auth should stay pending, got %s", status)
	}

	// Weak auth + mismatch: phishing tell.
	weak := base
	weak.AuthResults = AuthResults{SPF: "fail"}
	status, rules := prefilter(weak)
	if status != MsgStatusPrefilteredSpam || !slices.Contains(rules, "sender_link_mismatch") {
		t.Fatalf("expected sender_link_mismatch, got %s %v", status, rules)
	}
}

func TestPrefilterSenderLinkSameOrgPasses(t *testing.T) {
	status, _ := prefilter(MessageModel{
		SenderAddress: "noreply@example.com",
		LinkDomains:   []string{"mail.example.com"},
		AuthResults:   AuthResults{SPF: "fail"},
	})
	if status != MsgStatusPending {
		t.Fatalf("same-org subdomain link should not be a mismatch, got %s", status)
	}
}

func TestPrefilterSpamSubject(t *testing.T) {
	status, rules := prefilter(MessageModel{
		SenderAddress: "x@promo.com",
		Subject:       "CONGRATULATIONS, you are a winner! Claim your prize",
		AuthResults:   AuthResults{SPF: "pass", DKIM: "pass", DMARC: "pass"},
	})
	if status != MsgStatusPrefilteredSpam || !slices.Contains(rules, "spam_subject") {
		t.Fatalf("expected spam_subject, got %s %v", status, rules)
	}
}
