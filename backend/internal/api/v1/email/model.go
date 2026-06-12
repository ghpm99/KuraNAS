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
