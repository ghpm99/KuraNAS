package email

import (
	"nas-go/api/pkg/database"
)

type RepositoryInterface interface {
	GetDbContext() *database.DbContext
	ListAccounts() ([]AccountModel, error)
	GetAccountByID(id int) (AccountModel, error)
	UpsertAccount(model AccountModel) (int, error)
	UpdateAccountTokens(id int, tokenCiphertext []byte, status AccountStatus, lastError string) error
	UpdateSyncEnabled(id int, enabled bool) error
	DeleteAccount(id int) error
}

type ServiceInterface interface {
	ListAccounts() ([]AccountDto, error)
	DeleteAccount(id int) error
	SetSyncEnabled(id int, enabled bool) error
	GoogleAuthURL() (GoogleAuthURLDto, error)
	HandleGoogleCallback(state string, code string) error
	StartMicrosoftDeviceCode() (DeviceCodeDto, error)
	MicrosoftDeviceCodeStatus() DeviceCodeStatusDto
	// ValidAccessToken returns a usable access token for the account,
	// refreshing (and re-sealing) it when expired. This is the seam the sync
	// worker (task 15) will consume.
	ValidAccessToken(accountID int) (string, error)
}
