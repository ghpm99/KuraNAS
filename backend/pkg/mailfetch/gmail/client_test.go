package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newGmailServer(t *testing.T, fullMessage string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/attachments"):
			// Attachment content must never be downloaded.
			t.Errorf("attachment route must never be called: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		case strings.HasSuffix(r.URL.Path, "/messages") || strings.HasSuffix(r.URL.Path, "/messages/"):
			fmt.Fprint(w, `{"messages":[{"id":"m1"}]}`)
		case strings.Contains(r.URL.Path, "/messages/m1"):
			fmt.Fprint(w, fullMessage)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGmailListNewMessages(t *testing.T) {
	plain := base64.URLEncoding.EncodeToString([]byte("plain body text"))
	full := fmt.Sprintf(`{
		"id": "m1",
		"internalDate": "1700000000000",
		"payload": {
			"mimeType": "multipart/mixed",
			"headers": [
				{"name": "From", "value": "Alice Example <alice@example.com>"},
				{"name": "Subject", "value": "Hello there"},
				{"name": "Authentication-Results", "value": "mx; spf=pass; dkim=pass; dmarc=pass"}
			],
			"parts": [
				{"mimeType": "text/plain", "body": {"size": 15, "data": "%s"}},
				{"mimeType": "application/pdf", "filename": "invoice.pdf", "body": {"size": 2048, "attachmentId": "att-1"}}
			]
		}
	}`, plain)

	server := newGmailServer(t, full)
	defer server.Close()

	client := NewClient(server.URL)
	messages, err := client.ListNewMessages(context.Background(), "tok", time.Unix(1699000000, 0), 50)
	if err != nil {
		t.Fatalf("ListNewMessages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.SenderAddress != "alice@example.com" || msg.SenderName != "Alice Example" {
		t.Errorf("unexpected sender: %q / %q", msg.SenderName, msg.SenderAddress)
	}
	if msg.Subject != "Hello there" {
		t.Errorf("unexpected subject: %q", msg.Subject)
	}
	if msg.Body != "plain body text" || msg.BodyIsHTML {
		t.Errorf("unexpected body: %q html=%v", msg.Body, msg.BodyIsHTML)
	}
	if msg.AuthResults.DMARC != "pass" {
		t.Errorf("unexpected auth results: %+v", msg.AuthResults)
	}
	if len(msg.Attachments) != 1 || msg.Attachments[0].Filename != "invoice.pdf" || msg.Attachments[0].Size != 2048 {
		t.Errorf("unexpected attachments: %+v", msg.Attachments)
	}
	if msg.ReceivedAt.IsZero() {
		t.Error("expected a parsed received time")
	}
}

func TestGmailMaxZeroSkips(t *testing.T) {
	server := newGmailServer(t, "{}")
	defer server.Close()

	client := NewClient(server.URL)
	messages, err := client.ListNewMessages(context.Background(), "tok", time.Time{}, 0)
	if err != nil {
		t.Fatalf("ListNewMessages: %v", err)
	}
	if messages != nil {
		t.Fatalf("expected nil for max=0, got %v", messages)
	}
}

func TestGmailFallsBackToHTMLBody(t *testing.T) {
	html := base64.URLEncoding.EncodeToString([]byte("<p>hi</p>"))
	full := fmt.Sprintf(`{
		"id": "m1",
		"internalDate": "1700000000000",
		"payload": {
			"mimeType": "text/html",
			"headers": [{"name": "From", "value": "b@example.com"}],
			"body": {"size": 9, "data": "%s"}
		}
	}`, html)

	server := newGmailServer(t, full)
	defer server.Close()

	client := NewClient(server.URL)
	messages, err := client.ListNewMessages(context.Background(), "tok", time.Time{}, 10)
	if err != nil {
		t.Fatalf("ListNewMessages: %v", err)
	}
	if len(messages) != 1 || messages[0].Body != "<p>hi</p>" || !messages[0].BodyIsHTML {
		t.Fatalf("expected html body, got %+v", messages)
	}
}
