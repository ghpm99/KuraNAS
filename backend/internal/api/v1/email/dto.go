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
