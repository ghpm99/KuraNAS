// Package graph fetches inbox message metadata and bodies from Microsoft Graph.
// It selects only metadata + body fields and reads attachment metadata via
// $select=name,contentType,size, so attachment content (contentBytes) is NEVER
// returned or downloaded (hard rule of the e-mail feature).
package graph

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"nas-go/api/pkg/mailfetch"
)

// DefaultBaseURL is the production Microsoft Graph host. Tests point a client at
// an httptest server instead.
const DefaultBaseURL = "https://graph.microsoft.com"

type Client struct {
	baseURL string
	http    *mailfetch.GuardedClient
}

// NewClient builds a Graph fetcher reaching exactly the host of baseURL.
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

// NewDefaultClient builds a Graph fetcher pointed at the production API.
func NewDefaultClient() *Client { return NewClient(DefaultBaseURL) }

type listResponse struct {
	Value []messageResource `json:"value"`
}

type messageResource struct {
	ID               string `json:"id"`
	Subject          string `json:"subject"`
	ReceivedDateTime string `json:"receivedDateTime"`
	HasAttachments   bool   `json:"hasAttachments"`
	From             struct {
		EmailAddress struct {
			Name    string `json:"name"`
			Address string `json:"address"`
		} `json:"emailAddress"`
	} `json:"from"`
	Body struct {
		ContentType string `json:"contentType"`
		Content     string `json:"content"`
	} `json:"body"`
	InternetMessageHeaders []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"internetMessageHeaders"`
}

type attachmentsResponse struct {
	Value []struct {
		Name        string `json:"name"`
		ContentType string `json:"contentType"`
		Size        int64  `json:"size"`
	} `json:"value"`
}

// ListNewMessages lists inbox messages newer than `since` (capped at `max`) and
// reads each one's metadata, body and — when present — attachment metadata.
func (c *Client) ListNewMessages(ctx context.Context, accessToken string, since time.Time, max int) ([]mailfetch.RawMessage, error) {
	if max <= 0 {
		return nil, nil
	}

	query := url.Values{}
	query.Set("$select", "id,subject,from,receivedDateTime,internetMessageHeaders,body,hasAttachments")
	query.Set("$top", strconv.Itoa(max))
	query.Set("$orderby", "receivedDateTime desc")
	if !since.IsZero() {
		query.Set("$filter", "receivedDateTime ge "+since.UTC().Format(time.RFC3339))
	}

	var list listResponse
	listURL := fmt.Sprintf("%s/v1.0/me/messages?%s", c.baseURL, query.Encode())
	if err := c.http.GetJSON(ctx, listURL, accessToken, &list); err != nil {
		return nil, err
	}

	messages := make([]mailfetch.RawMessage, 0, len(list.Value))
	for _, resource := range list.Value {
		message := parseMessage(resource)
		if resource.HasAttachments {
			attachments, err := c.getAttachmentMeta(ctx, accessToken, resource.ID)
			if err != nil {
				return nil, err
			}
			message.Attachments = attachments
		}
		messages = append(messages, message)
	}
	return messages, nil
}

func (c *Client) getAttachmentMeta(ctx context.Context, accessToken, messageID string) ([]mailfetch.AttachmentMeta, error) {
	query := url.Values{}
	// Selecting only metadata fields keeps contentBytes out of the response —
	// the attachment payload is never transferred.
	query.Set("$select", "name,contentType,size")

	var raw attachmentsResponse
	attachmentsURL := fmt.Sprintf("%s/v1.0/me/messages/%s/attachments?%s", c.baseURL, url.PathEscape(messageID), query.Encode())
	if err := c.http.GetJSON(ctx, attachmentsURL, accessToken, &raw); err != nil {
		return nil, err
	}

	attachments := make([]mailfetch.AttachmentMeta, 0, len(raw.Value))
	for _, item := range raw.Value {
		attachments = append(attachments, mailfetch.AttachmentMeta{
			Filename: item.Name,
			Mime:     item.ContentType,
			Size:     item.Size,
		})
	}
	return attachments, nil
}

func parseMessage(resource messageResource) mailfetch.RawMessage {
	var receivedAt time.Time
	if parsed, err := time.Parse(time.RFC3339, resource.ReceivedDateTime); err == nil {
		receivedAt = parsed.UTC()
	}

	authHeaders := make([]string, 0, 2)
	for _, h := range resource.InternetMessageHeaders {
		if strings.EqualFold(h.Name, "Authentication-Results") {
			authHeaders = append(authHeaders, h.Value)
		}
	}

	return mailfetch.RawMessage{
		ProviderMessageID: resource.ID,
		SenderName:        resource.From.EmailAddress.Name,
		SenderAddress:     strings.ToLower(resource.From.EmailAddress.Address),
		Subject:           resource.Subject,
		ReceivedAt:        receivedAt,
		AuthResults:       mailfetch.ParseAuthResults(authHeaders...),
		Body:              resource.Body.Content,
		BodyIsHTML:        strings.EqualFold(resource.Body.ContentType, "html"),
	}
}
