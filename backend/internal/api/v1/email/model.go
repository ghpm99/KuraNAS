package email

import (
	"encoding/json"
	"time"
)

// Provider identifies a supported e-mail provider. It matches the CHECK
// constraint on the email_account table.
type Provider string

const (
	ProviderGoogle    Provider = "google"
	ProviderMicrosoft Provider = "microsoft"
)

func (p Provider) IsValid() bool {
	return p == ProviderGoogle || p == ProviderMicrosoft
}

// AccountStatus matches the CHECK constraint on the email_account table.
type AccountStatus string

const (
	StatusLinked         AccountStatus = "linked"
	StatusError          AccountStatus = "error"
	StatusReauthRequired AccountStatus = "reauth_required"
)

// AccountModel mirrors a row in the email_account table. TokenCiphertext is
// the AES-GCM-sealed TokenSet and never leaves the service layer.
type AccountModel struct {
	ID              int
	Provider        Provider
	Address         string
	DisplayName     string
	TokenCiphertext []byte
	Status          AccountStatus
	SyncEnabled     bool
	LastSyncAt      *time.Time
	LastError       string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// TokenSet is the plaintext shape sealed into token_ciphertext. It is never
// serialized into a DTO, log or error message.
type TokenSet struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
}

// MessageStatus matches the CHECK constraint on the email_message table.
type MessageStatus string

const (
	MsgStatusPending         MessageStatus = "pending"
	MsgStatusPrefilteredSpam MessageStatus = "prefiltered_spam"
	MsgStatusAnalyzed        MessageStatus = "analyzed"
	MsgStatusFailed          MessageStatus = "failed"
)

// AuthResults holds the SPF/DKIM/DMARC verdicts (stored as JSONB). It mirrors
// mailfetch.AuthResults but lives in the domain so the pre-filter never depends
// on the transport layer.
type AuthResults struct {
	SPF   string `json:"spf"`
	DKIM  string `json:"dkim"`
	DMARC string `json:"dmarc"`
}

// AttachmentMeta is metadata only — content is never stored (stored as JSONB).
type AttachmentMeta struct {
	Filename string `json:"filename"`
	Mime     string `json:"mime"`
	Size     int64  `json:"size"`
}

// MessageModel mirrors a row in the email_message table. SanitizedBody is plain
// text only; Attachments/LinkDomains are evidence, never fetched content.
type MessageModel struct {
	ID                int
	AccountID         int
	ProviderMessageID string
	SenderName        string
	SenderAddress     string
	Subject           string
	Snippet           string
	SanitizedBody     string
	ReceivedAt        time.Time
	AuthResults       AuthResults
	Attachments       []AttachmentMeta
	LinkDomains       []string
	PrefilterRules    []string
	Status            MessageStatus
	CreatedAt         time.Time

	// Analysis fields are populated only by the listing JOIN with email_analysis
	// (task 16); they stay zero-valued for a message not yet analyzed.
	Verdict    Verdict
	Importance Importance
	Summary    string
}

// Verdict is the AI classification of a message. It matches the CHECK
// constraint on the email_analysis table.
type Verdict string

const (
	VerdictLegitimate Verdict = "legitimate"
	VerdictSuspicious Verdict = "suspicious"
	VerdictMalicious  Verdict = "malicious"
)

func (v Verdict) IsValid() bool {
	return v == VerdictLegitimate || v == VerdictSuspicious || v == VerdictMalicious
}

// Importance is the priority hint the classifier assigns. It matches the CHECK
// constraint on the email_analysis table.
type Importance string

const (
	ImportanceLow    Importance = "low"
	ImportanceNormal Importance = "normal"
	ImportanceHigh   Importance = "high"
)

func (i Importance) IsValid() bool {
	return i == ImportanceLow || i == ImportanceNormal || i == ImportanceHigh
}

// AnalysisModel mirrors a row in the email_analysis table. Evidence is the
// structured signal list behind the verdict; Summary is set only for legitimate
// mail. The model answer is data: it is validated, persisted and shown — never
// fed back into another AI call.
type AnalysisModel struct {
	ID           int
	MessageID    int
	Verdict      Verdict
	RiskScore    int
	Evidence     []string
	Summary      string
	Importance   Importance
	ProviderUsed string
	ModelUsed    string
	AnalyzedAt   time.Time
}

func encodeTokenSet(tokens TokenSet) ([]byte, error) {
	return json.Marshal(tokens)
}

func decodeTokenSet(raw []byte) (TokenSet, error) {
	var tokens TokenSet
	if err := json.Unmarshal(raw, &tokens); err != nil {
		return TokenSet{}, err
	}
	return tokens, nil
}
