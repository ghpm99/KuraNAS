// Package mailfetch is the transport layer that pulls message metadata and
// bodies from external mail providers (Gmail, Microsoft Graph). It is
// deliberately dumb: it returns raw provider data and never sanitizes,
// classifies, follows a URL or downloads an attachment. Those decisions belong
// to the email domain. Every client talks to exactly one fixed, allowlisted
// host (hard rule of the e-mail feature).
package mailfetch

import (
	"context"
	"errors"
	"time"
)

// ErrHostNotAllowed is returned before any byte leaves the process when a URL
// resolves to a host outside the client's fixed allowlist.
var ErrHostNotAllowed = errors.New("mailfetch: host not in the allowlist")

// AuthResults holds the SPF/DKIM/DMARC verdicts parsed from the
// Authentication-Results header. Empty strings mean "absent / not evaluated".
type AuthResults struct {
	SPF   string `json:"spf"`
	DKIM  string `json:"dkim"`
	DMARC string `json:"dmarc"`
}

// AttachmentMeta is metadata ONLY — name, declared MIME type and byte size.
// Attachment content is never fetched, stored or executed.
type AttachmentMeta struct {
	Filename string `json:"filename"`
	Mime     string `json:"mime"`
	Size     int64  `json:"size"`
}

// RawMessage is the provider-neutral shape a fetcher returns. Body is the raw
// (possibly HTML) body; the email domain sanitizes it. The fetcher never
// interprets, follows or renders anything in it.
type RawMessage struct {
	ProviderMessageID string
	SenderName        string
	SenderAddress     string
	Subject           string
	ReceivedAt        time.Time
	AuthResults       AuthResults
	Attachments       []AttachmentMeta
	Body              string
	BodyIsHTML        bool
}

// Fetcher lists inbox messages newer than `since` (zero time = no lower bound)
// for one account, capped at `max` messages. Implementations talk to a single
// allowlisted provider host and MUST NOT request attachment content.
type Fetcher interface {
	ListNewMessages(ctx context.Context, accessToken string, since time.Time, max int) ([]RawMessage, error)
}
