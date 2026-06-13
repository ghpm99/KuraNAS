package email

import "time"

// AccountDto is the API representation of a linked account. It deliberately
// has no token field of any kind.
type AccountDto struct {
	ID          int        `json:"id"`
	Provider    string     `json:"provider"`
	Address     string     `json:"address"`
	DisplayName string     `json:"display_name"`
	Status      string     `json:"status"`
	SyncEnabled bool       `json:"sync_enabled"`
	LastSyncAt  *time.Time `json:"last_sync_at"`
	LastError   string     `json:"last_error"`
	CreatedAt   time.Time  `json:"created_at"`
}

// UpdateSyncEnabledDto is the body of PUT /email/accounts/:id/sync-enabled.
type UpdateSyncEnabledDto struct {
	SyncEnabled bool `json:"sync_enabled"`
}

// GoogleAuthURLDto is the response of POST /email/accounts/google/auth-url.
type GoogleAuthURLDto struct {
	AuthURL string `json:"auth_url"`
}

// DeviceCodeDto is the response of POST /email/accounts/microsoft/device-code:
// what the user needs to finish the login on another device.
type DeviceCodeDto struct {
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Message         string `json:"message"`
}

// DeviceCodeStatusDto reports the progress of a Microsoft device-code link.
type DeviceCodeStatusDto struct {
	Status string `json:"status"`
}

// Device-code link progress values (in-memory state, not the DB enum).
const (
	DeviceCodeIdle    = "idle"
	DeviceCodePending = "pending"
	DeviceCodeLinked  = "linked"
	DeviceCodeExpired = "expired"
	DeviceCodeError   = "error"
)

// MessageDto is the lean API representation of a synced message. It carries NO
// body — the listing is built for a low-powered kiosk, so the payload stays
// small. The verdict/importance/short-summary fields are present only once the
// message has been analyzed (task 16); they are omitted otherwise.
type MessageDto struct {
	ID            int       `json:"id"`
	AccountID     int       `json:"account_id"`
	SenderName    string    `json:"sender_name"`
	SenderAddress string    `json:"sender_address"`
	Subject       string    `json:"subject"`
	Snippet       string    `json:"snippet"`
	ReceivedAt    time.Time `json:"received_at"`
	Status        string    `json:"status"`
	Verdict       string    `json:"verdict,omitempty"`
	Importance    string    `json:"importance,omitempty"`
	Summary       string    `json:"summary,omitempty"`
}

func (m MessageModel) toDto() MessageDto {
	return MessageDto{
		ID:            m.ID,
		AccountID:     m.AccountID,
		SenderName:    m.SenderName,
		SenderAddress: m.SenderAddress,
		Subject:       m.Subject,
		Snippet:       m.Snippet,
		ReceivedAt:    m.ReceivedAt,
		Status:        string(m.Status),
		Verdict:       string(m.Verdict),
		Importance:    string(m.Importance),
		Summary:       m.Summary,
	}
}

// AnalysisDto is the API representation of one message's AI verdict, returned by
// GET /email/messages/:id/summary.
type AnalysisDto struct {
	MessageID    int      `json:"message_id"`
	Verdict      string   `json:"verdict"`
	RiskScore    int      `json:"risk_score"`
	Evidence     []string `json:"evidence"`
	Summary      string   `json:"summary"`
	Importance   string   `json:"importance"`
	ProviderUsed string   `json:"provider_used"`
	ModelUsed    string   `json:"model_used"`
}

func (m AnalysisModel) toDto() AnalysisDto {
	evidence := m.Evidence
	if evidence == nil {
		evidence = []string{}
	}
	return AnalysisDto{
		MessageID:    m.MessageID,
		Verdict:      string(m.Verdict),
		RiskScore:    m.RiskScore,
		Evidence:     evidence,
		Summary:      m.Summary,
		Importance:   string(m.Importance),
		ProviderUsed: m.ProviderUsed,
		ModelUsed:    m.ModelUsed,
	}
}

// ProviderPreferenceDto is the body/response of GET|PUT
// /email/settings/provider: which AI provider analyzes e-mail.
type ProviderPreferenceDto struct {
	Provider string `json:"provider"`
}

func (m AccountModel) toDto() AccountDto {
	return AccountDto{
		ID:          m.ID,
		Provider:    string(m.Provider),
		Address:     m.Address,
		DisplayName: m.DisplayName,
		Status:      string(m.Status),
		SyncEnabled: m.SyncEnabled,
		LastSyncAt:  m.LastSyncAt,
		LastError:   m.LastError,
		CreatedAt:   m.CreatedAt,
	}
}
