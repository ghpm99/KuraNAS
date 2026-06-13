package graph

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGraphListNewMessages(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/attachments"):
			// $select must keep contentBytes out; the test asserts no field
			// other than metadata is ever requested by returning only metadata.
			if strings.Contains(r.URL.RawQuery, "contentBytes") {
				t.Errorf("attachment content must never be requested: %s", r.URL.RawQuery)
			}
			fmt.Fprint(w, `{"value":[{"name":"report.pdf","contentType":"application/pdf","size":1024}]}`)
		case strings.HasSuffix(r.URL.Path, "/messages"):
			fmt.Fprint(w, `{"value":[{
				"id": "g1",
				"subject": "Quarterly",
				"receivedDateTime": "2024-01-02T15:04:05Z",
				"hasAttachments": true,
				"from": {"emailAddress": {"name": "Bob", "address": "BOB@Example.com"}},
				"body": {"contentType": "html", "content": "<p>body</p>"},
				"internetMessageHeaders": [
					{"name": "Authentication-Results", "value": "spf=pass; dkim=pass; dmarc=pass"}
				]
			}]}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL)
	messages, err := client.ListNewMessages(context.Background(), "tok", time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC), 25)
	if err != nil {
		t.Fatalf("ListNewMessages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}

	msg := messages[0]
	if msg.SenderAddress != "bob@example.com" || msg.SenderName != "Bob" {
		t.Errorf("unexpected sender: %q / %q", msg.SenderName, msg.SenderAddress)
	}
	if msg.Subject != "Quarterly" || msg.Body != "<p>body</p>" || !msg.BodyIsHTML {
		t.Errorf("unexpected message fields: %+v", msg)
	}
	if msg.AuthResults.DMARC != "pass" {
		t.Errorf("unexpected auth results: %+v", msg.AuthResults)
	}
	if len(msg.Attachments) != 1 || msg.Attachments[0].Filename != "report.pdf" || msg.Attachments[0].Size != 1024 {
		t.Errorf("unexpected attachments: %+v", msg.Attachments)
	}
	if !msg.ReceivedAt.Equal(time.Date(2024, 1, 2, 15, 4, 5, 0, time.UTC)) {
		t.Errorf("unexpected received time: %v", msg.ReceivedAt)
	}
}

func TestGraphNoAttachmentsSkipsAttachmentCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/attachments") {
			t.Errorf("attachment route must not be called when hasAttachments is false: %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"value":[{
			"id": "g2",
			"subject": "No files",
			"receivedDateTime": "2024-01-02T15:04:05Z",
			"hasAttachments": false,
			"from": {"emailAddress": {"name": "C", "address": "c@example.com"}},
			"body": {"contentType": "text", "content": "plain"}
		}]}`)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	messages, err := client.ListNewMessages(context.Background(), "tok", time.Time{}, 25)
	if err != nil {
		t.Fatalf("ListNewMessages: %v", err)
	}
	if len(messages) != 1 || len(messages[0].Attachments) != 0 || messages[0].BodyIsHTML {
		t.Fatalf("unexpected result: %+v", messages)
	}
}
