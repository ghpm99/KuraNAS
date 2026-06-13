// Package gmail fetches inbox message metadata and bodies from the Gmail API.
// It uses format=full to read the inline body and attachment metadata in a
// single call and NEVER calls users.messages.attachments.get — attachment
// content is never downloaded (hard rule of the e-mail feature).
package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nas-go/api/pkg/mailfetch"
)

// DefaultBaseURL is the production Gmail API host. Tests point a client at an
// httptest server instead.
const DefaultBaseURL = "https://gmail.googleapis.com"

type Client struct {
	baseURL string
	http    *mailfetch.GuardedClient
}

// NewClient builds a Gmail fetcher reaching exactly the host of baseURL.
func NewClient(baseURL string) *Client {
	host := ""
	if parsed, err := url.Parse(baseURL); err == nil {
		host = parsed.Hostname()
	}
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		http:    mailfetch.NewGuardedClient(host),
	}
}

// NewDefaultClient builds a Gmail fetcher pointed at the production API.
func NewDefaultClient() *Client { return NewClient(DefaultBaseURL) }

type listResponse struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

type messageResponse struct {
	ID           string         `json:"id"`
	InternalDate string         `json:"internalDate"`
	Payload      messagePayload `json:"payload"`
}

type messagePayload struct {
	MimeType string           `json:"mimeType"`
	Filename string           `json:"filename"`
	Headers  []header         `json:"headers"`
	Body     messageBody      `json:"body"`
	Parts    []messagePayload `json:"parts"`
}

type header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type messageBody struct {
	Size         int64  `json:"size"`
	Data         string `json:"data"`
	AttachmentID string `json:"attachmentId"`
}

// ListNewMessages lists inbox messages newer than `since` (capped at `max`) and
// reads each one's metadata and inline body.
func (c *Client) ListNewMessages(ctx context.Context, accessToken string, since time.Time, max int) ([]mailfetch.RawMessage, error) {
	if max <= 0 {
		return nil, nil
	}

	query := url.Values{}
	query.Set("maxResults", strconv.Itoa(max))
	q := "in:inbox"
	if !since.IsZero() {
		q += " after:" + strconv.FormatInt(since.Unix(), 10)
	}
	query.Set("q", q)

	var list listResponse
	listURL := fmt.Sprintf("%s/gmail/v1/users/me/messages?%s", c.baseURL, query.Encode())
	if err := c.http.GetJSON(ctx, listURL, accessToken, &list); err != nil {
		return nil, err
	}

	messages := make([]mailfetch.RawMessage, 0, len(list.Messages))
	for _, ref := range list.Messages {
		msg, err := c.getMessage(ctx, accessToken, ref.ID)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	return messages, nil
}

func (c *Client) getMessage(ctx context.Context, accessToken, id string) (mailfetch.RawMessage, error) {
	var raw messageResponse
	getURL := fmt.Sprintf("%s/gmail/v1/users/me/messages/%s?format=full", c.baseURL, url.PathEscape(id))
	if err := c.http.GetJSON(ctx, getURL, accessToken, &raw); err != nil {
		return mailfetch.RawMessage{}, err
	}
	return parseMessage(raw), nil
}

func parseMessage(raw messageResponse) mailfetch.RawMessage {
	headers := indexHeaders(raw.Payload.Headers)

	name, address := parseFrom(headers["from"])
	body, isHTML := extractBody(raw.Payload)
	attachments := collectAttachments(raw.Payload)

	return mailfetch.RawMessage{
		ProviderMessageID: raw.ID,
		SenderName:        name,
		SenderAddress:     address,
		Subject:           headers["subject"],
		ReceivedAt:        parseInternalDate(raw.InternalDate),
		AuthResults:       mailfetch.ParseAuthResults(headers["authentication-results"]),
		Attachments:       attachments,
		Body:              body,
		BodyIsHTML:        isHTML,
	}
}

// indexHeaders lower-cases header names so lookups are case-insensitive. When a
// header repeats, the first value wins.
func indexHeaders(headers []header) map[string]string {
	indexed := make(map[string]string, len(headers))
	for _, h := range headers {
		key := strings.ToLower(h.Name)
		if _, exists := indexed[key]; !exists {
			indexed[key] = h.Value
		}
	}
	return indexed
}

func parseFrom(value string) (name string, address string) {
	if value == "" {
		return "", ""
	}
	if parsed, err := mail.ParseAddress(value); err == nil {
		return parsed.Name, strings.ToLower(parsed.Address)
	}
	return "", strings.ToLower(strings.TrimSpace(value))
}

func parseInternalDate(raw string) time.Time {
	ms, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || ms <= 0 {
		return time.Time{}
	}
	return time.UnixMilli(ms).UTC()
}

// extractBody walks the MIME tree preferring text/plain, falling back to
// text/html. Attachment parts (those carrying an attachmentId) are skipped —
// their bytes are never requested.
func extractBody(payload messagePayload) (body string, isHTML bool) {
	plain, html := walkBody(payload)
	if plain != "" {
		return plain, false
	}
	return html, html != ""
}

func walkBody(payload messagePayload) (plain string, html string) {
	mimeType := strings.ToLower(payload.MimeType)
	isAttachment := payload.Body.AttachmentID != "" || payload.Filename != ""

	if !isAttachment && payload.Body.Data != "" {
		switch {
		case strings.HasPrefix(mimeType, "text/plain"):
			plain = decodeBody(payload.Body.Data)
		case strings.HasPrefix(mimeType, "text/html"):
			html = decodeBody(payload.Body.Data)
		}
	}

	for _, part := range payload.Parts {
		childPlain, childHTML := walkBody(part)
		if plain == "" {
			plain = childPlain
		}
		if html == "" {
			html = childHTML
		}
	}
	return plain, html
}

// collectAttachments gathers metadata for every part that declares a filename.
// Only name/MIME/size are kept — the attachment body is never fetched.
func collectAttachments(payload messagePayload) []mailfetch.AttachmentMeta {
	var attachments []mailfetch.AttachmentMeta
	var walk func(messagePayload)
	walk = func(p messagePayload) {
		if p.Filename != "" {
			attachments = append(attachments, mailfetch.AttachmentMeta{
				Filename: p.Filename,
				Mime:     p.MimeType,
				Size:     p.Body.Size,
			})
		}
		for _, child := range p.Parts {
			walk(child)
		}
	}
	walk(payload)
	return attachments
}

// decodeBody decodes Gmail's web-safe base64 body, tolerating missing padding.
func decodeBody(data string) string {
	if decoded, err := base64.URLEncoding.DecodeString(data); err == nil {
		return string(decoded)
	}
	if decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(data, "=")); err == nil {
		return string(decoded)
	}
	return ""
}
